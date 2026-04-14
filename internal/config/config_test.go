package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigValid(t *testing.T) {
	content := []byte("version: 1\nproject:\n  name: prune\n  language: js-ts\n")
	dir := t.TempDir()
	path := filepath.Join(dir, "prune.yaml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != 1 || cfg.Project.Language != "js-ts" {
		t.Fatalf("unexpected config values")
	}
}

func TestLoadConfigInvalid(t *testing.T) {
	content := []byte("project:\n  name: prune\n")
	dir := t.TempDir()
	path := filepath.Join(dir, "prune.yaml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for missing version")
	}
}

func TestWriteDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prune.yaml")
	if err := WriteDefault(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != 2 {
		t.Fatalf("expected version 2")
	}
	if cfg.Project.Language != "js-ts" {
		t.Fatalf("expected js-ts language")
	}
}
