package js

import (
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
)

type astResult struct {
	Identifiers   map[string]int
	FunctionDecls []string
	VariableDecls []string
}

func collectASTData(path string, content []byte) (*astResult, bool) {
	if !shouldUseJSAST(path) {
		return nil, false
	}

	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())
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
		FunctionDecls: []string{},
		VariableDecls: []string{},
	}

	collectIdentifiers(root, content, result)
	collectFunctionDecls(root, content, result)
	collectVariableDecls(root, content, result)

	return result, true
}

func shouldUseJSAST(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".js" || ext == ".jsx"
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
