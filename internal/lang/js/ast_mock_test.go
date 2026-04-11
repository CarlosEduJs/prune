package js

import (
	"testing"
	"unsafe"

	sitter "github.com/smacker/go-tree-sitter"
)

type fakeNode struct {
	name     string
	children []*fakeNode
	fields   map[string]*fakeNode
	parent   *fakeNode
	start    uint32
	end      uint32
}

func (f *fakeNode) Type() string    { return f.name }
func (f *fakeNode) ChildCount() int { return len(f.children) }
func (f *fakeNode) Child(i int) astNode {
	if i < 0 || i >= len(f.children) {
		return nil
	}
	return f.children[i]
}
func (f *fakeNode) ChildByFieldName(name string) astNode {
	if f.fields == nil {
		return nil
	}
	return f.fields[name]
}
func (f *fakeNode) Parent() astNode      { return f.parent }
func (f *fakeNode) StartPoint() astPoint { return astPoint{Row: f.start} }
func (f *fakeNode) StartByte() uint32    { return f.start }
func (f *fakeNode) EndByte() uint32      { return f.end }
func (f *fakeNode) ID() uint32           { return uint32(uintptr(unsafe.Pointer(f))) }

func TestCollectASTDataWithMockParser(t *testing.T) {
	root := &fakeNode{name: "program"}
	ident := &fakeNode{name: "identifier", start: 6, end: 10, parent: root}
	root.children = []*fakeNode{ident}
	content := []byte("const demo = 1")

	parseRoot = func(_ *sitter.Language, _ []byte) (astNode, bool) {
		return root, true
	}
	defer func() { parseRoot = parseASTRoot }()

	result, ok := collectASTData("file.ts", content, nil)
	if !ok {
		t.Fatalf("expected parse success")
	}
	if result.Identifiers["demo"] != 1 {
		t.Fatalf("expected identifier demo")
	}
}

func TestCollectASTDataExportsWithMock(t *testing.T) {
	root := &fakeNode{name: "program"}
	declName := &fakeNode{name: "identifier", start: 16, end: 19}
	decl := &fakeNode{name: "function_declaration", fields: map[string]*fakeNode{"name": declName}}
	declName.parent = decl
	decl.start = 7
	decl.end = 10
	decl.parent = root

	exportNode := &fakeNode{name: "export_statement", fields: map[string]*fakeNode{"declaration": decl}, parent: root}
	root.children = []*fakeNode{exportNode}

	content := []byte("export function foo() {}")
	parseRoot = func(_ *sitter.Language, _ []byte) (astNode, bool) { return root, true }
	defer func() { parseRoot = parseASTRoot }()

	result, ok := collectASTData("file.ts", content, nil)
	if !ok {
		t.Fatalf("expected parse success")
	}
	if len(result.ExportSymbols) != 1 || result.ExportSymbols[0].Name != "foo" {
		t.Fatalf("expected export symbol foo")
	}
}

func TestCollectASTDataImportWithMock(t *testing.T) {
	root := &fakeNode{name: "program"}
	importName := &fakeNode{name: "identifier", start: 7, end: 10}
	importClause := &fakeNode{name: "import_clause", children: []*fakeNode{importName}}
	importClause.fields = map[string]*fakeNode{"name": importName}
	importName.parent = importClause
	importSource := &fakeNode{name: "string", start: 16, end: 21}
	importStmt := &fakeNode{name: "import_statement", fields: map[string]*fakeNode{"name": importClause, "source": importSource}, parent: root}
	importClause.parent = importStmt
	importSource.parent = importStmt
	root.children = []*fakeNode{importStmt}

	content := []byte("import foo from 'bar'")
	parseRoot = func(_ *sitter.Language, _ []byte) (astNode, bool) { return root, true }
	defer func() { parseRoot = parseASTRoot }()

	result, ok := collectASTData("file.ts", content, nil)
	if !ok {
		t.Fatalf("expected parse success")
	}
	if len(result.ImportSpecs) != 1 {
		t.Fatalf("expected 1 import spec")
	}
	if result.ImportSpecs[0].Names[0] != "foo" || result.ImportSpecs[0].Source != "bar" {
		t.Fatalf("unexpected import spec")
	}
}

func TestCollectASTDataFeatureFlagsMock(t *testing.T) {
	root := &fakeNode{name: "program"}
	object := &fakeNode{name: "identifier", start: 0, end: 5}
	property := &fakeNode{name: "identifier", start: 6, end: 10}
	member := &fakeNode{name: "member_expression", fields: map[string]*fakeNode{"object": object, "property": property}, parent: root}
	object.parent = member
	property.parent = member
	root.children = []*fakeNode{member}

	content := []byte("flags.ABCD")
	parseRoot = func(_ *sitter.Language, _ []byte) (astNode, bool) { return root, true }
	defer func() { parseRoot = parseASTRoot }()

	result, ok := collectASTData("file.ts", content, []string{"flags\\.[A-Z]+"})
	if !ok {
		t.Fatalf("expected parse success")
	}
	if len(result.FlagHits) != 1 || result.FlagHits[0].Flag != "flags.ABCD" {
		t.Fatalf("expected flag hit")
	}
}
