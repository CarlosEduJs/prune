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
	if !strings.Contains(out, "unused_export") {
		t.Fatalf("expected unused_export in output")
	}
	if !strings.Contains(out, "unused_function") {
		t.Fatalf("expected unused_function in output")
	}
	if !strings.Contains(out, "unused_variable") {
		t.Fatalf("expected unused_variable in output")
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
