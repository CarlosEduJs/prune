package js

import (
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"prune/internal/config"
	"prune/internal/rules"
	"prune/internal/scan"
)

func applyRules(cfg *config.Config, data *Collected) []rules.Finding {
	findings := []rules.Finding{}
	if data == nil {
		return findings
	}

	findings = append(findings, ruleUnusedFiles(cfg, data)...)
	findings = append(findings, ruleUnusedExports(cfg, data)...)
	findings = append(findings, ruleUnusedSymbols(cfg, data)...)
	findings = append(findings, ruleDeadFeatureFlags(cfg, data)...)
	findings = append(findings, ruleSuspiciousDynamic(cfg, data)...)
	return findings
}

func ruleUnusedFiles(cfg *config.Config, data *Collected) []rules.Finding {
	if !isRuleEnabled(cfg, "unused_file") {
		return nil
	}

	entrypoints := buildEntrypointSet(cfg)
	imported := map[string]bool{}
	for _, imports := range data.ImportsResolved {
		for _, imp := range imports {
			imported[imp] = true
		}
	}

	findings := []rules.Finding{}
	for _, entry := range data.Files {
		if entrypoints[entry.Rel] {
			continue
		}
		if isEntrypointPattern(cfg, entry.Rel) {
			continue
		}
		if imported[entry.Rel] {
			continue
		}
		confidence := confidenceFor(cfg, "unused_file", "default", "safe")
		findings = append(findings, rules.Finding{
			ID:         "unused_file:" + entry.Rel,
			Kind:       "unused_file",
			Confidence: confidence,
			File:       entry.Rel,
			Line:       1,
			Symbol:     filepath.Base(entry.Rel),
			Reason:     "file is not referenced by any import",
		})
	}
	return findings
}

func ruleUnusedExports(cfg *config.Config, data *Collected) []rules.Finding {
	if !isRuleEnabled(cfg, "unused_export") {
		return nil
	}

	usedExports := map[string]map[string]bool{}
	for file, specs := range data.ImportSpecs {
		_ = file
		for _, spec := range specs {
			if !strings.HasPrefix(spec.Source, ".") {
				continue
			}
			resolved := resolveImportTarget(file, spec, data.Files)
			if resolved == "" {
				continue
			}
			if usedExports[resolved] == nil {
				usedExports[resolved] = map[string]bool{}
			}
			if spec.Wildcard || spec.SideEffect {
				usedExports[resolved]["*"] = true
				continue
			}
			for _, name := range spec.Names {
				usedExports[resolved][name] = true
			}
			if spec.IsReexport {
				if len(spec.Names) == 0 {
					usedExports[resolved]["*"] = true
				}
			}
		}
	}
	findings := []rules.Finding{}
	for file, exports := range data.Exports {
		lineMap := exportLineMap(data.ExportSymbols[file])
		for _, symbol := range exports {
			if usedExports[file] != nil {
				if usedExports[file]["*"] || usedExports[file][symbol] {
					continue
				}
			}
			confidence := confidenceFor(cfg, "unused_export", "if_not_imported", "safe")
			if isEntrypoint(cfg, file) {
				confidence = confidenceFor(cfg, "unused_export", "if_entrypoint", "review")
			}
			line := 1
			if value, ok := lineMap[symbol]; ok {
				line = value
			}
			findings = append(findings, rules.Finding{
				ID:         "unused_export:" + file + ":" + symbol,
				Kind:       "unused_export",
				Confidence: confidence,
				File:       file,
				Line:       line,
				Symbol:     symbol,
				Reason:     "exported symbol is never imported",
			})
		}
	}

	return findings
}

func ruleUnusedSymbols(cfg *config.Config, data *Collected) []rules.Finding {
	findings := []rules.Finding{}
	if !isRuleEnabled(cfg, "unused_function") && !isRuleEnabled(cfg, "unused_variable") {
		return findings
	}

	importedByFile := buildImportedSymbols(data)

	for file, symbols := range data.FunctionDecls {
		if !isRuleEnabled(cfg, "unused_function") {
			continue
		}
		counts := data.Identifiers[file]
		usage := data.UsageCounts[file]
		lines := data.FunctionLines[file]
		exported := exportedSymbolSet(data.Exports[file])
		for _, symbol := range symbols {
			if usage[symbol] > 0 {
				continue
			}
			if counts[symbol] > 1 {
				continue
			}
			if exported[symbol] && (isImportedSymbol(importedByFile[file], symbol) || isDefaultExportUsed(file, data)) {
				continue
			}
			confidence := confidenceFor(cfg, "unused_function", "default", "likely_dead")
			if len(data.DynamicIndicators[file]) > 0 {
				confidence = confidenceFor(cfg, "unused_function", "if_dynamic_usage", "review")
			}
			line := 1
			if value, ok := lines[symbol]; ok && value > 0 {
				line = value
			}
			findings = append(findings, rules.Finding{
				ID:         "unused_function:" + file + ":" + symbol,
				Kind:       "unused_function",
				Confidence: confidence,
				File:       file,
				Line:       line,
				Symbol:     symbol,
				Reason:     "function declared but never referenced",
			})
		}
	}

	for file, symbols := range data.VariableDecls {
		if !isRuleEnabled(cfg, "unused_variable") {
			continue
		}
		counts := data.Identifiers[file]
		usage := data.UsageCounts[file]
		lines := data.VariableLines[file]
		exported := exportedSymbolSet(data.Exports[file])
		for _, symbol := range symbols {
			if usage[symbol] > 0 {
				continue
			}
			if counts[symbol] > 1 {
				continue
			}
			if exported[symbol] && (isImportedSymbol(importedByFile[file], symbol) || isDefaultExportUsed(file, data)) {
				continue
			}
			confidence := confidenceFor(cfg, "unused_variable", "default", "safe")
			if len(data.DynamicIndicators[file]) > 0 {
				confidence = confidenceFor(cfg, "unused_variable", "if_dynamic_usage", "review")
			}
			line := 1
			if value, ok := lines[symbol]; ok && value > 0 {
				line = value
			}
			findings = append(findings, rules.Finding{
				ID:         "unused_variable:" + file + ":" + symbol,
				Kind:       "unused_variable",
				Confidence: confidence,
				File:       file,
				Line:       line,
				Symbol:     symbol,
				Reason:     "variable declared but never referenced",
			})
		}
	}
	return findings
}

func buildImportedSymbols(data *Collected) map[string]map[string]bool {
	imported := map[string]map[string]bool{}
	if data == nil {
		return imported
	}
	for file, specs := range data.ImportSpecs {
		for _, spec := range specs {
			if !strings.HasPrefix(spec.Source, ".") {
				continue
			}
			resolved := resolveImportTarget(file, spec, data.Files)
			if resolved == "" {
				continue
			}
			if imported[resolved] == nil {
				imported[resolved] = map[string]bool{}
			}
			if spec.Wildcard {
				imported[resolved]["*"] = true
				continue
			}
			for _, name := range spec.Names {
				imported[resolved][name] = true
			}
		}
	}
	return imported
}

func exportedSymbolSet(exports []string) map[string]bool {
	set := map[string]bool{}
	for _, name := range exports {
		if name != "" {
			set[name] = true
		}
	}
	return set
}

func isImportedSymbol(imported map[string]bool, symbol string) bool {
	if imported == nil {
		return false
	}
	if imported["*"] {
		return true
	}
	return imported[symbol]
}

func isDefaultExportUsed(file string, data *Collected) bool {
	if data == nil {
		return false
	}
	if len(data.ExportSymbols[file]) == 0 {
		return false
	}
	if !exportedSymbolSet(data.Exports[file])["default"] {
		return false
	}
	return isImportedSymbol(buildImportedSymbols(data)[file], "default")
}

func ruleDeadFeatureFlags(cfg *config.Config, data *Collected) []rules.Finding {
	if !isRuleEnabled(cfg, "dead_feature_flag") {
		return nil
	}

	findings := []rules.Finding{}
	if len(data.FeatureFlagRefs) == 0 && len(cfg.FeatureFlags.Patterns) > 0 {
		confidence := confidenceFor(cfg, "dead_feature_flag", "if_never_referenced", "safe")
		findings = append(findings, rules.Finding{
			ID:         "dead_feature_flag:patterns",
			Kind:       "dead_feature_flag",
			Confidence: confidence,
			File:       "",
			Line:       0,
			Symbol:     strings.Join(cfg.FeatureFlags.Patterns, ","),
			Reason:     "feature flag patterns never referenced",
		})
	}

	for file, hits := range data.FeatureFlagHits {
		for _, hit := range hits {
			if data.FeatureFlagRefs[hit.Flag] > 0 {
				continue
			}
			confidence := confidenceFor(cfg, "dead_feature_flag", "if_never_referenced", "safe")
			line := hit.Line
			if line == 0 {
				line = 1
			}
			findings = append(findings, rules.Finding{
				ID:         "dead_feature_flag:" + file + ":" + hit.Flag,
				Kind:       "dead_feature_flag",
				Confidence: confidence,
				File:       file,
				Line:       line,
				Symbol:     hit.Flag,
				Reason:     "feature flag never referenced",
			})
		}
	}
	return findings
}

func ruleSuspiciousDynamic(cfg *config.Config, data *Collected) []rules.Finding {
	if !isRuleEnabled(cfg, "suspicious_dynamic_usage") {
		return nil
	}
	findings := []rules.Finding{}
	for file, indicators := range data.DynamicIndicators {
		for _, indicator := range indicators {
			findings = append(findings, rules.Finding{
				ID:         "suspicious_dynamic_usage:" + file + ":" + indicator,
				Kind:       "suspicious_dynamic_usage",
				Confidence: confidenceFor(cfg, "suspicious_dynamic_usage", "default", "review"),
				File:       file,
				Line:       1,
				Symbol:     indicator,
				Reason:     "dynamic usage may hide references",
			})
		}
	}
	return findings
}

func isRuleEnabled(cfg *config.Config, key string) bool {
	if cfg == nil {
		return false
	}
	rule, ok := cfg.Rules[key]
	if !ok {
		return true
	}
	return rule.Enabled
}

func confidenceFor(cfg *config.Config, rule string, key string, fallback string) string {
	if cfg == nil {
		return fallback
	}
	if ruleCfg, ok := cfg.Rules[rule]; ok {
		if value, ok := ruleCfg.Confidence[key]; ok && value != "" {
			return value
		}
	}
	return fallback
}

func buildEntrypointSet(cfg *config.Config) map[string]bool {
	set := map[string]bool{}
	if cfg == nil {
		return set
	}

	scanPaths := cfg.Scan.Paths
	if len(scanPaths) == 0 {
		scanPaths = []string{"."}
	}

	for _, file := range cfg.Entrypoints.Files {
		entry := filepath.ToSlash(file)
		set[entry] = true

		absEntry, err := filepath.Abs(entry)
		if err == nil {
			absEntry = filepath.ToSlash(absEntry)
			set[absEntry] = true
		}

		for _, scanPath := range scanPaths {
			absScanPath, err := filepath.Abs(scanPath)
			if err != nil {
				continue
			}
			absScanPath = filepath.ToSlash(absScanPath)

			relPath, err := filepath.Rel(absScanPath, absEntry)
			if err == nil {
				relPath = filepath.ToSlash(relPath)
				set[relPath] = true
			}

			if !strings.Contains(entry, "/") {
				entryBase := filepath.Base(entry)
				scanBase := filepath.Base(scanPath)
				if entryBase == scanBase {
					set[entry] = true
				}
			}
		}
	}
	for _, pattern := range cfg.Entrypoints.Patterns {
		set[filepath.ToSlash(pattern)] = true
	}
	return set
}

func isEntrypointPattern(cfg *config.Config, file string) bool {
	if cfg == nil {
		return false
	}
	for _, pattern := range cfg.Entrypoints.Patterns {
		if match, err := doublestar.Match(pattern, file); err == nil && match {
			return true
		}
	}
	return false
}

func isEntrypoint(cfg *config.Config, file string) bool {
	if cfg == nil {
		return false
	}
	for _, entry := range cfg.Entrypoints.Files {
		if filepath.ToSlash(entry) == file {
			return true
		}
	}
	for _, pattern := range cfg.Entrypoints.Patterns {
		if match, _ := filepath.Match(pattern, file); match {
			return true
		}
	}
	return false
}

func resolveImportedFile(from string, source string, files []scan.FileEntry) string {
	index := map[string]bool{}
	for _, entry := range files {
		index[entry.Rel] = true
	}

	base := filepath.Dir(from)
	path := filepath.ToSlash(filepath.Clean(filepath.Join(base, source)))
	if index[path] {
		return path
	}
	if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".jsx") || strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".tsx") {
		return ""
	}

	extensions := []string{".ts", ".tsx", ".js", ".jsx"}
	for _, ext := range extensions {
		candidate := path + ext
		if index[candidate] {
			return candidate
		}
	}
	for _, ext := range extensions {
		candidate := filepath.ToSlash(filepath.Join(path, "index"+ext))
		if index[candidate] {
			return candidate
		}
	}
	return ""
}

func resolveImportTarget(from string, spec ImportSpec, files []scan.FileEntry) string {
	if spec.Resolved != "" {
		return spec.Resolved
	}
	return resolveImportedFile(from, spec.Source, files)
}

func exportLineMap(symbols []ExportSymbol) map[string]int {
	lineMap := map[string]int{}
	for _, symbol := range symbols {
		if symbol.Name == "" || symbol.Line <= 0 {
			continue
		}
		if _, exists := lineMap[symbol.Name]; !exists {
			lineMap[symbol.Name] = symbol.Line
		}
	}
	return lineMap
}
