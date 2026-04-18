package js

import (
	"context"
	"errors"

	"prune/internal/config"
	"prune/internal/rules"
	"prune/internal/scan"
)

func Analyze(ctx context.Context, cfg *config.Config) ([]rules.Finding, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}

	entries, err := scan.CollectWithContext(ctx, cfg)
	if err != nil {
		return nil, err
	}

	collector := NewCollector(cfg)
	collected, err := collector.Collect(ctx, entries)
	if err != nil {
		return nil, err
	}

	findings := applyRules(cfg, collected)

	collected.ReleaseUnused()

	return findings, nil
}
