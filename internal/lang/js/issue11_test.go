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
