package js

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"prune/internal/config"
	"prune/internal/scan"
)

type Collector struct {
	cfg               *config.Config
	imports           map[string][]string
	importsResolved   map[string][]string
	exports           map[string][]string
	identifiers       map[string]map[string]int
	featureFlagRefs   map[string]int
	dynamicIndicators map[string][]string
	importRegexes     []*regexp.Regexp
	requireRegexes    []*regexp.Regexp
}

func NewCollector(cfg *config.Config) *Collector {
	return &Collector{
		cfg:               cfg,
		imports:           map[string][]string{},
		importsResolved:   map[string][]string{},
		exports:           map[string][]string{},
		identifiers:       map[string]map[string]int{},
		featureFlagRefs:   map[string]int{},
		dynamicIndicators: map[string][]string{},
		importRegexes: []*regexp.Regexp{
			regexp.MustCompile(`(?m)^\s*import\s+[^;]*?from\s+["']([^"']+)["']`),
			regexp.MustCompile(`(?m)^\s*import\s+["']([^"']+)["']`),
		},
		requireRegexes: []*regexp.Regexp{
			regexp.MustCompile(`(?m)\brequire\(\s*["']([^"']+)["']\s*\)`),
		},
	}
}

type Collected struct {
	Files             []scan.FileEntry
	Imports           map[string][]string
	ImportsResolved   map[string][]string
	Exports           map[string][]string
	Identifiers       map[string]map[string]int
	FeatureFlagRefs   map[string]int
	DynamicIndicators map[string][]string
}

func (c *Collector) Collect(entries []scan.FileEntry) (*Collected, error) {
	if c == nil || c.cfg == nil {
		return nil, errors.New("collector config is required")
	}

	fileIndex := map[string]scan.FileEntry{}
	for _, entry := range entries {
		fileIndex[entry.Rel] = entry
	}

	patterns := c.cfg.FeatureFlags.Patterns
	flagRegexes := compileRegexes(patterns)

	for _, entry := range entries {
		content, err := readFile(entry.Path)
		if err != nil {
			return nil, err
		}

		rawImports := c.extractImports(content)
		c.imports[entry.Rel] = rawImports
		c.importsResolved[entry.Rel] = resolveLocalImports(entry.Rel, rawImports, fileIndex)
		c.exports[entry.Rel] = extractExports(content)
		c.identifiers[entry.Rel] = countIdentifiers(content)
		c.dynamicIndicators[entry.Rel] = detectDynamic(content, c.cfg)

		for _, re := range flagRegexes {
			for _, match := range re.FindAllString(content, -1) {
				c.featureFlagRefs[match]++
			}
		}
	}

	return &Collected{
		Files:             entries,
		Imports:           c.imports,
		ImportsResolved:   c.importsResolved,
		Exports:           c.exports,
		Identifiers:       c.identifiers,
		FeatureFlagRefs:   c.featureFlagRefs,
		DynamicIndicators: c.dynamicIndicators,
	}, nil
}

func (c *Collector) extractImports(content string) []string {
	imports := []string{}
	for _, re := range c.importRegexes {
		for _, match := range re.FindAllStringSubmatch(content, -1) {
			if len(match) > 1 {
				imports = append(imports, match[1])
			}
		}
	}
	for _, re := range c.requireRegexes {
		for _, match := range re.FindAllStringSubmatch(content, -1) {
			if len(match) > 1 {
				imports = append(imports, match[1])
			}
		}
	}
	return imports
}

func extractExports(content string) []string {
	results := []string{}
	reExport := regexp.MustCompile(`(?m)^\s*export\s+(?:const|let|var|function|class)\s+([A-Za-z_][A-Za-z0-9_]*)`)
	reNamed := regexp.MustCompile(`(?m)^\s*export\s*\{([^}]+)\}`)
	for _, match := range reExport.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			results = append(results, match[1])
		}
	}
	for _, match := range reNamed.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			parts := strings.Split(match[1], ",")
			for _, part := range parts {
				name := strings.TrimSpace(strings.Split(part, " as ")[0])
				if name != "" {
					results = append(results, name)
				}
			}
		}
	}
	return results
}

func countIdentifiers(content string) map[string]int {
	counts := map[string]int{}
	reIdent := regexp.MustCompile(`\b[A-Za-z_][A-Za-z0-9_]*\b`)
	for _, match := range reIdent.FindAllString(content, -1) {
		counts[match]++
	}
	return counts
}

func detectDynamic(content string, cfg *config.Config) []string {
	indicators := []string{}
	patterns := []string{"eval", "Function", "require", "import("}
	if rule, ok := cfg.Rules["suspicious_dynamic_usage"]; ok && len(rule.Patterns) > 0 {
		patterns = rule.Patterns
	}
	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			indicators = append(indicators, pattern)
		}
	}
	return indicators
}

func compileRegexes(patterns []string) []*regexp.Regexp {
	regexes := []*regexp.Regexp{}
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err == nil {
			regexes = append(regexes, re)
		}
	}
	return regexes
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func resolveLocalImports(from string, imports []string, index map[string]scan.FileEntry) []string {
	resolved := []string{}
	base := filepath.Dir(from)
	for _, imp := range imports {
		if !strings.HasPrefix(imp, ".") {
			continue
		}
		candidate := filepath.ToSlash(filepath.Clean(filepath.Join(base, imp)))
		if target, ok := resolveFile(candidate, index); ok {
			resolved = append(resolved, target)
		}
	}
	return resolved
}

func resolveFile(path string, index map[string]scan.FileEntry) (string, bool) {
	if _, ok := index[path]; ok {
		return path, true
	}
	if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".jsx") || strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".tsx") {
		return "", false
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
