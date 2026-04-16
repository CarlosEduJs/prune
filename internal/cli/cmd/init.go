package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"prune/internal/config"
)

// NewInitCommand returns a command that creates a default prune.yaml config file.
func NewInitCommand() *Command {
	return &Command{
		Name:  "init",
		Usage: "Create a default prune.yaml",
		Run:   runInit,
	}
}

// runInit executes the init command, writing a default config to the specified path.
func runInit(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	outPath := fs.String("out", "prune.yaml", "Output config path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if err := config.WriteDefault(*outPath); err != nil {
		return fmt.Errorf("writing default config: %w", err)
	}

	_, err := fmt.Fprintf(os.Stdout, "✅ Created %s\n", *outPath)
	return err
}
