package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"prune/internal/cli"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := cli.Execute(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		//nolint: gocritic
		os.Exit(1)
	}
}
