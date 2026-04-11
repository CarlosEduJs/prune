package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
)

const version = "0.0.1"

type rootOptions struct {
	configPath     string
	format         string
	minConfidence  string
	paths          stringSlice
	failOnFindings bool
	stream         bool
	streamInterval int
}

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", []string(*s))
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return printUsage(os.Stderr)
	}

	switch args[0] {
	case "version":
		fmt.Println("prune version ", version)
		return nil
	case "init":
		return runInit(ctx, args[1:])
	case "scan":
		return runScan(ctx, args[1:])
	case "rules":
		return runRules(ctx, args[1:])
	case "help", "-h", "--help":
		return printUsage(os.Stdout)
	default:
		return fmt.Errorf("unknown command: %q", args[0])
	}
}

func printUsage(w io.Writer) error {
	_, err := fmt.Fprint(w, `prune - dead code analysis CLI

Usage:
  prune <command> [flags]

Commands:
  version Print version
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
	fs.BoolVar(&opts.failOnFindings, "fail-on-findings", false, "Exit with error if findings are found")
	fs.BoolVar(&opts.stream, "stream", false, "Enable streaming mode with partial results")
	fs.IntVar(&opts.streamInterval, "stream-interval", 250, "Interval in ms between stream flushes")
}
