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
			resolved := spec.Resolved
			if resolved == "" {
				resolved = resolveImportedFile(file, spec.Source, data.Files)
			}
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
		}
	}
	findings := []rules.Finding{}
	for file, exports := range data.Exports {
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
			findings = append(findings, rules.Finding{
				ID:         "unused_export:" + file + ":" + symbol,
				Kind:       "unused_export",
				Confidence: confidence,
				File:       file,
				Line:       1,
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

	for file, symbols := range data.FunctionDecls {
		if !isRuleEnabled(cfg, "unused_function") {
			continue
		}
		counts := data.Identifiers[file]
		for _, symbol := range symbols {
			if counts[symbol] > 1 {
				continue
			}
			confidence := confidenceFor(cfg, "unused_function", "default", "likely_dead")
			if len(data.DynamicIndicators[file]) > 0 {
				confidence = confidenceFor(cfg, "unused_function", "if_dynamic_usage", "review")
			}
			findings = append(findings, rules.Finding{
				ID:         "unused_function:" + file + ":" + symbol,
				Kind:       "unused_function",
				Confidence: confidence,
				File:       file,
				Line:       1,
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
		for _, symbol := range symbols {
			if counts[symbol] > 1 {
				continue
			}
			confidence := confidenceFor(cfg, "unused_variable", "default", "safe")
			if len(data.DynamicIndicators[file]) > 0 {
				confidence = confidenceFor(cfg, "unused_variable", "if_dynamic_usage", "review")
			}
			findings = append(findings, rules.Finding{
				ID:         "unused_variable:" + file + ":" + symbol,
				Kind:       "unused_variable",
				Confidence: confidence,
				File:       file,
				Line:       1,
				Symbol:     symbol,
				Reason:     "variable declared but never referenced",
			})
		}
	}
	return findings
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
	for _, file := range cfg.Entrypoints.Files {
		set[filepath.ToSlash(file)] = true
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
		if match, _ := doublestar.Match(pattern, file); match {
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
