package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

type rootOptions struct {
	configPath    string
	format        string
	minConfidence string
	paths         stringSlice
}

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", []string(*s))
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func Execute(args []string) error {
	if len(args) == 0 {
		return printUsage()
	}

	switch args[0] {
	case "init":
		return runInit(args[1:])
	case "scan":
		return runScan(args[1:])
	case "rules":
		return runRules(args[1:])
	case "help", "-h", "--help":
		return printUsage()
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func printUsage() error {
	_, err := fmt.Fprint(os.Stdout, `prune - dead code analysis CLI

Usage:
  prune <command> [flags]

Commands:
  init    Create a default prune.yaml
  scan    Analyze project and report findings
  rules   List available rules
`)
	return err
}

func parseRootFlags(fs *flag.FlagSet, opts *rootOptions) {
	fs.StringVar(&opts.configPath, "config", "prune.yaml", "Path to prune config")
	fs.StringVar(&opts.format, "format", "table", "Output format: table or json")
	fs.StringVar(&opts.minConfidence, "min-confidence", "safe", "Minimum confidence to report")
	fs.Var(&opts.paths, "paths", "Paths to scan (repeatable)")
}

func requireCommand(args []string) error {
	if len(args) == 0 {
		return errors.New("missing subcommand")
	}
	return nil
}
