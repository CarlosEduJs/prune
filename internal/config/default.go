package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

func WriteDefault(path string) error {
	cfg := defaultConfig()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func defaultConfig() *Config {
	cfg := &Config{Version: 1}
	cfg.Project.Name = "prune"
	cfg.Project.Language = "js-ts"
	cfg.Scan.Paths = []string{"."}
	cfg.Scan.Include = []string{"**/*.js", "**/*.jsx", "**/*.ts", "**/*.tsx"}
	cfg.Scan.Exclude = []string{"node_modules/**", "dist/**", "build/**", ".next/**", "out/**", "coverage/**"}
	cfg.Entrypoints.Files = []string{"src/index.ts", "src/main.tsx"}
	cfg.Entrypoints.Patterns = []string{"src/pages/**", "src/routes/**"}
	cfg.Rules = map[string]RuleConfig{
		"unused_function": {
			Enabled: true,
			Confidence: map[string]string{
				"default":          "likely_dead",
				"if_dynamic_usage": "review",
			},
		},
		"unused_variable": {
			Enabled: true,
			Confidence: map[string]string{
				"default":          "safe",
				"if_exported":      "likely_dead",
				"if_dynamic_usage": "review",
			},
		},
		"unused_export": {
			Enabled: true,
			Confidence: map[string]string{
				"if_not_imported": "safe",
				"if_entrypoint":   "review",
			},
		},
		"unused_file": {
			Enabled: true,
			Confidence: map[string]string{
				"default":       "safe",
				"if_entrypoint": "review",
			},
		},
		"dead_feature_flag": {
			Enabled: true,
			Confidence: map[string]string{
				"if_never_referenced":  "safe",
				"if_dynamic_reference": "review",
			},
		},
		"suspicious_dynamic_usage": {
			Enabled: true,
			Patterns: []string{
				"eval",
				"Function",
				"require",
				"import(",
			},
			Confidence: map[string]string{
				"default": "review",
			},
		},
	}
	cfg.FeatureFlags.Patterns = []string{"flags\\.[A-Z0-9_]+", "featureFlags\\.[A-Za-z0-9_]+"}
	cfg.Report.Format = "table"
	cfg.Report.MinConfidence = "safe"
	cfg.Report.IncludeEvidence = true
	return cfg
}
