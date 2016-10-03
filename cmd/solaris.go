// +build solaris

package cmd

import "github.com/urfave/cli"

var (
	CheckpointCommand cli.Command
	EventsCommand     cli.Command
	RestoreCommand    cli.Command
	SpecCommand       cli.Command
	KillCommand       cli.Command
	DeleteCommand     cli.Command
	ExecCommand       cli.Command
	InitCommand       cli.Command
	ListCommand       cli.Command
	PauseCommand      cli.Command
	ResumeCommand     cli.Command
	StartCommand      cli.Command
	StateCommand      cli.Command
)
