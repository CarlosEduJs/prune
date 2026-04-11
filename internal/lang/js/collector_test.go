package js

import "testing"

func TestUniqueStrings(t *testing.T) {
	values := []string{"a", "b", "a", "", "c"}
	unique := uniqueStrings(values)
	if len(unique) != 3 {
		t.Fatalf("expected 3 values, got %d", len(unique))
	}
}

func TestMergeExportNames(t *testing.T) {
	base := []string{"a"}
	exports := []ExportSymbol{{Name: "b"}, {Name: ""}}
	merged := mergeExportNames(base, exports)
	if len(merged) != 2 {
		t.Fatalf("expected 2 names, got %d", len(merged))
	}
}

func TestFlagHitExists(t *testing.T) {
	hits := []FlagOccurrence{{Flag: "flags.A"}}
	if !flagHitExists(hits, "flags.A") {
		t.Fatalf("expected hit exists")
	}
	if flagHitExists(hits, "flags.B") {
		t.Fatalf("expected hit missing")
	}
}
