package scan

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"prune/internal/config"
)

type FileEntry struct {
	Path string
	Rel  string
}

func collectFiles(cfg *config.Config) ([]FileEntry, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}

	paths := cfg.Scan.Paths
	if len(paths) == 0 {
		paths = []string{"."}
	}

	include := cfg.Scan.Include
	exclude := cfg.Scan.Exclude

	entries := []FileEntry{}
	for _, root := range paths {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			return nil, err
		}

		err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			rel, err := filepath.Rel(absRoot, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)

			if len(include) > 0 && !matchesAny(rel, include) {
				return nil
			}
			if matchesAny(rel, exclude) {
				return nil
			}

			entries = append(entries, FileEntry{Path: path, Rel: rel})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return entries, nil
}

func matchesAny(path string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := doublestar.Match(pattern, path)
		if err == nil && matched {
			return true
		}
	}
	return false
}
