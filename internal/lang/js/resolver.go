package js

import (
	"path/filepath"
	"strings"

	"prune/internal/config"
	"prune/internal/scan"
)

type ImportType int

const (
	ImportTypeUnknown ImportType = iota
	ImportTypeRelative
	ImportTypeAlias
	ImportTypeExternal
)

type ResolvedImport struct {
	Type       ImportType
	Original   string
	Resolved   string
	Confidence string
}

type Resolver struct {
	cfg        *config.Config
	fileIndex  map[string]scan.FileEntry
	baseURL    string
	aliasPaths map[string][]string
}

func NewResolver(cfg *config.Config, fileIndex map[string]scan.FileEntry) *Resolver {
	r := &Resolver{
		cfg:        cfg,
		fileIndex:  fileIndex,
		baseURL:    cfg.TsConfig.BaseURL,
		aliasPaths: cfg.TsConfig.Paths,
	}
	if r.baseURL == "" {
		r.baseURL = "."
	}
	return r
}

func (r *Resolver) Resolve(source, fromFile string) ResolvedImport {
	importType := r.Classify(source)

	switch importType {
	case ImportTypeRelative:
		return r.resolveRelative(source, fromFile)
	case ImportTypeAlias:
		return r.resolveAlias(source, fromFile)
	case ImportTypeExternal:
		return ResolvedImport{
			Type:       ImportTypeExternal,
			Original:   source,
			Resolved:   "",
			Confidence: "safe",
		}
	default:
		return ResolvedImport{
			Type:       ImportTypeUnknown,
			Original:   source,
			Resolved:   "",
			Confidence: "review",
		}
	}
}

func (r *Resolver) Classify(source string) ImportType {
	if strings.HasPrefix(source, "./") || strings.HasPrefix(source, "../") {
		return ImportTypeRelative
	}

	if strings.HasPrefix(source, "@/") {
		return ImportTypeAlias
	}

	if r.findBestAlias(source) != "" {
		return ImportTypeAlias
	}

	if r.isExternal(source) {
		return ImportTypeExternal
	}

	return ImportTypeRelative
}

func (r *Resolver) resolveRelative(source, fromFile string) ResolvedImport {
	base := filepath.Dir(fromFile)
	candidate := filepath.ToSlash(filepath.Clean(filepath.Join(base, source)))

	if target, ok := r.resolveFile(candidate); ok {
		return ResolvedImport{
			Type:       ImportTypeRelative,
			Original:   source,
			Resolved:   target,
			Confidence: "safe",
		}
	}

	return ResolvedImport{
		Type:       ImportTypeRelative,
		Original:   source,
		Resolved:   "",
		Confidence: "review",
	}
}

func (r *Resolver) resolveAlias(source, fromFile string) ResolvedImport {
	if strings.HasPrefix(source, "@/") {
		baseURL := r.baseURL
		if baseURL == "." {
			baseURL = ""
		}
		resolvedPath := filepath.ToSlash(filepath.Clean(filepath.Join(baseURL, strings.TrimPrefix(source, "@/"))))
		if target, ok := r.resolveFile(resolvedPath); ok {
			return ResolvedImport{
				Type:       ImportTypeAlias,
				Original:   source,
				Resolved:   target,
				Confidence: "safe",
			}
		}
		return ResolvedImport{
			Type:       ImportTypeAlias,
			Original:   source,
			Resolved:   resolvedPath,
			Confidence: "safe",
		}
	}

	if r.aliasPaths == nil {
		return ResolvedImport{
			Type:       ImportTypeAlias,
			Original:   source,
			Resolved:   "",
			Confidence: "review",
		}
	}

	bestAlias := r.findBestAlias(source)
	if bestAlias == "" {
		return ResolvedImport{
			Type:       ImportTypeAlias,
			Original:   source,
			Resolved:   "",
			Confidence: "review",
		}
	}

	targets := r.aliasPaths[bestAlias]
	if len(targets) == 0 {
		return ResolvedImport{
			Type:       ImportTypeAlias,
			Original:   source,
			Resolved:   "",
			Confidence: "review",
		}
	}

	mapping := targets[0]
	suffix := ""
	if strings.HasSuffix(bestAlias, "/*") {
		prefix := strings.TrimSuffix(bestAlias, "/*")
		suffix = strings.TrimPrefix(source, prefix)
		suffix = strings.TrimPrefix(suffix, "/")
	}

	mappingBase := strings.TrimSuffix(mapping, "/*")
	baseURL := r.baseURL
	if baseURL == "." {
		baseURL = ""
	}
	resolvedPath := filepath.ToSlash(filepath.Clean(filepath.Join(baseURL, mappingBase, suffix)))

	if target, ok := r.resolveFile(resolvedPath); ok {
		return ResolvedImport{
			Type:       ImportTypeAlias,
			Original:   source,
			Resolved:   target,
			Confidence: "safe",
		}
	}

	return ResolvedImport{
		Type:       ImportTypeAlias,
		Original:   source,
		Resolved:   "",
		Confidence: "review",
	}
}

func (r *Resolver) resolveFile(path string) (string, bool) {
	if _, ok := r.fileIndex[path]; ok {
		return path, true
	}

	extensions := []string{".ts", ".tsx", ".js", ".jsx"}
	for _, ext := range extensions {
		candidate := path + ext
		if _, ok := r.fileIndex[candidate]; ok {
			return candidate, true
		}
	}
	for _, ext := range extensions {
		candidate := filepath.ToSlash(filepath.Join(path, "index"+ext))
		if _, ok := r.fileIndex[candidate]; ok {
			return candidate, true
		}
	}
	return "", false
}

// findBestAlias returns the alias key with the longest matching prefix for the
// given import source. This implements TypeScript's "longest prefix match" rule
// for path mappings (e.g. @scope/core/* wins over @scope/* for "@scope/core/util").
func (r *Resolver) findBestAlias(source string) string {
	for alias := range r.aliasPaths {
		if strings.HasSuffix(alias, "/*") {
			continue
		}
		if source == alias {
			return alias
		}
	}

	bestAlias := ""
	bestLen := 0
	for alias := range r.aliasPaths {
		if !strings.HasSuffix(alias, "/*") {
			continue
		}
		prefix := strings.TrimSuffix(alias, "/*")
		if strings.HasPrefix(source, prefix+"/") {
			if len(prefix) > bestLen {
				bestAlias = alias
				bestLen = len(prefix)
			}
		}
	}
	return bestAlias
}

func (r *Resolver) isExternal(source string) bool {
	if strings.HasPrefix(source, "./") || strings.HasPrefix(source, "../") || strings.HasPrefix(source, "@/") {
		return false
	}
	parts := strings.Split(source, "/")
	if len(parts) == 0 {
		return false
	}
	first := parts[0]
	if strings.HasPrefix(first, "@") && !strings.HasPrefix(source, "@/") {
		return true
	}
	if strings.HasPrefix(first, ".") {
		return false
	}
	return true
}

func ResolveFile(path string, index map[string]scan.FileEntry) (string, bool) {
	if _, ok := index[path]; ok {
		return path, true
	}

	extensions := []string{".ts", ".tsx", ".js", ".jsx"}
	for _, ext := range extensions {
		candidate := path + ext
		if _, ok := index[candidate]; ok {
			return candidate, true
		}
	}
	for _, ext := range extensions {
		candidate := filepath.ToSlash(filepath.Join(path, "index"+ext))
		if _, ok := index[candidate]; ok {
			return candidate, true
		}
	}
	return "", false
}
