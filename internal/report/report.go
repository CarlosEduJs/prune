package report

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"prune/internal/config"
	"prune/internal/rules"
)

type FormatterOptions struct {
	Duration  time.Duration
	Compact   bool
	Only      string
	Deletable bool
	Config    *config.Config
}

type Formatter interface {
	Format([]rules.Finding) ([]byte, error)
}

func NewFormatter(format string, opts ...FormatterOptions) (Formatter, error) {
	o := FormatterOptions{}
	if len(opts) > 0 {
		o = opts[0]
	}

	switch strings.ToLower(format) {
	case "json":
		return jsonFormatter{opts: o}, nil
	case "ndjson":
		return ndjsonFormatter{}, nil
	case "table", "pretty":
		return prettyFormatter{opts: o}, nil
	default:
		return nil, fmt.Errorf("unknown format: %q", format)
	}
}

func FilterByConfidence(findings []rules.Finding, min string) []rules.Finding {
	minRank := rules.ConfidenceRank(min)
	if minRank == 0 {
		return findings
	}

	filtered := make([]rules.Finding, 0, len(findings))
	for _, f := range findings {
		if rules.ConfidenceRank(f.Confidence) >= minRank {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

type JSONReport struct {
	Summary  Summary         `json:"summary"`
	Findings []rules.Finding `json:"findings"`
	Metadata Metadata        `json:"metadata"`
}

type Summary struct {
	Files  int `json:"files"`
	Issues int `json:"issues"`
	Safe   int `json:"safe"`
	Review int `json:"review"`
}

type Metadata struct {
	Entrypoints []string `json:"entrypoints"`
	ScanTimeMs  int64    `json:"scan_time_ms"`
}

func buildSummary(findings []rules.Finding) Summary {
	uniqueFiles := map[string]bool{}
	safeCount := 0
	reviewCount := 0

	for _, f := range findings {
		if f.File != "" {
			uniqueFiles[f.File] = true
		}
		switch f.Confidence {
		case "safe":
			safeCount++
		case "likely_dead", "review":
			reviewCount++
		}
	}

	return Summary{
		Files:  len(uniqueFiles),
		Issues: len(findings),
		Safe:   safeCount,
		Review: reviewCount,
	}
}

func buildMetadata(opts FormatterOptions) Metadata {
	entrypoints := []string{}
	if opts.Config != nil {
		entrypoints = append(entrypoints, opts.Config.Entrypoints.Files...)
		entrypoints = append(entrypoints, opts.Config.Entrypoints.Patterns...)
	}
	return Metadata{
		Entrypoints: entrypoints,
		ScanTimeMs:  opts.Duration.Milliseconds(),
	}
}

type jsonFormatter struct {
	opts FormatterOptions
}

func (f jsonFormatter) Format(findings []rules.Finding) ([]byte, error) {
	summary := buildSummary(findings)
	metadata := buildMetadata(f.opts)
	report := JSONReport{
		Summary:  summary,
		Findings: findings,
		Metadata: metadata,
	}
	return json.MarshalIndent(report, "", "  ")
}

type ndjsonFormatter struct{}

func (f ndjsonFormatter) Format(findings []rules.Finding) ([]byte, error) {
	var b strings.Builder
	for _, finding := range findings {
		data, err := json.Marshal(finding)
		if err != nil {
			return nil, err
		}
		b.Write(data)
		b.WriteString("\n")
	}
	return []byte(b.String()), nil
}
