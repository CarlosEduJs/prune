package js

import (
	"testing"

	"prune/internal/config"
	"prune/internal/scan"
)

func TestUniqueStrings(t *testing.T) {
	values := []string{"a", "b", "a", "", "c"}
	unique := uniqueStrings(values)
	if len(unique) != 3 {
		t.Fatalf("expected 3 values, got %d", len(unique))
	}
}

func TestMergeExportNames(t *testing.T) {
	base := []string{"a"}
	exports := []ExportSymbol{{Name: "b"}, {Name: ""}}
	merged := mergeExportNames(base, exports)
	if len(merged) != 2 {
		t.Fatalf("expected 2 names, got %d", len(merged))
	}
}

func TestFlagHitExists(t *testing.T) {
	hits := []FlagOccurrence{{Flag: "flags.A"}}
	if !flagHitExists(hits, "flags.A") {
		t.Fatalf("expected hit exists")
	}
	if flagHitExists(hits, "flags.B") {
		t.Fatalf("expected hit missing")
	}
}

func TestGetHighRiskPatterns(t *testing.T) {
	t.Run("default patterns", func(t *testing.T) {
		cfg := &config.Config{}
		patterns := getHighRiskPatterns(cfg, "unused_function")
		if len(patterns) != 3 {
			t.Fatalf("expected 3 default patterns, got %d: %v", len(patterns), patterns)
		}
	})

	t.Run("custom patterns from config", func(t *testing.T) {
		cfg := &config.Config{
			Rules: map[string]config.RuleConfig{
				"unused_function": {
					HighRiskPatterns: []string{"customEval", "customFunc"},
				},
			},
		}
		patterns := getHighRiskPatterns(cfg, "unused_function")
		if len(patterns) != 2 {
			t.Fatalf("expected 2 custom patterns, got %d: %v", len(patterns), patterns)
		}
		if patterns[0] != "customEval" {
			t.Fatalf("expected customEval, got %s", patterns[0])
		}
	})

	t.Run("rule key not configured", func(t *testing.T) {
		cfg := &config.Config{
			Rules: map[string]config.RuleConfig{
				"unused_variable": {
					HighRiskPatterns: []string{"test"},
				},
			},
		}
		patterns := getHighRiskPatterns(cfg, "unused_function")
		if len(patterns) != 3 {
			t.Fatalf("expected 3 defaults when unused_function not set, got %d", len(patterns))
		}
	})
}

func TestGetSafePatterns(t *testing.T) {
	t.Run("default patterns", func(t *testing.T) {
		cfg := &config.Config{}
		patterns := getSafePatterns(cfg, "unused_function")
		if len(patterns) != 11 {
			t.Fatalf("expected 11 default patterns, got %d: %v", len(patterns), patterns)
		}
		found := false
		for _, p := range patterns {
			if p == "Math" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected Math in defaults, got %v", patterns)
		}
	})

	t.Run("custom patterns from config", func(t *testing.T) {
		cfg := &config.Config{
			Rules: map[string]config.RuleConfig{
				"unused_function": {
					SafePatterns: []string{"customPrefix"},
				},
			},
		}
		patterns := getSafePatterns(cfg, "unused_function")
		if len(patterns) != 1 {
			t.Fatalf("expected 1 custom pattern, got %d: %v", len(patterns), patterns)
		}
		if patterns[0] != "customPrefix" {
			t.Fatalf("expected customPrefix, got %s", patterns[0])
		}
	})
}

func TestHasHighRiskDynamic(t *testing.T) {
	cfg := &config.Config{}

	t.Run("no indicators", func(t *testing.T) {
		if hasHighRiskDynamic([]string{}, cfg, "unused_function") {
			t.Fatal("expected false for empty indicators")
		}
	})

	t.Run("no high-risk indicators", func(t *testing.T) {
		if hasHighRiskDynamic([]string{"console.log", "Math.random"}, cfg, "unused_function") {
			t.Fatal("expected false for safe patterns")
		}
	})

	t.Run("eval detected", func(t *testing.T) {
		if !hasHighRiskDynamic([]string{"eval", "console.log"}, cfg, "unused_function") {
			t.Fatal("expected true for eval")
		}
	})

	t.Run("Function detected", func(t *testing.T) {
		if !hasHighRiskDynamic([]string{"new Function()"}, cfg, "unused_function") {
			t.Fatal("expected true for Function")
		}
	})

	t.Run("import() detected", func(t *testing.T) {
		if !hasHighRiskDynamic([]string{"import('./module')"}, cfg, "unused_function") {
			t.Fatal("expected true for import()")
		}
	})
}

func TestClassifyDynamicIndicators(t *testing.T) {
	cfg := &config.Config{}

	t.Run("high-risk indicator", func(t *testing.T) {
		indicators := []string{"eval"}
		result := classifyDynamicIndicators(indicators, cfg, "unused_function")
		if len(result) != 1 {
			t.Fatalf("expected 1 result, got %d", len(result))
		}
		if !result[0].IsHighRisk {
			t.Fatal("expected IsHighRisk to be true for eval")
		}
	})

	t.Run("safe indicator", func(t *testing.T) {
		indicators := []string{"console.log"}
		result := classifyDynamicIndicators(indicators, cfg, "unused_function")
		if len(result) != 1 {
			t.Fatalf("expected 1 result, got %d", len(result))
		}
		if result[0].IsHighRisk {
			t.Fatal("expected IsHighRisk to be false for console.log")
		}
		if result[0].SafeMatch != "console" {
			t.Fatalf("expected SafeMatch to be 'console', got %s", result[0].SafeMatch)
		}
	})

	t.Run("medium-risk indicator", func(t *testing.T) {
		indicators := []string{"someUnknownPattern"}
		result := classifyDynamicIndicators(indicators, cfg, "unused_function")
		if len(result) != 1 {
			t.Fatalf("expected 1 result, got %d", len(result))
		}
		if result[0].IsHighRisk {
			t.Fatal("expected IsHighRisk to be false for unknown pattern")
		}
		if result[0].SafeMatch != "" {
			t.Fatalf("expected SafeMatch to be empty for unknown pattern, got %s", result[0].SafeMatch)
		}
	})

	t.Run("multiple indicators", func(t *testing.T) {
		indicators := []string{"eval", "console.log", "Math.max", "unknown"}
		result := classifyDynamicIndicators(indicators, cfg, "unused_function")
		if len(result) != 4 {
			t.Fatalf("expected 4 results, got %d", len(result))
		}

		if !result[0].IsHighRisk {
			t.Fatal("expected first to be high-risk (eval)")
		}

		if result[1].SafeMatch != "console" {
			t.Fatalf("expected second SafeMatch to be 'console', got %s", result[1].SafeMatch)
		}

		if result[2].SafeMatch != "Math" {
			t.Fatalf("expected third SafeMatch to be 'Math', got %s", result[2].SafeMatch)
		}

		if result[3].IsHighRisk || result[3].SafeMatch != "" {
			t.Fatal("expected fourth to be medium-risk")
		}
	})
}

func TestCollectedReleaseUnused(t *testing.T) {
	c := &Collected{
		FeatureFlagRefs: map[string]int{"TODO": 5},
		FeatureFlagHits: map[string][]FlagOccurrence{
			"main.ts": {{Flag: "TODO", Line: 10}},
		},
	}

	c.ReleaseUnused()

	if c.FeatureFlagRefs != nil {
		t.Error("expected FeatureFlagRefs to be nil")
	}
	if c.FeatureFlagHits != nil {
		t.Error("expected FeatureFlagHits to be nil")
	}
}

func TestCollectedReset(t *testing.T) {
	c := &Collected{
		Files: []scan.FileEntry{{Rel: "a.ts"}},
		Exports: map[string][]string{"a.ts": {"foo"}},
	}

	c.Reset()

	if c.Files != nil {
		t.Error("expected Files to be nil after Reset")
	}
	if c.Exports != nil {
		t.Error("expected Exports to be nil after Reset")
	}
}