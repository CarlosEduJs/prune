package js

import (
	"testing"

	"prune/internal/config"
	"prune/internal/scan"
)

func TestUnusedExportUsesLines(t *testing.T) {
	cfg := &config.Config{}
	cfg.Rules = map[string]config.RuleConfig{
		"unused_export": {Enabled: true, Confidence: map[string]string{"if_not_imported": "safe"}},
	}

	data := &Collected{
		Exports:       map[string][]string{"src/a.ts": {"one"}},
		ExportSymbols: map[string][]ExportSymbol{"src/a.ts": {{Name: "one", Line: 7}}},
		ImportSpecs:   map[string][]ImportSpec{},
		Files:         []scan.FileEntry{{Rel: "src/a.ts"}},
	}

	findings := ruleUnusedExports(cfg, data)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 7 {
		t.Fatalf("expected line 7, got %d", findings[0].Line)
	}
}

func TestUnusedSymbolsUsesUsageCounts(t *testing.T) {
	cfg := &config.Config{}
	cfg.Rules = map[string]config.RuleConfig{
		"unused_function": {Enabled: true},
		"unused_variable": {Enabled: true},
	}

	data := &Collected{
		FunctionDecls: map[string][]string{"src/a.ts": {"used", "unused"}},
		VariableDecls: map[string][]string{"src/a.ts": {"alive", "dead"}},
		Identifiers:   map[string]map[string]int{"src/a.ts": {"used": 2, "unused": 1, "alive": 2, "dead": 1}},
		UsageCounts:   map[string]map[string]int{"src/a.ts": {"used": 1, "alive": 1}},
		DynamicIndicators: map[string][]string{
			"src/a.ts": {},
		},
	}

	findings := ruleUnusedSymbols(cfg, data)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}

func TestUnusedFileRespectsEntrypointPattern(t *testing.T) {
	cfg := &config.Config{}
	cfg.Rules = map[string]config.RuleConfig{"unused_file": {Enabled: true}}
	cfg.Entrypoints.Patterns = []string{"src/routes/**"}

	data := &Collected{
		Files:           []scan.FileEntry{{Rel: "src/routes/index.ts"}, {Rel: "src/other.ts"}},
		ImportsResolved: map[string][]string{},
	}

	findings := ruleUnusedFiles(cfg, data)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].File != "src/other.ts" {
		t.Fatalf("unexpected finding file %s", findings[0].File)
	}
}
