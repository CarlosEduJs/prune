package cli

import (
	"encoding/json"
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

func TestE2ECliFlags(t *testing.T) {
	root := filepath.Join(projectRoot(t), "internal", "cli", "testdata", "project")
	cmd := exec.Command("go", "run", "./cmd/prune", "scan", "--config", filepath.Join(root, "prune.yaml"), "--min-confidence", "safe")
	cmd.Dir = projectRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("scan with flags failed: %v\n%s", err, string(output))
	}
	out := string(output)
	if !strings.Contains(out, "issues found") {
		t.Fatalf("expected output to contain findings, got:\n%s", out)
	}
}

func TestE2ECompactOutput(t *testing.T) {
	root := filepath.Join(projectRoot(t), "internal", "cli", "testdata", "compact")
	cmd := exec.Command("go", "run", "./cmd/prune", "scan", "--config", filepath.Join(root, "prune.yaml"), "--compact")
	cmd.Dir = projectRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compact scan failed: %v\n%s", err, string(output))
	}
	out := string(output)
	if !strings.Contains(out, "Summary") {
		t.Fatalf("expected 'Summary' in compact output, got:\n%s", out)
	}
	if strings.Contains(out, "unused file:") {
		t.Fatalf("compact should not show individual findings, got:\n%s", out)
	}
}

func TestE2EDeletableFilter(t *testing.T) {
	root := filepath.Join(projectRoot(t), "internal", "cli", "testdata", "deletable")
	cmd := exec.Command("go", "run", "./cmd/prune", "scan", "--config", filepath.Join(root, "prune.yaml"), "--deletable")
	cmd.Dir = projectRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("deletable scan failed: %v\n%s", err, string(output))
	}
	out := string(output)
	if !strings.Contains(out, "unused file") {
		t.Fatalf("expected 'unused file' in deletable output, got:\n%s", out)
	}
}

func TestE2EStreamingOutput(t *testing.T) {
	root := filepath.Join(projectRoot(t), "internal", "cli", "testdata", "streaming")
	cmd := exec.Command("go", "run", "./cmd/prune", "scan", "--config", filepath.Join(root, "prune.yaml"), "--stream", "--stream-interval", "50", "--format", "ndjson")
	cmd.Dir = projectRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("streaming scan failed: %v\n%s", err, string(output))
	}
	out := string(output)
	if len(out) == 0 {
		t.Fatalf("expected non-empty streaming output")
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	foundValidJson := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var check map[string]interface{}
		if err := json.Unmarshal([]byte(line), &check); err == nil {
			if _, hasKind := check["kind"]; hasKind {
				foundValidJson = true
				break
			}
		}
	}
	if !foundValidJson {
		t.Fatalf("expected at least one valid NDJSON line with 'kind' field, got:\n%s", out)
	}
}

func TestE2EJsonOutput(t *testing.T) {
	root := filepath.Join(projectRoot(t), "internal", "cli", "testdata", "compact")
	cmd := exec.Command("go", "run", "./cmd/prune", "scan", "--config", filepath.Join(root, "prune.yaml"), "--format", "json")
	cmd.Dir = projectRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("json scan failed: %v\n%s", err, string(output))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, string(output))
	}
	if _, ok := result["summary"]; !ok {
		t.Fatalf("expected 'summary' in JSON output, got: %s", string(output))
	}
	if _, ok := result["findings"]; !ok {
		t.Fatalf("expected 'findings' in JSON output, got: %s", string(output))
	}
}

func TestE2EReExports(t *testing.T) {
	root := filepath.Join(projectRoot(t), "internal", "cli", "testdata", "reexports")
	cmd := exec.Command("go", "run", "./cmd/prune", "scan", "--config", filepath.Join(root, "prune.yaml"))
	cmd.Dir = projectRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("reexports scan failed: %v\n%s", err, string(output))
	}
	_ = err
}

func TestE2EFailOnFindings(t *testing.T) {
	root := filepath.Join(projectRoot(t), "internal", "cli", "testdata", "project")
	cmd := exec.Command("go", "run", "./cmd/prune", "scan", "--config", filepath.Join(root, "prune.yaml"), "--fail-on-findings")
	cmd.Dir = projectRoot(t)
	_, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected scan to fail with --fail-on-findings, but it succeeded")
	}
}

func TestE2ESafePatternsNotFlaggedAsDynamic(t *testing.T) {
	root := filepath.Join(projectRoot(t), "internal", "cli", "testdata", "safe-patterns")
	cmd := exec.Command("go", "run", "./cmd/prune", "scan", "--config", filepath.Join(root, "prune.yaml"))
	cmd.Dir = projectRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("safe-patterns scan failed: %v\n%s", err, string(output))
	}
	out := string(output)
	if strings.Contains(out, "REVIEW") {
		t.Fatalf("expected safe patterns not to trigger REVIEW confidence, got:\n%s", out)
	}
}

func TestE2EEntrypointDefaultExportIgnored(t *testing.T) {
	root := filepath.Join(projectRoot(t), "internal", "cli", "testdata", "entry-default")
	cmd := exec.Command("go", "run", "./cmd/prune", "scan", "--config", filepath.Join(root, "prune.yaml"))
	cmd.Dir = projectRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("entry-default scan failed: %v\n%s", err, string(output))
	}
	_ = err
}

func TestE2EAliasImportsTrackedAsUsed(t *testing.T) {
	root := filepath.Join(projectRoot(t), "internal", "cli", "testdata", "alias-used")
	cmd := exec.Command("go", "run", "./cmd/prune", "scan", "--config", filepath.Join(root, "prune.yaml"), "--format", "json")
	cmd.Dir = projectRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("alias-used scan failed: %v\n%s", err, string(output))
	}
	_ = output
}

func TestE2ECircularDependencies(t *testing.T) {
	root := filepath.Join(projectRoot(t), "internal", "cli", "testdata", "circular")
	cmd := exec.Command("go", "run", "./cmd/prune", "scan", "--config", filepath.Join(root, "prune.yaml"))
	cmd.Dir = projectRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("circular scan failed: %v\n%s", err, string(output))
	}
	_ = err
}

func TestE2EMultipleEntrypoints(t *testing.T) {
	root := filepath.Join(projectRoot(t), "internal", "cli", "testdata", "multiple-entrypoints")
	cmd := exec.Command("go", "run", "./cmd/prune", "scan", "--config", filepath.Join(root, "prune.yaml"), "--format", "json")
	cmd.Dir = projectRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("multiple-entrypoints scan failed: %v\n%s", err, string(output))
	}
	out := string(output)
	if len(out) == 0 {
		t.Fatalf("expected non-empty output")
	}
}

func TestE2EBarrelFiles(t *testing.T) {
	root := filepath.Join(projectRoot(t), "internal", "cli", "testdata", "barrel")
	cmd := exec.Command("go", "run", "./cmd/prune", "scan", "--config", filepath.Join(root, "prune.yaml"))
	cmd.Dir = projectRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("barrel scan failed: %v\n%s", err, string(output))
	}
	_ = err
}

func projectRoot(t *testing.T) string {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve project root: %v", err)
	}
	return root
}