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
	importSpecs       map[string][]ImportSpec
	importsResolved   map[string][]string
	exports           map[string][]string
	identifiers       map[string]map[string]int
	functionDecls     map[string][]string
	variableDecls     map[string][]string
	featureFlagRefs   map[string]int
	dynamicIndicators map[string][]string
	importRegexes     []*regexp.Regexp
	requireRegexes    []*regexp.Regexp
	importSpecRegexes importSpecRegexes
}

func NewCollector(cfg *config.Config) *Collector {
	return &Collector{
		cfg:               cfg,
		imports:           map[string][]string{},
		importSpecs:       map[string][]ImportSpec{},
		importsResolved:   map[string][]string{},
		exports:           map[string][]string{},
		identifiers:       map[string]map[string]int{},
		functionDecls:     map[string][]string{},
		variableDecls:     map[string][]string{},
		featureFlagRefs:   map[string]int{},
		dynamicIndicators: map[string][]string{},
		importRegexes: []*regexp.Regexp{
			regexp.MustCompile(`(?m)^\s*import\s+[^;]*?from\s+["']([^"']+)["']`),
			regexp.MustCompile(`(?m)^\s*import\s+["']([^"']+)["']`),
		},
		requireRegexes: []*regexp.Regexp{
			regexp.MustCompile(`(?m)\brequire\(\s*["']([^"']+)["']\s*\)`),
		},
		importSpecRegexes: buildImportSpecRegexes(),
	}
}

type Collected struct {
	Files             []scan.FileEntry
	Imports           map[string][]string
	ImportSpecs       map[string][]ImportSpec
	ImportsResolved   map[string][]string
	Exports           map[string][]string
	Identifiers       map[string]map[string]int
	FunctionDecls     map[string][]string
	VariableDecls     map[string][]string
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
		importSpecs := c.parseImportSpecs(content)
		for i := range importSpecs {
			importSpecs[i].Resolved = resolveImportSpec(entry.Rel, importSpecs[i], fileIndex)
		}
		c.imports[entry.Rel] = rawImports
		c.importSpecs[entry.Rel] = importSpecs
		c.importsResolved[entry.Rel] = resolveLocalImports(entry.Rel, importSpecs, fileIndex)
		c.exports[entry.Rel] = extractExports(content)
		c.identifiers[entry.Rel] = countIdentifiers(content)
		c.functionDecls[entry.Rel] = extractFunctionDecls(content)
		c.variableDecls[entry.Rel] = extractVariableDecls(content)
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
		ImportSpecs:       c.importSpecs,
		ImportsResolved:   c.importsResolved,
		Exports:           c.exports,
		Identifiers:       c.identifiers,
		FunctionDecls:     c.functionDecls,
		VariableDecls:     c.variableDecls,
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

func extractFunctionDecls(content string) []string {
	results := []string{}
	reFunc := regexp.MustCompile(`(?m)^\s*(?:export\s+)?function\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	for _, match := range reFunc.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			results = append(results, match[1])
		}
	}
	return results
}

func extractVariableDecls(content string) []string {
	results := []string{}
	reVar := regexp.MustCompile(`(?m)^\s*(?:export\s+)?(?:const|let|var)\s+([A-Za-z_][A-Za-z0-9_]*)`)
	for _, match := range reVar.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			results = append(results, match[1])
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

type ImportSpec struct {
	Source     string
	Resolved   string
	Names      []string
	Wildcard   bool
	SideEffect bool
}

type importSpecRegexes struct {
	defaultImport   *regexp.Regexp
	defaultNamed    *regexp.Regexp
	namedOnly       *regexp.Regexp
	namespaceImport *regexp.Regexp
	sideEffect      *regexp.Regexp
	requireDefault  *regexp.Regexp
	requireNamed    *regexp.Regexp
}

func buildImportSpecRegexes() importSpecRegexes {
	return importSpecRegexes{
		defaultImport:   regexp.MustCompile(`(?m)^\s*import\s+([A-Za-z_$][\w$]*)\s+from\s+["']([^"']+)["']`),
		defaultNamed:    regexp.MustCompile(`(?m)^\s*import\s+([A-Za-z_$][\w$]*)\s*,\s*\{([^}]+)\}\s+from\s+["']([^"']+)["']`),
		namedOnly:       regexp.MustCompile(`(?m)^\s*import\s*\{([^}]+)\}\s+from\s+["']([^"']+)["']`),
		namespaceImport: regexp.MustCompile(`(?m)^\s*import\s+\*\s+as\s+([A-Za-z_$][\w$]*)\s+from\s+["']([^"']+)["']`),
		sideEffect:      regexp.MustCompile(`(?m)^\s*import\s+["']([^"']+)["']`),
		requireDefault:  regexp.MustCompile(`(?m)^\s*(?:const|let|var)\s+([A-Za-z_$][\w$]*)\s*=\s*require\(\s*["']([^"']+)["']\s*\)`),
		requireNamed:    regexp.MustCompile(`(?m)^\s*(?:const|let|var)\s+\{([^}]+)\}\s*=\s*require\(\s*["']([^"']+)["']\s*\)`),
	}
}

func (c *Collector) parseImportSpecs(content string) []ImportSpec {
	specs := []ImportSpec{}
	re := c.importSpecRegexes

	for _, match := range re.defaultNamed.FindAllStringSubmatch(content, -1) {
		if len(match) > 3 {
			names := parseImportNames(match[2])
			names = append(names, match[1])
			specs = append(specs, ImportSpec{
				Source:   match[3],
				Names:    names,
				Wildcard: true,
			})
		}
	}

	for _, match := range re.namedOnly.FindAllStringSubmatch(content, -1) {
		if len(match) > 2 {
			specs = append(specs, ImportSpec{
				Source: match[2],
				Names:  parseImportNames(match[1]),
			})
		}
	}

	for _, match := range re.defaultImport.FindAllStringSubmatch(content, -1) {
		if len(match) > 2 {
			specs = append(specs, ImportSpec{
				Source:   match[2],
				Names:    []string{match[1]},
				Wildcard: true,
			})
		}
	}

	for _, match := range re.namespaceImport.FindAllStringSubmatch(content, -1) {
		if len(match) > 2 {
			specs = append(specs, ImportSpec{
				Source:   match[2],
				Names:    []string{match[1]},
				Wildcard: true,
			})
		}
	}

	for _, match := range re.requireNamed.FindAllStringSubmatch(content, -1) {
		if len(match) > 2 {
			specs = append(specs, ImportSpec{
				Source:   match[2],
				Names:    parseImportNames(match[1]),
				Wildcard: true,
			})
		}
	}

	for _, match := range re.requireDefault.FindAllStringSubmatch(content, -1) {
		if len(match) > 2 {
			specs = append(specs, ImportSpec{
				Source:   match[2],
				Names:    []string{match[1]},
				Wildcard: true,
			})
		}
	}

	for _, match := range re.sideEffect.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			specs = append(specs, ImportSpec{
				Source:     match[1],
				Wildcard:   true,
				SideEffect: true,
			})
		}
	}

	return specs
}

func parseImportNames(raw string) []string {
	parts := strings.Split(raw, ",")
	names := []string{}
	for _, part := range parts {
		clean := strings.TrimSpace(part)
		clean = strings.TrimPrefix(clean, "type ")
		if strings.Contains(clean, " as ") {
			clean = strings.TrimSpace(strings.Split(clean, " as ")[0])
		}
		if strings.Contains(clean, ":") {
			clean = strings.TrimSpace(strings.Split(clean, ":")[0])
		}
		if clean != "" {
			names = append(names, clean)
		}
	}
	return names
}

func resolveLocalImports(from string, specs []ImportSpec, index map[string]scan.FileEntry) []string {
	resolved := []string{}
	base := filepath.Dir(from)
	for _, spec := range specs {
		if !strings.HasPrefix(spec.Source, ".") {
			continue
		}
		candidate := filepath.ToSlash(filepath.Clean(filepath.Join(base, spec.Source)))
		if target, ok := resolveFile(candidate, index); ok {
			resolved = append(resolved, target)
		}
	}
	return resolved
}

func resolveImportSpec(from string, spec ImportSpec, index map[string]scan.FileEntry) string {
	if !strings.HasPrefix(spec.Source, ".") {
		return ""
	}
	base := filepath.Dir(from)
	candidate := filepath.ToSlash(filepath.Clean(filepath.Join(base, spec.Source)))
	if target, ok := resolveFile(candidate, index); ok {
		return target
	}
	return ""
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
