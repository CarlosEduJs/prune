package scan

import (
	"context"

	"prune/internal/config"
)

func Collect(cfg *config.Config) ([]FileEntry, error) {
	return collectFiles(context.Background(), cfg)
}

func CollectWithContext(ctx context.Context, cfg *config.Config) ([]FileEntry, error) {
	return collectFiles(ctx, cfg)
}
