package js

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"prune/internal/config"
	"prune/internal/scan"
)

type Collector struct {
	cfg               *config.Config
	resolver          *Resolver
	imports           map[string][]string
	importSpecs       map[string][]ImportSpec
	importsResolved   map[string][]string
	exports           map[string][]string
	exportSymbols     map[string][]ExportSymbol
	identifiers       map[string]map[string]int
	usageCounts       map[string]map[string]int
	functionDecls     map[string][]string
	variableDecls     map[string][]string
	functionLines     map[string]map[string]int
	variableLines     map[string]map[string]int
	featureFlagRefs   map[string]int
	featureFlagHits   map[string][]FlagOccurrence
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
		exportSymbols:     map[string][]ExportSymbol{},
		identifiers:       map[string]map[string]int{},
		usageCounts:       map[string]map[string]int{},
		functionDecls:     map[string][]string{},
		variableDecls:     map[string][]string{},
		functionLines:     map[string]map[string]int{},
		variableLines:     map[string]map[string]int{},
		featureFlagRefs:   map[string]int{},
		featureFlagHits:   map[string][]FlagOccurrence{},
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
	ExportSymbols     map[string][]ExportSymbol
	Identifiers       map[string]map[string]int
	UsageCounts       map[string]map[string]int
	FunctionDecls     map[string][]string
	VariableDecls     map[string][]string
	FunctionLines     map[string]map[string]int
	VariableLines     map[string]map[string]int
	FeatureFlagRefs   map[string]int
	FeatureFlagHits   map[string][]FlagOccurrence
	DynamicIndicators map[string][]string
}

func (c *Collector) Collect(ctx context.Context, entries []scan.FileEntry) (*Collected, error) {
	if c == nil || c.cfg == nil {
		return nil, errors.New("collector config is required")
	}

	fileIndex := map[string]scan.FileEntry{}
	for _, entry := range entries {
		fileIndex[entry.Rel] = entry
	}

	c.resolver = NewResolver(c.cfg, fileIndex)

	patterns := c.cfg.FeatureFlags.Patterns
	flagRegexes := compileRegexes(patterns)

	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		content, err := readFile(entry.Path)
		if err != nil {
			return nil, fmt.Errorf("reading file %q: %w", entry.Path, err)
		}
		contentBytes := []byte(content)

		rawImports := c.extractImports(content)
		importSpecs := c.parseImportSpecs(content)
		for i := range importSpecs {
			if importSpecs[i].SideEffect {
				continue
			}
			resolved := c.resolver.Resolve(importSpecs[i].Source, entry.Rel)
			importSpecs[i].Resolved = resolved.Resolved
			importSpecs[i].Confidence = resolved.Confidence
		}
		c.imports[entry.Rel] = rawImports
		c.importSpecs[entry.Rel] = importSpecs
		c.importsResolved[entry.Rel] = resolveLocalImports(entry.Rel, importSpecs, fileIndex)
		c.exports[entry.Rel] = extractExports(content)
		c.identifiers[entry.Rel] = countIdentifiers(content)
		c.usageCounts[entry.Rel] = map[string]int{}
		c.functionDecls[entry.Rel] = extractFunctionDecls(content)
		c.variableDecls[entry.Rel] = extractVariableDecls(content)
		c.functionLines[entry.Rel] = map[string]int{}
		c.variableLines[entry.Rel] = map[string]int{}
		if ast, err := collectASTData(ctx, entry.Rel, contentBytes, c.cfg.FeatureFlags.Patterns); err == nil {
			c.identifiers[entry.Rel] = ast.Identifiers
			c.usageCounts[entry.Rel] = ast.UsageCounts
			c.functionDecls[entry.Rel] = ast.FunctionDecls
			c.variableDecls[entry.Rel] = ast.VariableDecls
			c.functionLines[entry.Rel] = ast.FunctionLines
			c.variableLines[entry.Rel] = ast.VariableLines
			c.importSpecs[entry.Rel] = mergeImportSpecs(importSpecs, ast.ImportSpecs)
			c.exports[entry.Rel] = mergeExportNames(c.exports[entry.Rel], ast.ExportSymbols)
			c.exports[entry.Rel] = uniqueStrings(c.exports[entry.Rel])
			c.importsResolved[entry.Rel] = resolveLocalImports(entry.Rel, c.importSpecs[entry.Rel], fileIndex)
			c.exportSymbols[entry.Rel] = ast.ExportSymbols
			c.featureFlagHits[entry.Rel] = mergeFlagHits(c.featureFlagHits[entry.Rel], ast.FlagHits)
		}
		c.dynamicIndicators[entry.Rel] = detectDynamic(content, c.cfg)

		for _, re := range flagRegexes {
			for _, match := range re.FindAllString(content, -1) {
				c.featureFlagRefs[match]++
				if !flagHitExists(c.featureFlagHits[entry.Rel], match) {
					c.featureFlagHits[entry.Rel] = append(c.featureFlagHits[entry.Rel], FlagOccurrence{
						Flag: match,
						Line: 0,
					})
				}
			}
		}
	}

	return &Collected{
		Files:             entries,
		Imports:           c.imports,
		ImportSpecs:       c.importSpecs,
		ImportsResolved:   c.importsResolved,
		Exports:           c.exports,
		ExportSymbols:     c.exportSymbols,
		Identifiers:       c.identifiers,
		UsageCounts:       c.usageCounts,
		FunctionDecls:     c.functionDecls,
		VariableDecls:     c.variableDecls,
		FunctionLines:     c.functionLines,
		VariableLines:     c.variableLines,
		FeatureFlagRefs:   c.featureFlagRefs,
		FeatureFlagHits:   c.featureFlagHits,
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

type DynamicIndicator struct {
	Pattern    string
	IsHighRisk bool
	SafeMatch  string
}

func classifyDynamicIndicators(indicators []string, cfg *config.Config, ruleKey string) []DynamicIndicator {
	result := []DynamicIndicator{}

	highRiskPatterns := getHighRiskPatterns(cfg, ruleKey)
	safePatterns := getSafePatterns(cfg, ruleKey)

	for _, indicator := range indicators {
		di := DynamicIndicator{Pattern: indicator}

		for _, risk := range highRiskPatterns {
			if strings.Contains(indicator, risk) {
				di.IsHighRisk = true
				break
			}
		}

		if !di.IsHighRisk {
			for _, safe := range safePatterns {
				if strings.Contains(indicator, safe) {
					di.SafeMatch = safe
					break
				}
			}
		}

		result = append(result, di)
	}

	return result
}

func getHighRiskPatterns(cfg *config.Config, ruleKey string) []string {
	defaults := []string{"eval", "Function", "import("}

	if cfg == nil || cfg.Rules == nil {
		return defaults
	}

	rule, ok := cfg.Rules[ruleKey]
	if !ok {
		return defaults
	}

	if len(rule.HighRiskPatterns) > 0 {
		return rule.HighRiskPatterns
	}

	return defaults
}

func getSafePatterns(cfg *config.Config, ruleKey string) []string {
	defaults := []string{
		"window", "document", "Math", "JSON",
		"Object", "Array", "process", "Buffer",
		"setTimeout", "setInterval", "console",
	}

	if cfg == nil || cfg.Rules == nil {
		return defaults
	}

	rule, ok := cfg.Rules[ruleKey]
	if !ok {
		return defaults
	}

	if len(rule.SafePatterns) > 0 {
		return rule.SafePatterns
	}

	return defaults
}

func hasHighRiskDynamic(indicators []string, cfg *config.Config, ruleKey string) bool {
	highRisk := getHighRiskPatterns(cfg, ruleKey)

	for _, ind := range indicators {
		for _, risk := range highRisk {
			if strings.Contains(ind, risk) {
				return true
			}
		}
	}

	suspiciousPatterns := getSuspiciousDynamicPatterns(cfg)
	for _, ind := range indicators {
		for _, pattern := range suspiciousPatterns {
			if strings.Contains(ind, pattern) {
				return true
			}
		}
	}

	return false
}

func getSuspiciousDynamicPatterns(cfg *config.Config) []string {
	defaults := []string{"eval", "Function", "require", "import("}

	if cfg == nil || cfg.Rules == nil {
		return defaults
	}

	rule, ok := cfg.Rules["suspicious_dynamic_usage"]
	if !ok {
		return defaults
	}

	if len(rule.Patterns) > 0 {
		return rule.Patterns
	}

	return defaults
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

func (c *Collector) parseImportSpecs(content string) []ImportSpec {
	specs := []ImportSpec{}
	re := c.importSpecRegexes

	for _, match := range re.defaultNamed.FindAllStringSubmatch(content, -1) {
		if len(match) > 3 {
			names := parseImportNames(match[2])
			names = append(names, match[1])
			specs = append(specs, ImportSpec{
				Source:   match[3],
				Names:    append(names, "default"),
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
				Names:    []string{match[1], "default"},
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
				Names:    []string{match[1], "default"},
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

func mergeImportSpecs(base []ImportSpec, overrides []ImportSpec) []ImportSpec {
	if len(overrides) == 0 {
		return base
	}
	combined := make([]ImportSpec, 0, len(base)+len(overrides))
	combined = append(combined, base...)
	combined = append(combined, overrides...)
	return combined
}

func mergeExportNames(base []string, exports []ExportSymbol) []string {
	combined := make([]string, 0, len(base)+len(exports))
	combined = append(combined, base...)
	for _, symbol := range exports {
		if symbol.Name != "" {
			combined = append(combined, symbol.Name)
		}
	}
	return combined
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	unique := []string{}
	for _, value := range values {
		if value == "" {
			continue
		}
		if seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}

func flagHitExists(hits []FlagOccurrence, flag string) bool {
	for _, hit := range hits {
		if hit.Flag == flag {
			return true
		}
	}
	return false
}

func mergeFlagHits(base []FlagOccurrence, hits []FlagOccurrence) []FlagOccurrence {
	combined := make([]FlagOccurrence, 0, len(base)+len(hits))
	combined = append(combined, base...)
	for _, hit := range hits {
		if hit.Flag == "" {
			continue
		}
		if flagHitExists(combined, hit.Flag) {
			continue
		}
		combined = append(combined, hit)
	}
	return combined
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
