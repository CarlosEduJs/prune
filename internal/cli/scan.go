package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"prune/internal/config"
	"prune/internal/lang/js"
	"prune/internal/report"
	"prune/internal/rules"
)

func runScan(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	opts := rootOptions{}
	parseRootFlags(fs, &opts)
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(opts.configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if len(opts.paths) > 0 {
		cfg.Scan.Paths = opts.paths
	}

	findings, err := runLanguage(ctx, cfg)
	if err != nil {
		return err
	}

	out, err := report.NewFormatter(opts.format)
	if err != nil {
		return fmt.Errorf("creating formatter: %w", err)
	}

	filtered := report.FilterByConfidence(findings, opts.minConfidence)
	data, err := out.Format(filtered)
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(data)
	return err
}

func runLanguage(ctx context.Context, cfg *config.Config) ([]rules.Finding, error) {
	if cfg.Project.Language == "" {
		return nil, errors.New("project.language is required")
	}

	switch cfg.Project.Language {
	case "js-ts":
		return js.Analyze(ctx, cfg)
	default:
		return nil, errors.New("unsupported language")
	}
}
