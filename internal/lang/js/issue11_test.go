package js

import (
	"testing"

	"prune/internal/config"
	"prune/internal/scan"
)

func TestIssue11ScopedAlias(t *testing.T) {
	cfg := &config.Config{
		TsConfig: config.TsConfig{
			Enabled: true,
			BaseURL: ".",
			Paths: map[string][]string{
				"@scope/*":      {"others/*"},
				"@scope/core/*": {"libs/*"},
			},
		},
	}
	fileIndex := map[string]scan.FileEntry{
		"libs/util.ts":         {Path: "libs/util.ts", Rel: "libs/util.ts"},
		"others/core/util.ts":  {Path: "others/core/util.ts", Rel: "others/core/util.ts"},
		"others/something.ts":  {Path: "others/something.ts", Rel: "others/something.ts"},
	}

	r := NewResolver(cfg, fileIndex)

	tests := []struct {
		name     string
		source   string
		wantRes  string
		wantType ImportType
		wantConf string
	}{
		{
			name:     "longest prefix wins: @scope/core/* over @scope/*",
			source:   "@scope/core/util",
			wantRes:  "libs/util.ts",
			wantType: ImportTypeAlias,
			wantConf: "safe",
		},
		{
			name:     "shorter prefix matches when no longer prefix applies",
			source:   "@scope/something",
			wantRes:  "others/something.ts",
			wantType: ImportTypeAlias,
			wantConf: "safe",
		},
		{
			name:     "non-aliased scoped package is external",
			source:   "@not-aliased/pkg",
			wantRes:  "",
			wantType: ImportTypeExternal,
			wantConf: "safe",
		},
		{
			name:     "alias that does not resolve returns review",
			source:   "@scope/core/nonexistent",
			wantRes:  "",
			wantType: ImportTypeAlias,
			wantConf: "review",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.Resolve(tt.source, "src/main.ts")
			if got.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.Resolved != tt.wantRes {
				t.Errorf("Resolved = %q, want %q", got.Resolved, tt.wantRes)
			}
			if got.Confidence != tt.wantConf {
				t.Errorf("Confidence = %q, want %q", got.Confidence, tt.wantConf)
			}
		})
	}
}

func TestIssue11AliasFilesTrackedAsUsed(t *testing.T) {
	// Verify that files imported via scoped aliases are included in
	// ImportsResolved and therefore NOT flagged as unused_file.
	specs := []ImportSpec{
		{Source: "@scope/core/util", Resolved: "libs/util.ts", Confidence: "safe"},
		{Source: "./local", Resolved: "", Confidence: ""},
	}
	fileIndex := map[string]scan.FileEntry{
		"libs/util.ts":  {Path: "libs/util.ts", Rel: "libs/util.ts"},
		"src/local.ts":  {Path: "src/local.ts", Rel: "src/local.ts"},
	}

	resolved := resolveLocalImports("src/main.ts", specs, fileIndex)

	want := map[string]bool{
		"libs/util.ts": true,
		"src/local.ts": true,
	}

	got := map[string]bool{}
	for _, r := range resolved {
		got[r] = true
	}

	for file := range want {
		if !got[file] {
			t.Errorf("expected %q in resolved imports, but it was missing", file)
		}
	}
}

func TestIssue11Determinism(t *testing.T) {
	fileIndex := map[string]scan.FileEntry{
		"libs/util.ts":        {Path: "libs/util.ts", Rel: "libs/util.ts"},
		"others/core/util.ts": {Path: "others/core/util.ts", Rel: "others/core/util.ts"},
	}

	for i := 0; i < 100; i++ {
		cfg := &config.Config{
			TsConfig: config.TsConfig{
				Enabled: true,
				BaseURL: ".",
				Paths: map[string][]string{
					"@scope/*":      {"others/*"},
					"@scope/core/*": {"libs/*"},
				},
			},
		}
		r := NewResolver(cfg, fileIndex)
		got := r.Resolve("@scope/core/util", "src/main.ts")

		if got.Resolved != "libs/util.ts" {
			t.Fatalf("iteration %d: Resolved = %q, want %q (determinism broken)", i, got.Resolved, "libs/util.ts")
		}
		if got.Type != ImportTypeAlias {
			t.Fatalf("iteration %d: Type = %v, want %v", i, got.Type, ImportTypeAlias)
		}
	}
}

func TestIssue11PrefixBoundary(t *testing.T) {
	cfg := &config.Config{
		TsConfig: config.TsConfig{
			Enabled: true,
			BaseURL: ".",
			Paths: map[string][]string{
				"@scope/*": {"src/*"},
			},
		},
	}
	fileIndex := map[string]scan.FileEntry{
		"src/util.ts": {Path: "src/util.ts", Rel: "src/util.ts"},
	}

	r := NewResolver(cfg, fileIndex)

	tests := []struct {
		name     string
		source   string
		wantType ImportType
	}{
		{
			name:     "@scope/util is an alias",
			source:   "@scope/util",
			wantType: ImportTypeAlias,
		},
		{
			name:     "bare @scope is NOT an alias — external",
			source:   "@scope",
			wantType: ImportTypeExternal,
		},
		{
			name:     "@scope-core/util is NOT an alias — external",
			source:   "@scope-core/util",
			wantType: ImportTypeExternal,
		},
		{
			name:     "@scoped/util is NOT an alias — external",
			source:   "@scoped/util",
			wantType: ImportTypeExternal,
		},
		{
			name:     "@scope-ui/components is NOT an alias — external",
			source:   "@scope-ui/components",
			wantType: ImportTypeExternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.Classify(tt.source)
			if got != tt.wantType {
				t.Errorf("Classify(%q) = %v, want %v", tt.source, got, tt.wantType)
			}
		})
	}
}

func TestIssue11ResolverCollectorIntegration(t *testing.T) {
	cfg := &config.Config{
		Version: 1,
		TsConfig: config.TsConfig{
			Enabled: true,
			BaseURL: ".",
			Paths: map[string][]string{
				"@lib/*": {"libs/*"},
			},
		},
		Rules: map[string]config.RuleConfig{
			"unused_file":   {Enabled: true},
			"unused_export": {Enabled: false},
			"unused_function": {Enabled: false},
			"unused_variable": {Enabled: false},
			"possible_dynamic_usage": {Enabled: false},
			"suspicious_dynamic_usage": {Enabled: false},
		},
	}

	entries := []scan.FileEntry{
		{Path: "src/main.ts", Rel: "src/main.ts"},
		{Path: "libs/helper.ts", Rel: "libs/helper.ts"},
	}

	fileIndex := map[string]scan.FileEntry{
		"src/main.ts":    entries[0],
		"libs/helper.ts": entries[1],
	}

	resolver := NewResolver(cfg, fileIndex)
	importSource := "@lib/helper"
	resolved := resolver.Resolve(importSource, "src/main.ts")

	if resolved.Resolved != "libs/helper.ts" {
		t.Fatalf("resolver: Resolved = %q, want %q", resolved.Resolved, "libs/helper.ts")
	}

	specs := []ImportSpec{
		{
			Source:     importSource,
			Resolved:   resolved.Resolved,
			Confidence: resolved.Confidence,
			Names:      []string{"helper"},
		},
	}
	importsResolved := resolveLocalImports("src/main.ts", specs, fileIndex)

	collected := &Collected{
		Files:           entries,
		Imports:         map[string][]string{"src/main.ts": {importSource}},
		ImportSpecs:     map[string][]ImportSpec{"src/main.ts": specs},
		ImportsResolved: map[string][]string{"src/main.ts": importsResolved},
		Exports:         map[string][]string{},
		ExportSymbols:   map[string][]ExportSymbol{},
		Identifiers:     map[string]map[string]int{},
		UsageCounts:     map[string]map[string]int{},
		FunctionDecls:   map[string][]string{},
		VariableDecls:   map[string][]string{},
		FunctionLines:   map[string]map[string]int{},
		VariableLines:   map[string]map[string]int{},
		FeatureFlagRefs: map[string]int{},
		FeatureFlagHits: map[string][]FlagOccurrence{},
		DynamicIndicators: map[string][]string{},
	}

	findings := applyRules(cfg, collected)

	for _, f := range findings {
		if f.Kind == "unused_file" && f.File == "libs/helper.ts" {
			t.Errorf("libs/helper.ts was flagged as unused_file, but it is imported via alias @lib/helper")
		}
	}
}

func TestFindDeclLineNumber(t *testing.T) {
	// The identifier "count" appears first in a comment (line 1) and
	// as a usage (line 2), but the declaration is on line 3.
	// findDeclLineNumber must return line 3, not line 1.
	content := `// count is used for tracking
console.log(count)
const count = 42
let other = count + 1`

	tests := []struct {
		name     string
		varName  string
		wantLine int
	}{
		{
			name:     "finds const declaration, not comment or usage",
			varName:  "count",
			wantLine: 3,
		},
		{
			name:     "finds let declaration",
			varName:  "other",
			wantLine: 4,
		},
		{
			name:     "falls back to first occurrence when no declaration keyword",
			varName:  "console",
			wantLine: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findDeclLineNumber(content, tt.varName)
			if got != tt.wantLine {
				t.Errorf("findDeclLineNumber(%q) = %d, want %d", tt.varName, got, tt.wantLine)
			}
		})
	}
}
