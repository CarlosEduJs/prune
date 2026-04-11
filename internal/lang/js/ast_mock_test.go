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
