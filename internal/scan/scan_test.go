package scan

import (
	"testing"

	"prune/internal/config"
)

func TestCollectFilesIncludeExclude(t *testing.T) {
	root := "testdata"
	cfg := &config.Config{}
	cfg.Scan.Paths = []string{root}
	cfg.Scan.Include = []string{"**/*.js", "**/*.ts"}
	cfg.Scan.Exclude = []string{"node_modules/**"}

	entries, err := Collect(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	paths := map[string]bool{}
	for _, entry := range entries {
		paths[entry.Rel] = true
	}

	if !paths["src/a.js"] {
		t.Fatalf("expected src/a.js")
	}
	if !paths["src/b.ts"] {
		t.Fatalf("expected src/b.ts")
	}
	if paths["node_modules/ignored.js"] {
		t.Fatalf("expected ignored file excluded")
	}
}
