package cmd

import (
	"context"
	"flag"
)

type CommandRunner func(ctx context.Context, args []string) error

type Command struct {
	Name    string
	Run     CommandRunner
	FlagSet *flag.FlagSet
	Usage   string
}

var registry = map[string]*Command{}

func init() {
	Register(NewScanCommand())
	Register(NewInitCommand())
	Register(NewRulesCommand())
}

func Register(cmd *Command) {
	registry[cmd.Name] = cmd
}

func Get(name string) *Command {
	return registry[name]
}

func All() map[string]*Command {
	return registry
}
