package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"prune/internal/config"
)

func NewInitCommand() *Command {
	return &Command{
		Name:  "init",
		Usage: "Create a default prune.yaml",
		Run:   runInit,
	}
}

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
