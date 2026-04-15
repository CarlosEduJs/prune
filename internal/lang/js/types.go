package js

import "regexp"

type ImportSpec struct {
	Source     string
	Resolved   string
	Confidence string
	Names      []string
	Wildcard   bool
	SideEffect bool
	IsReexport bool
}

type ExportSymbol struct {
	Name string
	Line int
}

type FlagOccurrence struct {
	Flag string
	Line int
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
		defaultNamed:    regexp.MustCompile(`(?m)^\s*import\s+([A-Za-z_$][\w$]*)\s*,\s*\{([^}]*)\}\s+from\s+["']([^"']+)["']`),
		namedOnly:       regexp.MustCompile(`(?m)^\s*import\s*\{([^}]*)\}\s+from\s+["']([^"']+)["']`),
		namespaceImport: regexp.MustCompile(`(?m)^\s*import\s+\*\s+as\s+([A-Za-z_$][\w$]*)\s+from\s+["']([^"']+)["']`),
		sideEffect:      regexp.MustCompile(`(?m)^\s*import\s+["']([^"']+)["']`),
		requireDefault:  regexp.MustCompile(`(?m)^\s*(?:const|let|var)\s+([A-Za-z_$][\w$]*)\s*=\s*require\(\s*["']([^"']+)["']\s*\)`),
		requireNamed:    regexp.MustCompile(`(?m)^\s*(?:const|let|var)\s+\{([^}]*)}\s*=\s*require\(\s*["']([^"']+)["']\s*\)`),
	}
}
