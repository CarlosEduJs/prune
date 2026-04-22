package js

import (
	"testing"

	"prune/internal/config"
	"prune/internal/rules"
	"prune/internal/scan"
)

// TestExportedFunctionSkippedByUnusedSymbols verifies that exported functions
// are NOT reported by ruleUnusedSymbols (they are handled by ruleUnusedExports instead).
func TestExportedFunctionSkippedByUnusedSymbols(t *testing.T) {
	cfg := &config.Config{}
	cfg.Rules = map[string]config.RuleConfig{
		"unused_function": {Enabled: true},
		"unused_variable": {Enabled: true},
	}

	data := &Collected{
		FunctionDecls: map[string][]string{"src/a.ts": {"exportedFn", "localFn"}},
		VariableDecls: map[string][]string{},
		Exports:       map[string][]string{"src/a.ts": {"exportedFn"}},
		Identifiers:   map[string]map[string]int{"src/a.ts": {"exportedFn": 1, "localFn": 1}},
		UsageCounts:   map[string]map[string]int{"src/a.ts": {}},
		FunctionLines: map[string]map[string]int{"src/a.ts": {"exportedFn": 1, "localFn": 5}},
		VariableLines: map[string]map[string]int{},
		DynamicIndicators: map[string][]string{},
	}

	findings := ruleUnusedSymbols(cfg, data)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding (only localFn), got %d", len(findings))
	}
	if findings[0].Symbol != "localFn" {
		t.Fatalf("expected finding for localFn, got %s", findings[0].Symbol)
	}
	if findings[0].Kind != "unused_function" {
		t.Fatalf("expected kind unused_function, got %s", findings[0].Kind)
	}
}

// TestExportedVariableSkippedByUnusedSymbols verifies that exported variables
// are NOT reported by ruleUnusedSymbols.
func TestExportedVariableSkippedByUnusedSymbols(t *testing.T) {
	cfg := &config.Config{}
	cfg.Rules = map[string]config.RuleConfig{
		"unused_function": {Enabled: true},
		"unused_variable": {Enabled: true},
	}

	data := &Collected{
		FunctionDecls: map[string][]string{},
		VariableDecls: map[string][]string{"src/a.ts": {"exportedVar", "localVar"}},
		Exports:       map[string][]string{"src/a.ts": {"exportedVar"}},
		Identifiers:   map[string]map[string]int{"src/a.ts": {"exportedVar": 1, "localVar": 1}},
		UsageCounts:   map[string]map[string]int{"src/a.ts": {}},
		FunctionLines: map[string]map[string]int{},
		VariableLines: map[string]map[string]int{"src/a.ts": {"exportedVar": 1, "localVar": 3}},
		DynamicIndicators: map[string][]string{},
	}

	findings := ruleUnusedSymbols(cfg, data)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding (only localVar), got %d", len(findings))
	}
	if findings[0].Symbol != "localVar" {
		t.Fatalf("expected finding for localVar, got %s", findings[0].Symbol)
	}
}

// TestNoDuplicateBetweenExportAndFunction verifies that a symbol that is both
// declared as a function and exported only appears once via ruleUnusedExports,
// not also via ruleUnusedSymbols.
func TestNoDuplicateBetweenExportAndFunction(t *testing.T) {
	cfg := &config.Config{}
	cfg.Rules = map[string]config.RuleConfig{
		"unused_export":   {Enabled: true, Confidence: map[string]string{"if_not_imported": "safe"}},
		"unused_function": {Enabled: true},
	}

	data := &Collected{
		FunctionDecls:   map[string][]string{"src/a.ts": {"myFunc"}},
		VariableDecls:   map[string][]string{},
		Exports:         map[string][]string{"src/a.ts": {"myFunc"}},
		ExportSymbols:   map[string][]ExportSymbol{"src/a.ts": {{Name: "myFunc", Line: 3}}},
		ImportSpecs:     map[string][]ImportSpec{},
		Identifiers:     map[string]map[string]int{"src/a.ts": {"myFunc": 1}},
		UsageCounts:     map[string]map[string]int{"src/a.ts": {}},
		FunctionLines:   map[string]map[string]int{"src/a.ts": {"myFunc": 3}},
		VariableLines:   map[string]map[string]int{},
		Files:           []scan.FileEntry{{Rel: "src/a.ts"}},
		DynamicIndicators: map[string][]string{},
	}

	allFindings := applyRules(cfg, data)

	count := 0
	for _, f := range allFindings {
		if f.Symbol == "myFunc" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 finding for myFunc, got %d", count)
	}

	for _, f := range allFindings {
		if f.Symbol == "myFunc" && f.Kind != "unused_export" {
			t.Fatalf("expected myFunc finding to be unused_export, got %s", f.Kind)
		}
	}
}

// TestDefaultExportInEntrypointSuppressed verifies that the default export
// in an entrypoint file (e.g., Next.js page.tsx) is never reported.
func TestDefaultExportInEntrypointSuppressed(t *testing.T) {
	cfg := &config.Config{}
	cfg.Rules = map[string]config.RuleConfig{
		"unused_export": {Enabled: true, Confidence: map[string]string{
			"if_not_imported": "safe",
			"if_entrypoint":   "review",
		}},
	}
	cfg.Entrypoints.Patterns = []string{"apps/**/page.tsx"}

	data := &Collected{
		Exports:       map[string][]string{"apps/web/page.tsx": {"default", "unusedNamedExport"}},
		ExportSymbols: map[string][]ExportSymbol{"apps/web/page.tsx": {{Name: "default", Line: 1}, {Name: "unusedNamedExport", Line: 5}}},
		ImportSpecs:   map[string][]ImportSpec{},
		Files:         []scan.FileEntry{{Rel: "apps/web/page.tsx"}},
	}

	findings := ruleUnusedExports(cfg, data)

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
	if findings[0].Symbol != "unusedNamedExport" {
		t.Fatalf("expected finding for unusedNamedExport, got %s", findings[0].Symbol)
	}
	if findings[0].Confidence != "review" {
		t.Fatalf("expected confidence review for entrypoint export, got %s", findings[0].Confidence)
	}
}

// TestDefaultExportInNonEntrypointReported verifies that default export
// in a non-entrypoint file IS reported as unused.
func TestDefaultExportInNonEntrypointReported(t *testing.T) {
	cfg := &config.Config{}
	cfg.Rules = map[string]config.RuleConfig{
		"unused_export": {Enabled: true, Confidence: map[string]string{"if_not_imported": "safe"}},
	}
	cfg.Entrypoints.Patterns = []string{"apps/**/page.tsx"}

	data := &Collected{
		Exports:       map[string][]string{"src/utils.ts": {"default"}},
		ExportSymbols: map[string][]ExportSymbol{"src/utils.ts": {{Name: "default", Line: 1}}},
		ImportSpecs:   map[string][]ImportSpec{},
		Files:         []scan.FileEntry{{Rel: "src/utils.ts"}},
	}

	findings := ruleUnusedExports(cfg, data)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for non-entrypoint default export, got %d", len(findings))
	}
	if findings[0].Symbol != "default" {
		t.Fatalf("expected finding symbol 'default', got %s", findings[0].Symbol)
	}
}

// TestDeepGlobstarEntrypointMatch verifies that deeply nested paths
// (like Next.js App Router) correctly match globstar patterns.
func TestDeepGlobstarEntrypointMatch(t *testing.T) {
	cfg := &config.Config{}
	cfg.Rules = map[string]config.RuleConfig{
		"unused_export": {Enabled: true, Confidence: map[string]string{
			"if_not_imported": "safe",
			"if_entrypoint":   "review",
		}},
	}
	cfg.Entrypoints.Patterns = []string{"apps/web/src/app/**/page.tsx"}

	deepPath := "apps/web/src/app/workspaces/[slug]/teams/[team]/page.tsx"
	data := &Collected{
		Exports:       map[string][]string{deepPath: {"default", "namedExport"}},
		ExportSymbols: map[string][]ExportSymbol{deepPath: {{Name: "default", Line: 1}, {Name: "namedExport", Line: 10}}},
		ImportSpecs:   map[string][]ImportSpec{},
		Files:         []scan.FileEntry{{Rel: deepPath}},
	}

	findings := ruleUnusedExports(cfg, data)

	// default should be suppressed; namedExport gets review confidence
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
	if findings[0].Symbol != "namedExport" {
		t.Fatalf("expected namedExport, got %s", findings[0].Symbol)
	}
	if findings[0].Confidence != "review" {
		t.Fatalf("expected review confidence, got %s", findings[0].Confidence)
	}
}

// TestUnusedFileSuppressesAllSubFindings verifies that when a file is detected
// as unused, all other findings for that file are suppressed by applyRules + dedup.
func TestUnusedFileSuppressesAllSubFindings(t *testing.T) {
	cfg := &config.Config{}
	cfg.Rules = map[string]config.RuleConfig{
		"unused_file":     {Enabled: true},
		"unused_export":   {Enabled: true, Confidence: map[string]string{"if_not_imported": "safe"}},
		"unused_function": {Enabled: true},
		"unused_variable": {Enabled: true},
	}

	data := &Collected{
		Files:           []scan.FileEntry{{Rel: "src/dead.ts"}, {Rel: "src/alive.ts"}},
		ImportsResolved: map[string][]string{"src/alive.ts": {}},
		Exports:         map[string][]string{"src/dead.ts": {"deadExport"}},
		ExportSymbols:   map[string][]ExportSymbol{"src/dead.ts": {{Name: "deadExport", Line: 1}}},
		FunctionDecls:   map[string][]string{"src/dead.ts": {"deadFn"}},
		VariableDecls:   map[string][]string{},
		ImportSpecs:     map[string][]ImportSpec{},
		Identifiers:     map[string]map[string]int{"src/dead.ts": {"deadFn": 1}},
		UsageCounts:     map[string]map[string]int{"src/dead.ts": {}},
		FunctionLines:   map[string]map[string]int{"src/dead.ts": {"deadFn": 5}},
		VariableLines:   map[string]map[string]int{},
		DynamicIndicators: map[string][]string{},
	}

	allFindings := applyRules(cfg, data)

	// After dedup in the formatter, only unused_file should remain for dead.ts
	deduped := deduplicateFindings(allFindings)

	for _, f := range deduped {
		if f.File == "src/dead.ts" && f.Kind != "unused_file" {
			t.Fatalf("expected only unused_file for dead.ts after dedup, got %s:%s", f.Kind, f.Symbol)
		}
	}

	count := 0
	for _, f := range deduped {
		if f.File == "src/dead.ts" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 finding for dead.ts, got %d", count)
	}
}

// TestHighRiskDynamicDemotesExportConfidence verifies that unused exports
// in files with eval/Function get demoted to review confidence.
func TestHighRiskDynamicDemotesExportConfidence(t *testing.T) {
	cfg := &config.Config{}
	cfg.Rules = map[string]config.RuleConfig{
		"unused_export": {Enabled: true, Confidence: map[string]string{
			"if_not_imported":      "safe",
			"if_high_risk_dynamic": "review",
		}},
	}

	data := &Collected{
		Exports:           map[string][]string{"src/a.ts": {"myExport"}},
		ExportSymbols:     map[string][]ExportSymbol{"src/a.ts": {{Name: "myExport", Line: 3}}},
		ImportSpecs:       map[string][]ImportSpec{},
		Files:             []scan.FileEntry{{Rel: "src/a.ts"}},
		DynamicIndicators: map[string][]string{"src/a.ts": {"eval"}},
	}

	findings := ruleUnusedExports(cfg, data)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Confidence != "review" {
		t.Fatalf("expected review confidence due to eval, got %s", findings[0].Confidence)
	}
}

// deduplicateFindings mimics what the pretty formatter does —
// it suppresses sub-findings for files that have an unused_file finding.
func deduplicateFindings(findings []rules.Finding) []rules.Finding {
	unusedFiles := map[string]bool{}
	for _, f := range findings {
		if f.Kind == "unused_file" {
			unusedFiles[f.File] = true
		}
	}
	result := []rules.Finding{}
	for _, f := range findings {
		if unusedFiles[f.File] && f.Kind != "unused_file" {
			continue
		}
		result = append(result, f)
	}
	return result
}
