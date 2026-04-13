package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2EScanOutputsFindings(t *testing.T) {
	root := filepath.Join(projectRoot(t), "internal", "cli", "testdata", "project")
	cmd := exec.Command("go", "run", "./cmd/prune", "scan", "--config", filepath.Join(root, "prune.yaml"))
	cmd.Dir = projectRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("scan failed: %v\n%s", err, string(output))
	}

	out := string(output)
	// unused.ts is detected as an unused_file, so individual findings are deduplicated.
	if !strings.Contains(out, "unused file") {
		t.Fatalf("expected 'unused file' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "unused function") {
		t.Fatalf("expected 'unused function' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "unused export") {
		t.Fatalf("expected 'unused export' in output, got:\n%s", out)
	}
}

func TestE2EInitCreatesConfig(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "prune.yaml")
	cmd := exec.Command("go", "run", "./cmd/prune", "init", "--out", path)
	cmd.Dir = projectRoot(t)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\n%s", err, string(output))
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected prune.yaml to exist: %v", err)
	}
}

func projectRoot(t *testing.T) string {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve project root: %v", err)
	}
	return root
}
