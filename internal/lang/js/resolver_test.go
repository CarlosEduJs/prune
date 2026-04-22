package js

import (
	"testing"

	"prune/internal/config"
	"prune/internal/scan"
)

func TestResolveRelative(t *testing.T) {
	cfg := &config.Config{}
	fileIndex := map[string]scan.FileEntry{
		"src/local.ts":        {Path: "testdata/src/local.ts", Rel: "src/local.ts"},
		"src/utils/helper.ts": {Path: "testdata/src/utils/helper.ts", Rel: "src/utils/helper.ts"},
		"src/main.ts":         {Path: "testdata/src/main.ts", Rel: "src/main.ts"},
	}

	r := NewResolver(cfg, fileIndex)

	tests := []struct {
		name     string
		source   string
		fromFile string
		wantType ImportType
		wantRes  string
		wantConf string
	}{
		{"relative dot", "./local", "src/main.ts", ImportTypeRelative, "src/local.ts", "safe"},
		{"relative dotdot", "../other", "src/utils/helper.ts", ImportTypeRelative, "", "review"},
		{"relative subdir", "./utils/helper", "src/main.ts", ImportTypeRelative, "src/utils/helper.ts", "safe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.Resolve(tt.source, tt.fromFile)
			if got.Type != tt.wantType {
				t.Errorf("type = %v, want %v", got.Type, tt.wantType)
			}
			if got.Resolved != tt.wantRes {
				t.Errorf("resolved = %q, want %q", got.Resolved, tt.wantRes)
			}
			if got.Confidence != tt.wantConf {
				t.Errorf("confidence = %q, want %q", got.Confidence, tt.wantConf)
			}
		})
	}
}

func TestResolveAlias(t *testing.T) {
	cfg := &config.Config{
		TsConfig: config.TsConfig{
			Enabled: true,
			BaseURL: ".",
			Paths: map[string][]string{
				"@/*": {"src/*"},
			},
		},
	}
	fileIndex := map[string]scan.FileEntry{
		"src/main.ts":         {Path: "testdata/src/main.ts", Rel: "src/main.ts"},
		"src/utils/helper.ts": {Path: "testdata/src/utils/helper.ts", Rel: "src/utils/helper.ts"},
	}

	r := NewResolver(cfg, fileIndex)

	got := r.Resolve("@/utils/helper", "src/main.ts")
	if got.Type != ImportTypeAlias {
		t.Errorf("type = %v, want %v", got.Type, ImportTypeAlias)
	}
	if got.Resolved != "utils/helper" {
		t.Errorf("resolved = %q, want %q", got.Resolved, "utils/helper")
	}
	if got.Confidence != "safe" {
		t.Errorf("confidence = %q, want %q", got.Confidence, "safe")
	}
}

func TestResolveAliasNotFound(t *testing.T) {
	cfg := &config.Config{
		TsConfig: config.TsConfig{
			Enabled: true,
			BaseURL: ".",
			Paths: map[string][]string{
				"@/*": {"src/*"},
			},
		},
	}
	fileIndex := map[string]scan.FileEntry{
		"src/main.ts": {Path: "testdata/src/main.ts", Rel: "src/main.ts"},
	}

	r := NewResolver(cfg, fileIndex)

	got := r.Resolve("@/nonexistent", "src/main.ts")
	if got.Type != ImportTypeAlias {
		t.Errorf("type = %v, want %v", got.Type, ImportTypeAlias)
	}
	if got.Resolved != "nonexistent" {
		t.Errorf("resolved = %q, want %q", got.Resolved, "nonexistent")
	}
	if got.Confidence != "safe" {
		t.Errorf("confidence = %q, want %q", got.Confidence, "safe")
	}
}

func TestResolveAliasAtSlashUsesBaseURL(t *testing.T) {
	cfg := &config.Config{
		TsConfig: config.TsConfig{
			Enabled: true,
			BaseURL: "src",
		},
	}
	fileIndex := map[string]scan.FileEntry{
		"src/main.ts":         {Path: "testdata/src/main.ts", Rel: "src/main.ts"},
		"src/utils/helper.ts": {Path: "testdata/src/utils/helper.ts", Rel: "src/utils/helper.ts"},
	}

	r := NewResolver(cfg, fileIndex)

	got := r.Resolve("@/utils/helper", "src/main.ts")
	if got.Type != ImportTypeAlias {
		t.Errorf("type = %v, want %v", got.Type, ImportTypeAlias)
	}
	if got.Resolved != "src/utils/helper.ts" {
		t.Errorf("resolved = %q, want %q", got.Resolved, "src/utils/helper.ts")
	}
	if got.Confidence != "safe" {
		t.Errorf("confidence = %q, want %q", got.Confidence, "safe")
	}
}

func TestFindBestAliasPrefersExactOverWildcard(t *testing.T) {
	cfg := &config.Config{
		TsConfig: config.TsConfig{
			Enabled: true,
			BaseURL: ".",
			Paths: map[string][]string{
				"@lib":   {"src/lib/index.ts"},
				"@lib/*": {"src/lib/*"},
			},
		},
	}

	r := NewResolver(cfg, nil)

	if got := r.findBestAlias("@lib"); got != "@lib" {
		t.Fatalf("findBestAlias = %q, want %q", got, "@lib")
	}
}

func TestFindBestAliasWildcardBoundary(t *testing.T) {
	cfg := &config.Config{
		TsConfig: config.TsConfig{
			Enabled: true,
			BaseURL: ".",
			Paths: map[string][]string{
				"@scope/*": {"src/scope/*"},
			},
		},
	}

	r := NewResolver(cfg, nil)

	if got := r.findBestAlias("@scopeX/pkg"); got != "" {
		t.Fatalf("findBestAlias = %q, want empty", got)
	}
}

func TestResolveExternal(t *testing.T) {
	cfg := &config.Config{}
	fileIndex := map[string]scan.FileEntry{
		"src/main.ts": {Path: "testdata/src/main.ts", Rel: "src/main.ts"},
	}

	r := NewResolver(cfg, fileIndex)

	tests := []struct {
		name     string
		source   string
		wantType ImportType
		wantConf string
	}{
		{"lodash", "lodash", ImportTypeExternal, "safe"},
		{"react", "react", ImportTypeExternal, "safe"},
		{"@scope/pkg", "@scope/pkg", ImportTypeExternal, "safe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.Resolve(tt.source, "src/main.ts")
			if got.Type != tt.wantType {
				t.Errorf("type = %v, want %v", got.Type, tt.wantType)
			}
			if got.Confidence != tt.wantConf {
				t.Errorf("confidence = %q, want %q", got.Confidence, tt.wantConf)
			}
		})
	}
}

func TestClassify(t *testing.T) {
	cfg := &config.Config{
		TsConfig: config.TsConfig{
			Enabled: true,
			BaseURL: ".",
			Paths: map[string][]string{
				"@/*": {"src/*"},
			},
		},
	}
	r := NewResolver(cfg, nil)

	tests := []struct {
		source   string
		wantType ImportType
	}{
		{"./local", ImportTypeRelative},
		{"../parent", ImportTypeRelative},
		{"@/utils", ImportTypeAlias},
		{"lodash", ImportTypeExternal},
		{"react", ImportTypeExternal},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			if got := r.Classify(tt.source); got != tt.wantType {
				t.Errorf("classify(%q) = %v, want %v", tt.source, got, tt.wantType)
			}
		})
	}
}
