package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"prune/internal/config"
	"prune/internal/lang/js"
	"prune/internal/report"
	"prune/internal/rules"
)

type scanFlags struct {
	configPath     string
	format         string
	minConfidence  string
	paths          stringSlice
	failOnFindings bool
	stream         bool
	streamInterval int
	compact        bool
	only           string
	deletable      bool
}

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", []string(*s))
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func parseScanFlags(fs *flag.FlagSet, opts *scanFlags) {
	fs.StringVar(&opts.configPath, "config", "prune.yaml", "Path to prune config")
	fs.StringVar(&opts.format, "format", "pretty", "Output format: pretty, json, or ndjson")
	fs.StringVar(&opts.minConfidence, "min-confidence", "safe", "Minimum confidence to report")
	fs.Var(&opts.paths, "paths", "Paths to scan (repeatable)")
	fs.BoolVar(&opts.failOnFindings, "fail-on-findings", false, "Exit with error if findings are found")
	fs.BoolVar(&opts.stream, "stream", false, "Enable streaming mode with partial results")
	fs.IntVar(&opts.streamInterval, "stream-interval", 250, "Interval in ms between stream flushes")
	fs.BoolVar(&opts.compact, "compact", false, "Show only summary counts")
	fs.StringVar(&opts.only, "only", "", "Show only findings with this confidence (safe, review, likely_dead)")
	fs.BoolVar(&opts.deletable, "deletable", false, "Show only files that are safe to delete")
}

func NewScanCommand() *Command {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	opts := scanFlags{}
	parseScanFlags(fs, &opts)

	return &Command{
		Name:    "scan",
		FlagSet: fs,
		Usage:   "Analyze project and report findings",
		Run:     runScan,
	}
}

func runScan(ctx context.Context, args []string) error {
	start := time.Now()
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	opts := scanFlags{}
	parseScanFlags(fs, &opts)

	if len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			return err
		}
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
		findings, err = analyzeLanguage(ctx, cfg)
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
		Config:    cfg,
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

func analyzeLanguage(ctx context.Context, cfg *config.Config) ([]rules.Finding, error) {
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
