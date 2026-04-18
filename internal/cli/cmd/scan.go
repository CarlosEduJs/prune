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
	configPath        string
	format            string
	minConfidence     string
	paths             stringSlice
	failOnFindings    bool
	stream            bool
	streamInterval    int
	streamSet         bool
	streamIntervalSet bool
	streamBatchSize  int
	streamBatchSet   bool
	compact           bool
	only              string
	deletable         bool
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
	fs.IntVar(&opts.streamBatchSize, "stream-batch-size", 50, "Number of files to process per batch")
	fs.BoolVar(&opts.compact, "compact", false, "Show only summary counts")
	fs.StringVar(&opts.only, "only", "", "Show only findings with this confidence (safe, review, likely_dead)")
	fs.BoolVar(&opts.deletable, "deletable", false, "Show only files that are safe to delete")
}

func NewScanCommand() *Command {
	return &Command{
		Name:  "scan",
		Usage: "Analyze project and report findings",
		Run:   runScan,
	}
}

func runScan(ctx context.Context, args []string) error {
	start := time.Now()

	cfg, opts, err := parseFlagsAndConfig(args)
	if err != nil {
		return err
	}

	applyPathOverrides(cfg, opts.paths, opts.configPath)

	if opts.streamSet {
		cfg.Scan.Stream.Enabled = opts.stream
	}
	if opts.streamIntervalSet && opts.streamInterval > 0 {
		cfg.Scan.Stream.IntervalMs = opts.streamInterval
	}
	if opts.streamBatchSet && opts.streamBatchSize > 0 {
		cfg.Scan.Stream.BatchSize = opts.streamBatchSize
	}

	findings, streamedOutput, err := runAnalysis(ctx, cfg, opts)
	if err != nil {
		return err
	}

	if streamedOutput {
		return nil
	}

	return outputResults(findings, opts, cfg, start)
}

func parseFlagsAndConfig(args []string) (*config.Config, scanFlags, error) {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	opts := scanFlags{}
	parseScanFlags(fs, &opts)

	if len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			return nil, opts, err
		}
		fs.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "stream-interval":
				opts.streamIntervalSet = true
			case "stream":
				opts.streamSet = true
			case "stream-batch-size":
				opts.streamBatchSet = true
			}
		})
	}

	cfg, err := config.Load(opts.configPath)
	if err != nil {
		return nil, opts, fmt.Errorf("loading config: %w", err)
	}

	return cfg, opts, nil
}

func applyPathOverrides(cfg *config.Config, paths []string, configPath string) {
	if len(paths) > 0 {
		cfg.Scan.Paths = paths
		return
	}

	configDir, err := filepath.Abs(filepath.Dir(configPath))
	if err != nil {
		return
	}
	for i, p := range cfg.Scan.Paths {
		if !filepath.IsAbs(p) {
			cfg.Scan.Paths[i] = filepath.Join(configDir, p)
		}
	}
}

func runAnalysis(ctx context.Context, cfg *config.Config, opts scanFlags) ([]rules.Finding, bool, error) {
	if !cfg.Scan.Stream.Enabled {
		findings, err := analyzeLanguage(ctx, cfg)
		return findings, false, err
	}

	format := opts.format
	if format == "json" {
		format = "ndjson"
	}

	var streamHandler js.StreamHandler
	streamedOutput := false

	if format == "ndjson" {
		formatter, err := report.NewFormatter("ndjson")
		if err != nil {
			return nil, false, err
		}
		streamHandler = createStreamingHandler(formatter, opts)
		streamedOutput = true
	}

	findings, err := js.AnalyzeStreaming(ctx, cfg, streamHandler)
	return findings, streamedOutput, err
}

func createStreamingHandler(formatter report.Formatter, opts scanFlags) js.StreamHandler {
	return func(batchFindings []rules.Finding) error {
		filtered := report.FilterByConfidence(batchFindings, opts.minConfidence)
		data, err := formatter.Format(filtered)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(data)
		return err
	}
}

func outputResults(findings []rules.Finding, opts scanFlags, cfg *config.Config, start time.Time) error {
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
		return nil, fmt.Errorf("unsupported language: %q", cfg.Project.Language)
	}
}
