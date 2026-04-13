package report

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"prune/internal/rules"
)

var version = "0.0.3"

type FormatterOptions struct {
	Duration  time.Duration
	Compact   bool
	Only      string
	Deletable bool
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
		return jsonFormatter{}, nil
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

type jsonFormatter struct{}

func (f jsonFormatter) Format(findings []rules.Finding) ([]byte, error) {
	return json.MarshalIndent(findings, "", "  ")
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

type tableFormatter struct{}

func (f tableFormatter) Format(findings []rules.Finding) ([]byte, error) {
	if len(findings) == 0 {
		return []byte("✨ No dead code found!\n"), nil
	}

	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, "CONFIDENCE\tKIND\tFILE\tLINE\tSYMBOL\tREASON")
	for _, finding := range findings {
		file := finding.File
		if file == "" {
			file = "-"
		}
		_, _ = fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%d\t%s\t%s\n",
			strings.ToUpper(finding.Confidence),
			finding.Kind,
			file,
			finding.Line,
			finding.Symbol,
			finding.Reason,
		)
	}
	_ = w.Flush()
	return []byte(b.String()), nil
}
