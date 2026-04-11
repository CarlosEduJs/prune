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
}

type StreamConfig struct {
	Enabled    bool `yaml:"enabled"`
	IntervalMs int  `yaml:"interval_ms"`
}

type RuleConfig struct {
	Enabled    bool              `yaml:"enabled"`
	Confidence map[string]string `yaml:"confidence"`
	Patterns   []string          `yaml:"patterns"`
	Attributes map[string]string `yaml:"attributes"`
	Limit      int               `yaml:"limit"`
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

	return &cfg, nil
}
