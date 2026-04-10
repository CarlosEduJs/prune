package rules

type Finding struct {
	ID         string                 `json:"id"`
	Kind       string                 `json:"kind"`
	Confidence string                 `json:"confidence"`
	File       string                 `json:"file"`
	Line       int                    `json:"line"`
	Symbol     string                 `json:"symbol"`
	Reason     string                 `json:"reason"`
	Evidence   map[string]interface{} `json:"evidence,omitempty"`
}

type RuleDefinition struct {
	ID          string
	Description string
}

func All() []RuleDefinition {
	return []RuleDefinition{
		{ID: "unused_function", Description: "Function declared but never used"},
		{ID: "unused_variable", Description: "Variable declared but never used"},
		{ID: "unused_export", Description: "Exported symbol never imported"},
		{ID: "unused_file", Description: "File never imported or referenced"},
		{ID: "dead_feature_flag", Description: "Feature flag with dead code"},
		{ID: "suspicious_dynamic_usage", Description: "Dynamic usage that blocks certainty"},
	}
}

func ConfidenceRank(confidence string) int {
	switch confidence {
	case "safe":
		return 1
	case "likely_dead":
		return 2
	case "review":
		return 3
	default:
		return 0
	}
}
