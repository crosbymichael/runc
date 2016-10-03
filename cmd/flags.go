package cmd

import "github.com/urfave/cli"

var GlobalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "debug",
		Usage: "enable debug output for logging",
	},
	cli.StringFlag{
		Name:  "log",
		Value: "/dev/null",
		Usage: "set the log file path where internal debug information is written",
	},
	cli.StringFlag{
		Name:  "log-format",
		Value: "text",
		Usage: "set the format used by logs ('text' (default), or 'json')",
	},
	cli.StringFlag{
		Name:  "root",
		Value: "/run/runc",
		Usage: "root directory for storage of container state (this should be located in tmpfs)",
	},
	cli.StringFlag{
		Name:  "criu",
		Value: "criu",
		Usage: "path to the criu binary used for checkpoint and restore",
	},
	cli.BoolFlag{
		Name:  "systemd-cgroup",
		Usage: "enable systemd cgroup support, expects cgroupsPath to be of form \"slice:prefix:name\" for e.g. \"system.slice:runc:434234\"",
	},
}
