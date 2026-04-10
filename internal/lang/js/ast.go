package js

import (
	"path/filepath"
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

type astResult struct {
	Identifiers   map[string]int
	UsageCounts   map[string]int
	FunctionDecls []string
	VariableDecls []string
	ImportSpecs   []ImportSpec
	ExportSymbols []ExportSymbol
	FlagHits      []FlagOccurrence
}

func collectASTData(path string, content []byte, flagPatterns []string) (*astResult, bool) {
	lang := languageForPath(path)
	if lang == nil {
		return nil, false
	}

	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	defer parser.Close()

	tree, err := parser.ParseCtx(nil, nil, content)
	if err != nil || tree == nil {
		return nil, false
	}
	root := tree.RootNode()
	if root == nil || root.HasError() {
		return nil, false
	}

	result := &astResult{
		Identifiers:   map[string]int{},
		UsageCounts:   map[string]int{},
		FunctionDecls: []string{},
		VariableDecls: []string{},
		ImportSpecs:   []ImportSpec{},
		ExportSymbols: []ExportSymbol{},
		FlagHits:      []FlagOccurrence{},
	}

	collectIdentifiers(root, content, result)
	collectUsageCounts(root, content, result)
	collectFunctionDecls(root, content, result)
	collectVariableDecls(root, content, result)
	collectImportSpecs(root, content, result)
	collectExportSymbols(root, content, result)
	collectFlagHits(root, content, result, flagPatterns)

	return result, true
}

func shouldUseJSAST(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".js" || ext == ".jsx"
}

func languageForPath(path string) *sitter.Language {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".js", ".jsx":
		return javascript.GetLanguage()
	case ".ts":
		return typescript.GetLanguage()
	case ".tsx":
		return tsx.GetLanguage()
	default:
		return nil
	}
}

func collectIdentifiers(root *sitter.Node, content []byte, result *astResult) {
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	for {
		node := cursor.CurrentNode()
		if node.Type() == "identifier" {
			name := nodeContent(node, content)
			if name != "" {
				result.Identifiers[name]++
			}
		}

		if cursor.GoToFirstChild() {
			continue
		}
		for !cursor.GoToNextSibling() {
			if !cursor.GoToParent() {
				return
			}
		}
	}
}

func collectUsageCounts(root *sitter.Node, content []byte, result *astResult) {
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	for {
		node := cursor.CurrentNode()
		if node.Type() == "identifier" {
			if isDeclarationIdentifier(node) {
				if cursor.GoToFirstChild() {
					continue
				}
			} else {
				name := nodeContent(node, content)
				if name != "" {
					result.UsageCounts[name]++
				}
			}
		}

		if cursor.GoToFirstChild() {
			continue
		}
		for !cursor.GoToNextSibling() {
			if !cursor.GoToParent() {
				return
			}
		}
	}
}

func isDeclarationIdentifier(node *sitter.Node) bool {
	if node == nil {
		return false
	}
	parent := node.Parent()
	if parent == nil {
		return false
	}
	if parent.Type() == "function_declaration" || parent.Type() == "class_declaration" {
		nameNode := parent.ChildByFieldName("name")
		return nameNode != nil && nameNode.ID() == node.ID()
	}
	if parent.Type() == "variable_declarator" {
		nameNode := parent.ChildByFieldName("name")
		return nameNode != nil && nameNode.ID() == node.ID()
	}
	if parent.Type() == "import_specifier" || parent.Type() == "namespace_import" || parent.Type() == "import_clause" {
		return true
	}
	if parent.Type() == "shorthand_property_identifier_pattern" || parent.Type() == "pair_pattern" {
		return true
	}
	return false
}

func collectFunctionDecls(root *sitter.Node, content []byte, result *astResult) {
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	for {
		node := cursor.CurrentNode()
		if node.Type() == "function_declaration" {
			nameNode := node.ChildByFieldName("name")
			if nameNode != nil {
				name := nodeContent(nameNode, content)
				if name != "" {
					result.FunctionDecls = append(result.FunctionDecls, name)
				}
			}
		}

		if cursor.GoToFirstChild() {
			continue
		}
		for !cursor.GoToNextSibling() {
			if !cursor.GoToParent() {
				return
			}
		}
	}
}

func collectVariableDecls(root *sitter.Node, content []byte, result *astResult) {
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	for {
		node := cursor.CurrentNode()
		if node.Type() == "variable_declarator" {
			nameNode := node.ChildByFieldName("name")
			if nameNode != nil {
				name := nodeContent(nameNode, content)
				if name != "" {
					result.VariableDecls = append(result.VariableDecls, name)
				}
			}
		}

		if cursor.GoToFirstChild() {
			continue
		}
		for !cursor.GoToNextSibling() {
			if !cursor.GoToParent() {
				return
			}
		}
	}
}

func collectImportSpecs(root *sitter.Node, content []byte, result *astResult) {
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	for {
		node := cursor.CurrentNode()
		switch node.Type() {
		case "import_statement":
			specs := parseImportStatement(node, content)
			result.ImportSpecs = append(result.ImportSpecs, specs...)
		case "import_clause":
			if parent := node.Parent(); parent != nil && parent.Type() == "import_statement" {
				specs := parseImportStatement(parent, content)
				result.ImportSpecs = append(result.ImportSpecs, specs...)
			}
		case "lexical_declaration":
			if spec := parseRequireStatement(node, content); spec != nil {
				result.ImportSpecs = append(result.ImportSpecs, *spec)
			}
		}

		if cursor.GoToFirstChild() {
			continue
		}
		for !cursor.GoToNextSibling() {
			if !cursor.GoToParent() {
				return
			}
		}
	}
}

func collectExportSymbols(root *sitter.Node, content []byte, result *astResult) {
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	for {
		node := cursor.CurrentNode()
		switch node.Type() {
		case "export_statement", "export_named_declaration":
			parsed := parseExportStatement(node, content)
			result.ExportSymbols = append(result.ExportSymbols, parsed.Symbols...)
			result.ImportSpecs = append(result.ImportSpecs, parsed.Reexports...)
		case "export_clause":
			if parent := node.Parent(); parent != nil && parent.Type() == "export_statement" {
				parsed := parseExportStatement(parent, content)
				result.ExportSymbols = append(result.ExportSymbols, parsed.Symbols...)
				result.ImportSpecs = append(result.ImportSpecs, parsed.Reexports...)
			}
		case "export_default_declaration":
			result.ExportSymbols = append(result.ExportSymbols, ExportSymbol{
				Name: "default",
				Line: int(node.StartPoint().Row) + 1,
			})
		}

		if cursor.GoToFirstChild() {
			continue
		}
		for !cursor.GoToNextSibling() {
			if !cursor.GoToParent() {
				return
			}
		}
	}
}

func parseImportStatement(node *sitter.Node, content []byte) []ImportSpec {
	if node == nil {
		return nil
	}
	sourceNode := node.ChildByFieldName("source")
	if sourceNode == nil {
		return nil
	}
	source := trimQuotes(nodeContent(sourceNode, content))
	if source == "" {
		return nil
	}

	clause := node.ChildByFieldName("name")
	if clause == nil {
		clause = node.ChildByFieldName("import")
	}
	if clause == nil {
		clause = node.ChildByFieldName("clause")
	}

	if clause == nil {
		return []ImportSpec{{Source: source, SideEffect: true, Wildcard: true}}
	}

	specs := []ImportSpec{}
	switch clause.Type() {
	case "identifier":
		specs = append(specs, ImportSpec{Source: source, Names: []string{nodeContent(clause, content)}, Wildcard: true})
	case "namespace_import":
		nameNode := clause.ChildByFieldName("name")
		if nameNode != nil {
			specs = append(specs, ImportSpec{Source: source, Names: []string{nodeContent(nameNode, content)}, Wildcard: true})
		}
	case "named_imports":
		names := parseNamedImports(clause, content)
		specs = append(specs, ImportSpec{Source: source, Names: names})
	case "import_clause":
		specs = append(specs, parseImportClause(clause, content, source)...)
	}
	if len(specs) == 0 {
		specs = append(specs, ImportSpec{Source: source, SideEffect: true, Wildcard: true})
	}
	return specs
}

func parseImportClause(node *sitter.Node, content []byte, source string) []ImportSpec {
	if node == nil {
		return nil
	}
	specs := []ImportSpec{}
	defaultNode := node.ChildByFieldName("name")
	if defaultNode != nil && defaultNode.Type() == "identifier" {
		specs = append(specs, ImportSpec{Source: source, Names: []string{nodeContent(defaultNode, content)}, Wildcard: true})
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Type() {
		case "named_imports":
			names := parseNamedImports(child, content)
			specs = append(specs, ImportSpec{Source: source, Names: names})
		case "namespace_import":
			nameNode := child.ChildByFieldName("name")
			if nameNode != nil {
				specs = append(specs, ImportSpec{Source: source, Names: []string{nodeContent(nameNode, content)}, Wildcard: true})
			}
		}
	}

	return specs
}

func parseNamedImports(node *sitter.Node, content []byte) []string {
	if node == nil {
		return nil
	}
	names := []string{}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil || child.Type() != "import_specifier" {
			continue
		}
		nameNode := child.ChildByFieldName("name")
		if nameNode == nil {
			nameNode = child.ChildByFieldName("value")
		}
		if nameNode != nil {
			name := nodeContent(nameNode, content)
			if name != "" {
				names = append(names, name)
			}
		}
	}
	return names
}

func parseRequireStatement(node *sitter.Node, content []byte) *ImportSpec {
	if node == nil {
		return nil
	}
	declarator := findDescendant(node, "variable_declarator")
	if declarator == nil {
		return nil
	}
	value := declarator.ChildByFieldName("value")
	if value == nil || value.Type() != "call_expression" {
		return nil
	}
	callee := value.ChildByFieldName("function")
	if callee == nil || nodeContent(callee, content) != "require" {
		return nil
	}
	args := value.ChildByFieldName("arguments")
	if args == nil {
		return nil
	}
	sourceNode := firstStringChild(args, content)
	if sourceNode == "" {
		return nil
	}

	nameNode := declarator.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	if nameNode.Type() == "identifier" {
		return &ImportSpec{Source: sourceNode, Names: []string{nodeContent(nameNode, content)}, Wildcard: true}
	}
	if nameNode.Type() == "object_pattern" {
		names := parseObjectPattern(nameNode, content)
		return &ImportSpec{Source: sourceNode, Names: names, Wildcard: true}
	}
	return nil
}

type exportParseResult struct {
	Symbols   []ExportSymbol
	Reexports []ImportSpec
}

func parseExportStatement(node *sitter.Node, content []byte) exportParseResult {
	if node == nil {
		return exportParseResult{}
	}
	line := int(node.StartPoint().Row) + 1
	results := []ExportSymbol{}
	reexports := []ImportSpec{}

	decl := node.ChildByFieldName("declaration")
	if decl != nil {
		switch decl.Type() {
		case "function_declaration", "class_declaration":
			nameNode := decl.ChildByFieldName("name")
			if nameNode != nil {
				name := nodeContent(nameNode, content)
				if name != "" {
					results = append(results, ExportSymbol{Name: name, Line: line})
				}
			}
		case "lexical_declaration", "variable_declaration":
			for _, name := range parseVarDeclarationNames(decl, content) {
				results = append(results, ExportSymbol{Name: name, Line: line})
			}
		}
		return exportParseResult{Symbols: results}
	}

	clause := node.ChildByFieldName("clause")
	if clause != nil && clause.Type() == "export_clause" {
		exportedNames := parseExportSpecifiers(clause, content, true)
		for _, name := range exportedNames {
			if name != "" {
				results = append(results, ExportSymbol{Name: name, Line: line})
			}
		}
	}

	if sourceNode := node.ChildByFieldName("source"); sourceNode != nil {
		source := trimQuotes(nodeContent(sourceNode, content))
		if source != "" {
			spec := ImportSpec{Source: source, Wildcard: true, IsReexport: true}
			if clause == nil {
				spec.Names = []string{"*"}
			} else {
				localNames := parseExportSpecifiers(clause, content, false)
				spec.Names = localNames
			}
			reexports = append(reexports, spec)
		}
	}
	return exportParseResult{Symbols: results, Reexports: reexports}
}

func parseExportSpecifiers(clause *sitter.Node, content []byte, exportedNames bool) []string {
	if clause == nil {
		return nil
	}
	names := []string{}
	for i := 0; i < int(clause.ChildCount()); i++ {
		child := clause.Child(i)
		if child == nil || child.Type() != "export_specifier" {
			continue
		}
		var nameNode *sitter.Node
		if exportedNames {
			nameNode = child.ChildByFieldName("name")
			if nameNode == nil {
				nameNode = child.ChildByFieldName("value")
			}
		} else {
			nameNode = child.ChildByFieldName("value")
			if nameNode == nil {
				nameNode = child.ChildByFieldName("name")
			}
		}
		if nameNode != nil {
			name := nodeContent(nameNode, content)
			if name != "" {
				names = append(names, name)
			}
		}
	}
	return names
}

func parseVarDeclarationNames(node *sitter.Node, content []byte) []string {
	if node == nil {
		return nil
	}
	names := []string{}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Type() == "variable_declarator" {
			nameNode := child.ChildByFieldName("name")
			if nameNode != nil {
				names = append(names, collectPatternNames(nameNode, content)...)
			}
		}
	}
	return names
}

func collectPatternNames(node *sitter.Node, content []byte) []string {
	if node == nil {
		return nil
	}
	if node.Type() == "identifier" {
		name := nodeContent(node, content)
		if name != "" {
			return []string{name}
		}
	}
	if node.Type() == "object_pattern" {
		return parseObjectPattern(node, content)
	}
	if node.Type() == "array_pattern" {
		return parseArrayPattern(node, content)
	}

	names := []string{}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		names = append(names, collectPatternNames(child, content)...)
	}
	return names
}

func parseObjectPattern(node *sitter.Node, content []byte) []string {
	if node == nil {
		return nil
	}
	names := []string{}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Type() == "pair_pattern" || child.Type() == "shorthand_property_identifier_pattern" {
			nameNode := child.ChildByFieldName("name")
			if nameNode == nil {
				nameNode = child.ChildByFieldName("value")
			}
			if nameNode != nil {
				names = append(names, collectPatternNames(nameNode, content)...)
			}
		}
	}
	return names
}

func parseArrayPattern(node *sitter.Node, content []byte) []string {
	if node == nil {
		return nil
	}
	names := []string{}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		names = append(names, collectPatternNames(child, content)...)
	}
	return names
}

func firstStringChild(node *sitter.Node, content []byte) string {
	if node == nil {
		return ""
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Type() == "string" || child.Type() == "string_fragment" || child.Type() == "template_string" {
			return trimQuotes(nodeContent(child, content))
		}
	}
	return ""
}

func findDescendant(node *sitter.Node, nodeType string) *sitter.Node {
	if node == nil {
		return nil
	}
	if node.Type() == nodeType {
		return node
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if found := findDescendant(child, nodeType); found != nil {
			return found
		}
	}
	return nil
}

func trimQuotes(value string) string {
	return strings.Trim(value, "\"'`")
}

func collectFlagHits(root *sitter.Node, content []byte, result *astResult, flagPatterns []string) {
	regexes := compileRegexes(flagPatterns)
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	for {
		node := cursor.CurrentNode()
		if node.Type() == "member_expression" || node.Type() == "subscript_expression" {
			if hit := parseFlagHit(node, content, regexes); hit != nil {
				result.FlagHits = append(result.FlagHits, *hit)
			}
		}

		if cursor.GoToFirstChild() {
			continue
		}
		for !cursor.GoToNextSibling() {
			if !cursor.GoToParent() {
				return
			}
		}
	}
}

func parseFlagHit(node *sitter.Node, content []byte, regexes []*regexp.Regexp) *FlagOccurrence {
	if node == nil {
		return nil
	}
	objectNode := node.ChildByFieldName("object")
	propertyNode := node.ChildByFieldName("property")
	if objectNode == nil || propertyNode == nil {
		return nil
	}

	objectName := nodeContent(objectNode, content)
	propertyName := nodeContent(propertyNode, content)
	if objectName == "" || propertyName == "" {
		return nil
	}
	flag := objectName + "." + propertyName
	if len(regexes) > 0 {
		matched := false
		for _, re := range regexes {
			if re.MatchString(flag) {
				matched = true
				break
			}
		}
		if !matched {
			return nil
		}
	}
	return &FlagOccurrence{
		Flag: flag,
		Line: int(node.StartPoint().Row) + 1,
	}
}

func nodeContent(node *sitter.Node, content []byte) string {
	if node == nil {
		return ""
	}
	start := node.StartByte()
	end := node.EndByte()
	if start >= end || int(end) > len(content) {
		return ""
	}
	return string(content[start:end])
}
