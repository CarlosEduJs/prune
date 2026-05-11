package js

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"prune/internal/config"
	"prune/internal/rules"
)

func TestAnalyzeStreamingCancelsOnHandlerError(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}

	cfg := &config.Config{Version: 1}
	cfg.Scan.Paths = []string{filepath.Join(root, "internal", "cli", "testdata", "streaming", "src")}
	cfg.Scan.Include = []string{"**/*.ts"}
	cfg.Scan.Stream.Enabled = true
	cfg.Scan.Stream.IntervalMs = 1
	cfg.Scan.Stream.BatchSize = 1
	cfg.Entrypoints.Files = []string{
		filepath.ToSlash(cfg.Scan.Paths[0]) + "/index.ts",
	}

	start := time.Now()
	want := errors.New("handler failed")
	_, err = AnalyzeStreaming(context.Background(), cfg, func(batch []rules.Finding) error {
		return want
	})
	if err == nil {
		t.Fatalf("expected handler error")
	}
	if !errors.Is(err, want) {
		t.Fatalf("expected handler error, got %v", err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("expected AnalyzeStreaming to return promptly")
	}
}

func TestAnalyzeStreamingHonorsContextCancel(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}

	cfg := &config.Config{Version: 1}
	cfg.Scan.Paths = []string{filepath.Join(root, "internal", "cli", "testdata", "streaming", "src")}
	cfg.Scan.Include = []string{"**/*.ts"}
	cfg.Scan.Stream.Enabled = true
	cfg.Scan.Stream.IntervalMs = 1
	cfg.Scan.Stream.BatchSize = 1
	cfg.Entrypoints.Files = []string{
		filepath.ToSlash(cfg.Scan.Paths[0]) + "/index.ts",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = AnalyzeStreaming(ctx, cfg, func(batch []rules.Finding) error {
		return nil
	})
	if err == nil {
		t.Fatalf("expected context cancellation error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
