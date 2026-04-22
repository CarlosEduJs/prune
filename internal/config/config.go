package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version int `yaml:"version"`
	Project struct {
		Name     string `yaml:"name"`
		Language string `yaml:"language"`
	} `yaml:"project"`
	Scan struct {
		Paths   []string     `yaml:"paths"`
		Include []string     `yaml:"include"`
		Exclude []string     `yaml:"exclude"`
		Stream  StreamConfig `yaml:"stream"`
	} `yaml:"scan"`
	Entrypoints struct {
		Files    []string `yaml:"files"`
		Patterns []string `yaml:"patterns"`
	} `yaml:"entrypoints"`
	Rules        map[string]RuleConfig `yaml:"rules"`
	FeatureFlags struct {
		Patterns []string `yaml:"patterns"`
	} `yaml:"feature_flags"`
	Report struct {
		Format          string `yaml:"format"`
		MinConfidence   string `yaml:"min_confidence"`
		IncludeEvidence bool   `yaml:"include_evidence"`
	} `yaml:"report"`
	TsConfig TsConfig `yaml:"ts_config"`
}

type StreamConfig struct {
	Enabled    bool `yaml:"enabled"`
	IntervalMs int  `yaml:"interval_ms"`
	BatchSize  int `yaml:"batch_size"`
}

type TsConfig struct {
	Enabled bool                `yaml:"enabled"`
	BaseURL string              `yaml:"baseUrl"`
	Paths   map[string][]string `yaml:"paths"`
}

type RuleConfig struct {
	Enabled          bool              `json:"enabled" yaml:"enabled"`
	Confidence       map[string]string `json:"confidence" yaml:"confidence"`
	Patterns         []string          `json:"patterns" yaml:"patterns"`
	Attributes       map[string]string `json:"attributes" yaml:"attributes"`
	Limit            int               `json:"limit" yaml:"limit"`
	SafePatterns     []string          `json:"safe_patterns" yaml:"safe_patterns"`
	HighRiskPatterns []string          `json:"high_risk_patterns" yaml:"high_risk_patterns"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if cfg.Version == 0 {
		return nil, errors.New("invalid config: version is required")
	}

	if cfg.Scan.Stream.Enabled {
		if cfg.Scan.Stream.IntervalMs <= 0 {
			return nil, errors.New("invalid config: stream.interval_ms must be positive")
		}
		if cfg.Scan.Stream.BatchSize <= 0 {
			return nil, errors.New("invalid config: stream.batch_size must be positive")
		}
	}

	return &cfg, nil
}
