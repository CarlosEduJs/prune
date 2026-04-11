package rules

import "testing"

func TestConfidenceRank(t *testing.T) {
	if ConfidenceRank("safe") != 1 {
		t.Fatalf("expected safe rank 1")
	}
	if ConfidenceRank("likely_dead") != 2 {
		t.Fatalf("expected likely_dead rank 2")
	}
	if ConfidenceRank("review") != 3 {
		t.Fatalf("expected review rank 3")
	}
	if ConfidenceRank("unknown") != 0 {
		t.Fatalf("expected unknown rank 0")
	}
}

func TestAllRulesIncludesExpected(t *testing.T) {
	list := All()
	expected := map[string]bool{
		"unused_function":          false,
		"unused_variable":          false,
		"unused_export":            false,
		"unused_file":              false,
		"dead_feature_flag":        false,
		"suspicious_dynamic_usage": false,
	}

	for _, rule := range list {
		if _, ok := expected[rule.ID]; ok {
			expected[rule.ID] = true
		}
	}

	for id, found := range expected {
		if !found {
			t.Fatalf("missing rule id %s", id)
		}
	}
}
