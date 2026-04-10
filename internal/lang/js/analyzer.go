package js

import (
	"errors"

	"prune/internal/config"
	"prune/internal/rules"
	"prune/internal/scan"
)

func Analyze(cfg *config.Config) ([]rules.Finding, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}

	entries, err := scan.Collect(cfg)
	if err != nil {
		return nil, err
	}

	collector := NewCollector(cfg)
	collected, err := collector.Collect(entries)
	if err != nil {
		return nil, err
	}

	findings := applyRules(cfg, collected)
	return findings, nil
}
