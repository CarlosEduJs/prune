package report

import (
	"strings"
	"testing"
	"time"

	"prune/internal/rules"
)

func TestDeduplicateUnusedFiles(t *testing.T) {
	findings := []rules.Finding{
		{Kind: "unused_file", File: "src/dead.js", Confidence: "safe"},
		{Kind: "unused_export", File: "src/dead.js", Confidence: "safe", Symbol: "foo"},
		{Kind: "unused_function", File: "src/dead.js", Confidence: "safe", Symbol: "bar"},
		{Kind: "unused_export", File: "src/alive.js", Confidence: "review", Symbol: "baz"},
	}

	result := deduplicateUnusedFiles(findings)
	if len(result) != 2 {
		t.Fatalf("expected 2 findings after dedup, got %d", len(result))
	}

	for _, f := range result {
		if f.File == "src/dead.js" && f.Kind != "unused_file" {
			t.Fatalf("expected only unused_file for dead.js, got %s", f.Kind)
		}
	}
}

func TestGroupByConfidence(t *testing.T) {
	findings := []rules.Finding{
		{Confidence: "safe", File: "a.js", Kind: "unused_export"},
		{Confidence: "safe", File: "b.js", Kind: "unused_function"},
		{Confidence: "review", File: "c.js", Kind: "suspicious_dynamic_usage"},
	}

	grouped := groupByConfidence(findings)
	if len(grouped["safe"]) != 2 {
		t.Fatalf("expected 2 files in safe group, got %d", len(grouped["safe"]))
	}
	if len(grouped["review"]) != 1 {
		t.Fatalf("expected 1 file in review group, got %d", len(grouped["review"]))
	}
}

func TestPrettyCompactMode(t *testing.T) {
	formatter, err := NewFormatter("pretty", FormatterOptions{
		Duration: 5 * time.Millisecond,
		Compact:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	findings := []rules.Finding{
		{Confidence: "safe", Kind: "unused_export", File: "a.js", Symbol: "x"},
		{Confidence: "safe", Kind: "unused_function", File: "b.js", Symbol: "y"},
		{Confidence: "review", Kind: "suspicious_dynamic_usage", File: "c.js", Symbol: "z"},
	}

	data, err := formatter.Format(findings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := string(data)
	// Compact mode should include the summary but NOT the individual file details.
	if !strings.Contains(output, "Summary") {
		t.Fatalf("expected Summary section, got: %s", output)
	}
	// Should NOT have tree markers in compact output.
	if strings.Contains(output, "└─") {
		t.Fatalf("compact mode should not show tree details, got: %s", output)
	}
}

func TestPrettyOnlyFilter(t *testing.T) {
	formatter, err := NewFormatter("pretty", FormatterOptions{
		Duration: 5 * time.Millisecond,
		Only:     "safe",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	findings := []rules.Finding{
		{Confidence: "safe", Kind: "unused_export", File: "a.js", Symbol: "x"},
		{Confidence: "review", Kind: "suspicious_dynamic_usage", File: "c.js", Symbol: "z"},
	}

	data, err := formatter.Format(findings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "a.js") {
		t.Fatalf("expected safe finding a.js in output, got: %s", output)
	}
	if strings.Contains(output, "c.js") {
		t.Fatalf("review finding c.js should be filtered out, got: %s", output)
	}
}

func TestPrettyDeletableFilter(t *testing.T) {
	formatter, err := NewFormatter("pretty", FormatterOptions{
		Duration:  5 * time.Millisecond,
		Deletable: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	findings := []rules.Finding{
		{Confidence: "safe", Kind: "unused_file", File: "src/dead.js"},
		{Confidence: "safe", Kind: "unused_export", File: "src/alive.js", Symbol: "x"},
		{Confidence: "review", Kind: "unused_file", File: "src/maybe.js"},
	}

	data, err := formatter.Format(findings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "src/dead.js") {
		t.Fatalf("expected safe unused_file in output, got: %s", output)
	}
	if strings.Contains(output, "src/alive.js") {
		t.Fatalf("non-unused-file findings should be filtered, got: %s", output)
	}
	if strings.Contains(output, "src/maybe.js") {
		t.Fatalf("non-safe unused_file should be filtered, got: %s", output)
	}
}

func TestPrettyGroupingOrder(t *testing.T) {
	formatter, err := NewFormatter("pretty", FormatterOptions{
		Duration: 5 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	findings := []rules.Finding{
		{Confidence: "review", Kind: "suspicious_dynamic_usage", File: "z.js", Symbol: "a"},
		{Confidence: "safe", Kind: "unused_export", File: "b.js", Symbol: "x"},
		{Confidence: "safe", Kind: "unused_function", File: "a.js", Symbol: "y"},
	}

	data, err := formatter.Format(findings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := string(data)
	safeIdx := strings.Index(output, "SAFE")
	reviewIdx := strings.Index(output, "REVIEW")
	if safeIdx == -1 || reviewIdx == -1 {
		t.Fatalf("expected both SAFE and REVIEW sections, got: %s", output)
	}
	if safeIdx > reviewIdx {
		t.Fatalf("SAFE should appear before REVIEW, got: %s", output)
	}

	// Inside SAFE section, files should be sorted alphabetically.
	aIdx := strings.Index(output, "a.js")
	bIdx := strings.Index(output, "b.js")
	if aIdx == -1 || bIdx == -1 {
		t.Fatalf("expected both a.js and b.js, got: %s", output)
	}
	if aIdx > bIdx {
		t.Fatalf("a.js should appear before b.js, got: %s", output)
	}
}
