package js

import (
	"context"
	"errors"
	"time"

	"prune/internal/config"
	"prune/internal/rules"
	"prune/internal/scan"
)

type StreamHandler func([]rules.Finding) error

func AnalyzeStreaming(ctx context.Context, cfg *config.Config, handler StreamHandler) ([]rules.Finding, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}

	stream := cfg.Scan.Stream

	entries, err := scan.CollectWithContext(ctx, cfg)
	if err != nil {
		return nil, err
	}

	collector := NewCollector(cfg)
	allFindings := []rules.Finding{}

	entriesCh := make(chan []scan.FileEntry, 1)
	errCh := make(chan error, 1)

	go func() {
		defer close(entriesCh)
		batch := []scan.FileEntry{}
		lastEmit := time.Now()
		interval := time.Duration(stream.IntervalMs) * time.Millisecond

		for _, entry := range entries {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
			}

			batch = append(batch, entry)

			if len(batch) >= 50 || time.Since(lastEmit) >= interval {
				entriesCh <- batch
				batch = []scan.FileEntry{}
				lastEmit = time.Now()
			}
		}

		if len(batch) > 0 {
			entriesCh <- batch
		}
	}()

	processed := 0
	for batch := range entriesCh {
		select {
		case err := <-errCh:
			return nil, err
		default:
		}

		collected, err := collector.Collect(ctx, batch)
		if err != nil {
			return nil, err
		}

		findings := applyRules(cfg, collected)
		allFindings = append(allFindings, findings...)
		processed += len(batch)

		if stream.Enabled && stream.IntervalMs > 0 && handler != nil {
			if err := handler(findings); err != nil {
				return nil, err
			}
		}
	}

	return allFindings, nil
}
