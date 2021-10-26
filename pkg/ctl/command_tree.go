package ctl

type CommandTree interface {
	Command
	Subcommands() []Command
}
