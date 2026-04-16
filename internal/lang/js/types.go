// Package js provides JavaScript/TypeScript analysis capabilities.
package js

import "regexp"

// ImportSpec represents a single import statement with its resolution details.
type ImportSpec struct {
	Source     string
	Resolved   string
	Confidence string
	Names      []string
	Wildcard   bool
	SideEffect bool
	IsReexport bool
}

// ExportSymbol represents a named export with its line number.
type ExportSymbol struct {
	Name string
	Line int
}

// FlagOccurrence represents a detected flag (like TODO or FIXME) with its line.
type FlagOccurrence struct {
	Flag string
	Line int
}

// importSpecRegexes holds compiled regular expressions for parsing import statements.
type importSpecRegexes struct {
	defaultImport   *regexp.Regexp
	defaultNamed    *regexp.Regexp
	namedOnly       *regexp.Regexp
	namespaceImport *regexp.Regexp
	sideEffect      *regexp.Regexp
	requireDefault  *regexp.Regexp
	requireNamed    *regexp.Regexp
}

// buildImportSpecRegexes returns a compiled set of regex patterns for import parsing.
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
