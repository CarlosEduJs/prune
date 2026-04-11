package js

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

type astPoint struct {
	Row uint32
}

type astNode interface {
	Type() string
	ChildCount() int
	Child(int) astNode
	ChildByFieldName(string) astNode
	Parent() astNode
	StartPoint() astPoint
	StartByte() uint32
	EndByte() uint32
	ID() uint32
}

type tsNode struct {
	n *sitter.Node
}

func wrapNode(node *sitter.Node) astNode {
	if node == nil {
		return nil
	}
	return &tsNode{n: node}
}

func (t *tsNode) Type() string { return t.n.Type() }

func (t *tsNode) ChildCount() int { return int(t.n.ChildCount()) }

func (t *tsNode) Child(index int) astNode { return wrapNode(t.n.Child(index)) }

func (t *tsNode) ChildByFieldName(name string) astNode {
	return wrapNode(t.n.ChildByFieldName(name))
}

func (t *tsNode) Parent() astNode { return wrapNode(t.n.Parent()) }

func (t *tsNode) StartPoint() astPoint { return astPoint{Row: t.n.StartPoint().Row} }

func (t *tsNode) StartByte() uint32 { return t.n.StartByte() }

func (t *tsNode) EndByte() uint32 { return t.n.EndByte() }

func (t *tsNode) ID() uint32 { return uint32(t.n.ID()) }

type astResult struct {
	Identifiers   map[string]int
	UsageCounts   map[string]int
	FunctionDecls []string
	VariableDecls []string
	FunctionLines map[string]int
	VariableLines map[string]int
	ImportSpecs   []ImportSpec
	ExportSymbols []ExportSymbol
	FlagHits      []FlagOccurrence
}

func collectASTData(ctx context.Context, path string, content []byte, flagPatterns []string) (*astResult, error) {
	lang := languageForPath(path)
	if lang == nil {
		return nil, fmt.Errorf("no language found for path %s", path)
	}

	root, err := parseRoot(ctx, lang, content)
	if err != nil {
		return nil, err
	}

	result := &astResult{
		Identifiers:   map[string]int{},
		UsageCounts:   map[string]int{},
		FunctionDecls: []string{},
		VariableDecls: []string{},
		FunctionLines: map[string]int{},
		VariableLines: map[string]int{},
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

	return result, nil
}

func parseASTRoot(ctx context.Context, lang *sitter.Language, content []byte) (astNode, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	defer parser.Close()

	tree, err := parser.ParseCtx(ctx, nil, content)
	if err != nil {
		return nil, err
	}
	if tree == nil {
		return nil, errors.New("parsed tree is nil")
	}
	root := tree.RootNode()
	if root == nil || root.HasError() {
		return nil, errors.New("root node is nil or has error")
	}
	return wrapNode(root), nil
}

var parseRoot = parseASTRoot

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

func collectIdentifiers(root astNode, content []byte, result *astResult) {
	walkAST(root, func(node astNode) {
		if node.Type() != "identifier" {
			return
		}
		name := nodeContent(node, content)
		if name != "" {
			result.Identifiers[name]++
		}
	})
}

func collectUsageCounts(root astNode, content []byte, result *astResult) {
	walkAST(root, func(node astNode) {
		if node.Type() != "identifier" {
			return
		}
		if isDeclarationIdentifier(node) {
			return
		}
		name := nodeContent(node, content)
		if name != "" {
			result.UsageCounts[name]++
		}
	})
}

func isDeclarationIdentifier(node astNode) bool {
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

func collectFunctionDecls(root astNode, content []byte, result *astResult) {
	walkAST(root, func(node astNode) {
		if node.Type() != "function_declaration" {
			return
		}
		nameNode := node.ChildByFieldName("name")
		if nameNode == nil {
			return
		}
		name := nodeContent(nameNode, content)
		if name == "" {
			return
		}
		result.FunctionDecls = append(result.FunctionDecls, name)
		if _, exists := result.FunctionLines[name]; !exists {
			result.FunctionLines[name] = int(nameNode.StartPoint().Row) + 1
		}
	})
}

func collectVariableDecls(root astNode, content []byte, result *astResult) {
	walkAST(root, func(node astNode) {
		if node.Type() != "variable_declarator" {
			return
		}
		nameNode := node.ChildByFieldName("name")
		if nameNode == nil {
			return
		}
		for _, name := range collectPatternNames(nameNode, content) {
			if name != "" {
				result.VariableDecls = append(result.VariableDecls, name)
				if _, exists := result.VariableLines[name]; !exists {
					result.VariableLines[name] = int(nameNode.StartPoint().Row) + 1
				}
			}
		}
	})
}

func collectImportSpecs(root astNode, content []byte, result *astResult) {
	walkAST(root, func(node astNode) {
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
	})
}

func collectExportSymbols(root astNode, content []byte, result *astResult) {
	walkAST(root, func(node astNode) {
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
			defaultLine := int(node.StartPoint().Row) + 1
			result.ExportSymbols = append(result.ExportSymbols, ExportSymbol{
				Name: "default",
				Line: defaultLine,
			})
			if decl := node.ChildByFieldName("declaration"); decl != nil {
				if decl.Type() == "function_declaration" || decl.Type() == "class_declaration" {
					nameNode := decl.ChildByFieldName("name")
					if nameNode != nil {
						name := nodeContent(nameNode, content)
						if name != "" {
							result.ExportSymbols = append(result.ExportSymbols, ExportSymbol{
								Name: name,
								Line: int(nameNode.StartPoint().Row) + 1,
							})
						}
					}
				} else if decl.Type() == "function" || decl.Type() == "class" {
					result.ExportSymbols = append(result.ExportSymbols, ExportSymbol{
						Name: "default",
						Line: int(decl.StartPoint().Row) + 1,
					})
				} else if decl.Type() == "arrow_function" || decl.Type() == "function_expression" {
					result.ExportSymbols = append(result.ExportSymbols, ExportSymbol{
						Name: "default",
						Line: int(decl.StartPoint().Row) + 1,
					})
				}
			}
		}
	})
}

func parseImportStatement(node astNode, content []byte) []ImportSpec {
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
		specs = append(specs, ImportSpec{Source: source, Names: []string{nodeContent(clause, content), "default"}, Wildcard: true})
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

func parseImportClause(node astNode, content []byte, source string) []ImportSpec {
	if node == nil {
		return nil
	}
	specs := []ImportSpec{}
	defaultNode := node.ChildByFieldName("name")
	if defaultNode != nil && defaultNode.Type() == "identifier" {
		specs = append(specs, ImportSpec{Source: source, Names: []string{nodeContent(defaultNode, content), "default"}, Wildcard: true})
	}

	for i := 0; i < node.ChildCount(); i++ {
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

func parseNamedImports(node astNode, content []byte) []string {
	if node == nil {
		return nil
	}
	names := []string{}
	for i := 0; i < node.ChildCount(); i++ {
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

func parseRequireStatement(node astNode, content []byte) *ImportSpec {
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

func parseExportStatement(node astNode, content []byte) exportParseResult {
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

func parseExportSpecifiers(clause astNode, content []byte, exportedNames bool) []string {
	if clause == nil {
		return nil
	}
	names := []string{}
	for i := 0; i < clause.ChildCount(); i++ {
		child := clause.Child(i)
		if child == nil || child.Type() != "export_specifier" {
			continue
		}
		var nameNode astNode
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

func parseVarDeclarationNames(node astNode, content []byte) []string {
	if node == nil {
		return nil
	}
	names := []string{}
	for i := 0; i < node.ChildCount(); i++ {
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

func collectPatternNames(node astNode, content []byte) []string {
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
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		names = append(names, collectPatternNames(child, content)...)
	}
	return names
}

func parseObjectPattern(node astNode, content []byte) []string {
	if node == nil {
		return nil
	}
	names := []string{}
	for i := 0; i < node.ChildCount(); i++ {
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

func parseArrayPattern(node astNode, content []byte) []string {
	if node == nil {
		return nil
	}
	names := []string{}
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		names = append(names, collectPatternNames(child, content)...)
	}
	return names
}

func firstStringChild(node astNode, content []byte) string {
	if node == nil {
		return ""
	}
	for i := 0; i < node.ChildCount(); i++ {
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

func findDescendant(node astNode, nodeType string) astNode {
	if node == nil {
		return nil
	}
	if node.Type() == nodeType {
		return node
	}
	for i := 0; i < node.ChildCount(); i++ {
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

func collectFlagHits(root astNode, content []byte, result *astResult, flagPatterns []string) {
	regexes := compileRegexes(flagPatterns)
	walkAST(root, func(node astNode) {
		if node.Type() == "member_expression" || node.Type() == "subscript_expression" {
			if hit := parseFlagHit(node, content, regexes); hit != nil {
				result.FlagHits = append(result.FlagHits, *hit)
			}
		}
	})
}

func parseFlagHit(node astNode, content []byte, regexes []*regexp.Regexp) *FlagOccurrence {
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

func nodeContent(node astNode, content []byte) string {
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

func walkAST(root astNode, visit func(astNode)) {
	if root == nil {
		return
	}
	stack := []astNode{root}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		visit(node)
		for i := node.ChildCount() - 1; i >= 0; i-- {
			child := node.Child(i)
			if child != nil {
				stack = append(stack, child)
			}
		}
	}
}

var _ astNode = (*tsNode)(nil)
