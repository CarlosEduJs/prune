package report

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"prune/internal/config"
	"prune/internal/rules"
)

func TestNewFormatter(t *testing.T) {
	if f, err := NewFormatter("json"); err != nil || f == nil {
		t.Fatalf("expected json formatter, got err: %v", err)
	}
	if f, err := NewFormatter("pretty"); err != nil || f == nil {
		t.Fatalf("expected pretty formatter, got err: %v", err)
	}
	if f, err := NewFormatter("table"); err != nil || f == nil {
		t.Fatalf("expected table (alias for pretty) formatter, got err: %v", err)
	}
	if f, err := NewFormatter("unknown"); err == nil || f != nil {
		t.Fatalf("expected error for unknown formatter")
	}
}

func TestFilterByConfidence(t *testing.T) {
	findings := []rules.Finding{
		{Confidence: "safe"},
		{Confidence: "likely_dead"},
		{Confidence: "review"},
	}

	filtered := FilterByConfidence(findings, "safe")
	if len(filtered) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(filtered))
	}

	filtered = FilterByConfidence(findings, "likely_dead")
	if len(filtered) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(filtered))
	}

	filtered = FilterByConfidence(findings, "review")
	if len(filtered) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(filtered))
	}

	filtered = FilterByConfidence(findings, "unknown")
	if len(filtered) != 3 {
		t.Fatalf("expected 3 findings for unknown confidence, got %d", len(filtered))
	}
}

func TestJSONFormatter(t *testing.T) {
	formatter, err := NewFormatter("json")
	if err != nil || formatter == nil {
		t.Fatalf("missing json formatter, err: %v", err)
	}

	findings := []rules.Finding{{ID: "f1", Confidence: "safe", File: "a.js", Line: 1}}
	data, err := formatter.Format(findings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var report JSONReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}
	if len(report.Findings) != 1 || report.Findings[0].ID != "f1" {
		t.Fatalf("unexpected json output")
	}
	if report.Summary.Files != 1 {
		t.Fatalf("expected 1 file in summary, got %d", report.Summary.Files)
	}
	if report.Summary.Issues != 1 {
		t.Fatalf("expected 1 issue in summary, got %d", report.Summary.Issues)
	}
}

func TestJSONFormatterSummary(t *testing.T) {
	formatter, err := NewFormatter("json", FormatterOptions{
		Duration: 984 * time.Millisecond,
	})
	if err != nil || formatter == nil {
		t.Fatalf("missing json formatter, err: %v", err)
	}

	findings := []rules.Finding{
		{Confidence: "safe", Kind: "unused_file", File: "a.ts"},
		{Confidence: "safe", Kind: "unused_file", File: "b.ts"},
		{Confidence: "review", Kind: "suspicious_dynamic_usage", File: "c.ts"},
		{Confidence: "likely_dead", Kind: "unused_function", File: "d.ts"},
	}
	data, err := formatter.Format(findings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var report JSONReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}

	if report.Summary.Files != 4 {
		t.Errorf("expected 4 files, got %d", report.Summary.Files)
	}
	if report.Summary.Issues != 4 {
		t.Errorf("expected 4 issues, got %d", report.Summary.Issues)
	}
	if report.Summary.Safe != 2 {
		t.Errorf("expected 2 safe, got %d", report.Summary.Safe)
	}
	if report.Summary.Review != 2 {
		t.Errorf("expected 2 review (includes likely_dead), got %d", report.Summary.Review)
	}
	if report.Metadata.ScanTimeMs != 984 {
		t.Errorf("expected scan_time_ms 984, got %d", report.Metadata.ScanTimeMs)
	}
}

func TestJSONFormatterWithConfig(t *testing.T) {
	cfg := &config.Config{
		Version: 1,
	}
	cfg.Entrypoints.Files = []string{"src/index.ts", "src/main.tsx"}
	cfg.Entrypoints.Patterns = []string{"src/pages/**"}

	formatter, err := NewFormatter("json", FormatterOptions{
		Config: cfg,
	})
	if err != nil || formatter == nil {
		t.Fatalf("missing json formatter, err: %v", err)
	}

	data, err := formatter.Format(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var report JSONReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}

	if len(report.Metadata.Entrypoints) != 3 {
		t.Errorf("expected 3 entrypoints, got %d: %v", len(report.Metadata.Entrypoints), report.Metadata.Entrypoints)
	}
}

func TestJSONFormatterNoFindings(t *testing.T) {
	formatter, err := NewFormatter("json", FormatterOptions{
		Duration: 100 * time.Millisecond,
	})
	if err != nil || formatter == nil {
		t.Fatalf("missing json formatter, err: %v", err)
	}

	data, err := formatter.Format(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var report JSONReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}

	if report.Summary.Files != 0 {
		t.Errorf("expected 0 files, got %d", report.Summary.Files)
	}
	if report.Summary.Issues != 0 {
		t.Errorf("expected 0 issues, got %d", report.Summary.Issues)
	}
	if report.Summary.Safe != 0 {
		t.Errorf("expected 0 safe, got %d", report.Summary.Safe)
	}
	if report.Summary.Review != 0 {
		t.Errorf("expected 0 review, got %d", report.Summary.Review)
	}
}

func TestPrettyFormatterNoFindings(t *testing.T) {
	formatter, err := NewFormatter("pretty")
	if err != nil || formatter == nil {
		t.Fatalf("missing pretty formatter, err: %v", err)
	}

	data, err := formatter.Format(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(data), "No dead code found") {
		t.Fatalf("unexpected no-findings output: %s", string(data))
	}
}

func TestPrettyFormatterWithFindings(t *testing.T) {
	formatter, err := NewFormatter("pretty")
	if err != nil || formatter == nil {
		t.Fatalf("missing pretty formatter, err: %v", err)
	}

	findings := []rules.Finding{{Confidence: "safe", Kind: "unused_export", File: "a.js", Line: 2, Symbol: "x", Reason: "r"}}
	data, err := formatter.Format(findings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := string(data)
	if !strings.Contains(output, "SAFE") {
		t.Fatalf("missing SAFE section")
	}
	if !strings.Contains(output, "a.js") {
		t.Fatalf("missing file path")
	}
	if !strings.Contains(output, "unused export: x") {
		t.Fatalf("missing finding detail, got: %s", output)
	}
}

func TestPrettyFormatterMissingFile(t *testing.T) {
	formatter, err := NewFormatter("pretty")
	if err != nil || formatter == nil {
		t.Fatalf("missing pretty formatter, err: %v", err)
	}
	data, err := formatter.Format([]rules.Finding{{Confidence: "review", Kind: "possible_dynamic_usage"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(data), "REVIEW") {
		t.Fatalf("expected REVIEW section in output")
	}
}
