package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"prune/internal/cli/cmd"
	"prune/internal/version"
)

func Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return printUsage(os.Stderr)
	}

	switch args[0] {
	case "version":
		fmt.Println("prune version", version.Version)
		return nil
	case "help", "-h", "--help":
		return printUsage(os.Stdout)
	default:
		c := cmd.Get(args[0])
		if c == nil {
			return fmt.Errorf("unknown command: %q", args[0])
		}
		return c.Run(ctx, args[1:])
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
