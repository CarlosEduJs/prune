package scan

import "prune/internal/config"

func Collect(cfg *config.Config) ([]FileEntry, error) {
	return collectFiles(cfg)
}
