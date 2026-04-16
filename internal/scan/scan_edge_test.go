package scan

import (
	"os"
	"testing"

	"prune/internal/config"
)

func TestCollectFilesDefaultPath(t *testing.T) {
	root := "testdata"
	if err := withWorkdir(root, func() error {
		cfg := &config.Config{}
		cfg.Scan.Include = []string{"**/*.js"}
		entries, err := Collect(cfg)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			t.Fatalf("expected entries when using default path")
		}
		return nil
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func withWorkdir(dir string, fn func() error) error {
	old, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(dir); err != nil {
		return err
	}
	defer func() { _ = os.Chdir(old) }()
	return fn()
}
