package report

import (
	"encoding/json"
	"fmt"
	"strings"

	"prune/internal/rules"
)

type Formatter interface {
	Format([]rules.Finding) ([]byte, error)
}

func NewFormatter(format string) (Formatter, error) {
	switch strings.ToLower(format) {
	case "json":
		return jsonFormatter{}, nil
	case "table":
		return tableFormatter{}, nil
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

type tableFormatter struct{}

func (f tableFormatter) Format(findings []rules.Finding) ([]byte, error) {
	if len(findings) == 0 {
		return []byte("No findings\n"), nil
	}

	var b strings.Builder
	_, _ = fmt.Fprintln(&b, "confidence\tkind\tfile\tline\tsymbol\treason")
	for _, finding := range findings {
		file := finding.File
		if file == "" {
			file = "-"
		}
		_, _ = fmt.Fprintf(
			&b,
			"%s\t%s\t%s\t%d\t%s\t%s\n",
			finding.Confidence,
			finding.Kind,
			file,
			finding.Line,
			finding.Symbol,
			finding.Reason,
		)
	}
	return []byte(b.String()), nil
}
