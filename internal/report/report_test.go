package report

import (
	"encoding/json"
	"strings"
	"testing"

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
	var decoded []rules.Finding
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}
	if len(decoded) != 1 || decoded[0].ID != "f1" {
		t.Fatalf("unexpected json output")
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
