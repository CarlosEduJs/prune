package cmd

import (
	"context"
)

type CommandRunner func(ctx context.Context, args []string) error

type Command struct {
	Name  string
	Run   CommandRunner
	Usage string
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
	result := make(map[string]*Command, len(registry))
	for k, v := range registry {
		result[k] = v
	}
	return result
}
