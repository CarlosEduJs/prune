package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"prune/internal/config"
	"prune/internal/lang/js"
	"prune/internal/report"
	"prune/internal/rules"
	"time"
)

func runScan(ctx context.Context, args []string) error {
	start := time.Now()
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
	} else {
		configDir, err := filepath.Abs(filepath.Dir(opts.configPath))
		if err == nil {
			for i, p := range cfg.Scan.Paths {
				if !filepath.IsAbs(p) {
					cfg.Scan.Paths[i] = filepath.Join(configDir, p)
				}
			}
		}
	}

	if opts.stream {
		cfg.Scan.Stream.Enabled = true
	}
	if opts.streamInterval > 0 {
		cfg.Scan.Stream.IntervalMs = opts.streamInterval
	}

	var findings []rules.Finding
	streamedOutput := false

	if cfg.Scan.Stream.Enabled {
		var streamHandler js.StreamHandler
		format := opts.format

		if format == "json" && cfg.Scan.Stream.Enabled {
			format = "ndjson"
		}

		if format == "ndjson" {
			streamHandler = func(batchFindings []rules.Finding) error {
				filtered := report.FilterByConfidence(batchFindings, opts.minConfidence)
				out, err := report.NewFormatter("ndjson")
				if err != nil {
					return err
				}
				data, err := out.Format(filtered)
				if err != nil {
					return err
				}
				os.Stdout.Write(data)
				return nil
			}
			streamedOutput = true
		}

		findings, err = js.AnalyzeStreaming(ctx, cfg, streamHandler)
	} else {
		findings, err = runLanguage(ctx, cfg)
	}
	if err != nil {
		return err
	}

	if streamedOutput {
		return nil
	}

	out, err := report.NewFormatter(opts.format, report.FormatterOptions{
		Duration:  time.Since(start),
		Compact:   opts.compact,
		Only:      opts.only,
		Deletable: opts.deletable,
	})
	if err != nil {
		return fmt.Errorf("creating formatter: %w", err)
	}

	filtered := report.FilterByConfidence(findings, opts.minConfidence)
	data, err := out.Format(filtered)
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(data)
	if err != nil {
		return err
	}

	if opts.failOnFindings && len(filtered) > 0 {
		return fmt.Errorf("found %d findings", len(filtered))
	}

	return nil
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
