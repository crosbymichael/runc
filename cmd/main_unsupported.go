// +build !linux,!solaris

package cmd

import "github.com/urfave/cli"

var (
	CheckpointCommand cli.Command
	EventsCommand     cli.Command
	RestoreCommand    cli.Command
	SpecCommand       cli.Command
	KillCommand       cli.Command
)
