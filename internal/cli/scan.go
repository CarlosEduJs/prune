package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"prune/internal/config"
	"prune/internal/lang/js"
	"prune/internal/report"
	"prune/internal/rules"
)

func runScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	opts := rootOptions{}
	parseRootFlags(fs, &opts)
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(opts.configPath)
	if err != nil {
		return err
	}

	if len(opts.paths) > 0 {
		cfg.Scan.Paths = opts.paths
	}

	findings, err := runLanguage(cfg)
	if err != nil {
		return err
	}

	out := report.NewFormatter(opts.format)
	if out == nil {
		return fmt.Errorf("unknown format: %s", opts.format)
	}

	filtered := report.FilterByConfidence(findings, opts.minConfidence)
	data, err := out.Format(filtered)
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(data)
	return err
}

func runLanguage(cfg *config.Config) ([]rules.Finding, error) {
	if cfg.Project.Language == "" {
		return nil, errors.New("project.language is required")
	}

	switch cfg.Project.Language {
	case "js-ts":
		return js.Analyze(cfg)
	default:
		return nil, errors.New("unsupported language")
	}
}
