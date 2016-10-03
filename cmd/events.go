// +build linux

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/urfave/cli"
)

// Event struct for encoding the Event data to json.
type Event struct {
	Type string      `json:"type"`
	ID   string      `json:"id"`
	Data interface{} `json:"data,omitempty"`
}

var EventsCommand = cli.Command{
	Name:  "events",
	Usage: "display container events such as OOM notifications, cpu, memory, and IO usage statistics",
	ArgsUsage: `<container-id>

Where "<container-id>" is the name for the instance of the container.`,
	Description: `The events command displays information about the container. By default the
information is displayed once every 5 seconds.`,
	Flags: []cli.Flag{
		cli.DurationFlag{Name: "interval", Value: 5 * time.Second, Usage: "set the stats collection interval"},
		cli.BoolFlag{Name: "stats", Usage: "display the container's stats then exit"},
	},
	Action: func(context *cli.Context) error {
		container, err := getContainer(context)
		if err != nil {
			return err
		}
		duration := context.Duration("interval")
		if duration <= 0 {
			return fmt.Errorf("duration interval must be greater than 0")
		}
		status, err := container.Status()
		if err != nil {
			return err
		}
		if status == libcontainer.Stopped {
			return fmt.Errorf("container with id %s is not running", container.ID())
		}
		var (
			stats  = make(chan *libcontainer.Stats, 1)
			events = make(chan *Event, 1024)
			group  = &sync.WaitGroup{}
		)
		group.Add(1)
		go func() {
			defer group.Done()
			enc := json.NewEncoder(os.Stdout)
			for e := range events {
				if err := enc.Encode(e); err != nil {
					logrus.Error(err)
				}
			}
		}()
		if context.Bool("stats") {
			s, err := container.Stats()
			if err != nil {
				return err
			}
			events <- &Event{Type: "stats", ID: container.ID(), Data: convertLibcontainerStats(s)}
			close(events)
			group.Wait()
			return nil
		}
		go func() {
			for range time.Tick(context.Duration("interval")) {
				s, err := container.Stats()
				if err != nil {
					logrus.Error(err)
					continue
				}
				stats <- s
			}
		}()
		n, err := container.NotifyOOM()
		if err != nil {
			return err
		}
		for {
			select {
			case _, ok := <-n:
				if ok {
					// this means an oom Event was received, if it is !ok then
					// the channel was closed because the container stopped and
					// the cgroups no longer exist.
					events <- &Event{Type: "oom", ID: container.ID()}
				} else {
					n = nil
				}
			case s := <-stats:
				events <- &Event{Type: "stats", ID: container.ID(), Data: convertLibcontainerStats(s)}
			}
			if n == nil {
				close(events)
				break
			}
		}
		group.Wait()
		return nil
	},
}

func convertLibcontainerStats(ls *libcontainer.Stats) *Stats {
	cg := ls.CgroupStats
	if cg == nil {
		return nil
	}
	var s Stats
	s.Pids.Current = cg.PidsStats.Current
	s.Pids.Limit = cg.PidsStats.Limit

	s.Cpu.Usage.Kernel = cg.CpuStats.CpuUsage.UsageInKernelmode
	s.Cpu.Usage.User = cg.CpuStats.CpuUsage.UsageInUsermode
	s.Cpu.Usage.Total = cg.CpuStats.CpuUsage.TotalUsage
	s.Cpu.Usage.Percpu = cg.CpuStats.CpuUsage.PercpuUsage
	s.Cpu.Throttling.Periods = cg.CpuStats.ThrottlingData.Periods
	s.Cpu.Throttling.ThrottledPeriods = cg.CpuStats.ThrottlingData.ThrottledPeriods
	s.Cpu.Throttling.ThrottledTime = cg.CpuStats.ThrottlingData.ThrottledTime

	s.Memory.Cache = cg.MemoryStats.Cache
	s.Memory.Kernel = convertMemoryEntry(cg.MemoryStats.KernelUsage)
	s.Memory.KernelTCP = convertMemoryEntry(cg.MemoryStats.KernelTCPUsage)
	s.Memory.Swap = convertMemoryEntry(cg.MemoryStats.SwapUsage)
	s.Memory.Usage = convertMemoryEntry(cg.MemoryStats.Usage)
	s.Memory.Raw = cg.MemoryStats.Stats

	s.Blkio.IoServiceBytesRecursive = convertBlkioEntry(cg.BlkioStats.IoServiceBytesRecursive)
	s.Blkio.IoServicedRecursive = convertBlkioEntry(cg.BlkioStats.IoServicedRecursive)
	s.Blkio.IoQueuedRecursive = convertBlkioEntry(cg.BlkioStats.IoQueuedRecursive)
	s.Blkio.IoServiceTimeRecursive = convertBlkioEntry(cg.BlkioStats.IoServiceTimeRecursive)
	s.Blkio.IoWaitTimeRecursive = convertBlkioEntry(cg.BlkioStats.IoWaitTimeRecursive)
	s.Blkio.IoMergedRecursive = convertBlkioEntry(cg.BlkioStats.IoMergedRecursive)
	s.Blkio.IoTimeRecursive = convertBlkioEntry(cg.BlkioStats.IoTimeRecursive)
	s.Blkio.SectorsRecursive = convertBlkioEntry(cg.BlkioStats.SectorsRecursive)

	s.Hugetlb = make(map[string]Hugetlb)
	for k, v := range cg.HugetlbStats {
		s.Hugetlb[k] = convertHugtlb(v)
	}
	return &s
}

func convertHugtlb(c cgroups.HugetlbStats) Hugetlb {
	return Hugetlb{
		Usage:   c.Usage,
		Max:     c.MaxUsage,
		Failcnt: c.Failcnt,
	}
}

func convertMemoryEntry(c cgroups.MemoryData) MemoryEntry {
	return MemoryEntry{
		Limit:   c.Limit,
		Usage:   c.Usage,
		Max:     c.MaxUsage,
		Failcnt: c.Failcnt,
	}
}

func convertBlkioEntry(c []cgroups.BlkioStatEntry) []BlkioEntry {
	var out []BlkioEntry
	for _, e := range c {
		out = append(out, BlkioEntry{
			Major: e.Major,
			Minor: e.Minor,
			Op:    e.Op,
			Value: e.Value,
		})
	}
	return out
}
