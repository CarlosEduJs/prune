package report

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"prune/internal/rules"
	"prune/internal/version"
)

var confidenceOrder = []string{"safe", "likely_dead", "review"}

var kindLabels = map[string]string{
	"unused_file":              "unused file",
	"unused_export":            "unused export",
	"unused_function":          "unused function",
	"unused_variable":          "unused variable",
	"suspicious_dynamic_usage": "suspicious dynamic usage",
	"possible_dynamic_usage":   "possible dynamic usage",
}

var confidenceLabels = map[string]string{
	"safe":        "SAFE",
	"likely_dead": "LIKELY DEAD",
	"review":      "REVIEW",
}

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorDim    = "\033[2m"
	colorBold   = "\033[1m"
	colorCyan   = "\033[36m"
)

var confidenceColors = map[string]string{
	"safe":        colorGreen,
	"likely_dead": colorRed,
	"review":      colorYellow,
}

var confidenceIcons = map[string]string{
	"safe":        "✔",
	"likely_dead": "✖",
	"review":      "⚠",
}

var classificationRules = map[string]string{
	"unused_file":              "safe",
	"unused_export":            "safe",
	"unused_function":          "likely_dead",
	"possible_dynamic_usage":   "review",
	"suspicious_dynamic_usage": "review",
}

type prettyFormatter struct {
	opts FormatterOptions
}

func normalizeClassification(findings []rules.Finding) []rules.Finding {
	result := make([]rules.Finding, len(findings))
	copy(result, findings)
	for i := range result {
		if override, ok := classificationRules[result[i].Kind]; ok {
			result[i].Confidence = override
		}
	}
	return result
}

func (f prettyFormatter) Format(findings []rules.Finding) ([]byte, error) {
	useColor := supportsColor()
	var b strings.Builder

	if len(findings) == 0 {
		b.WriteString(f.header(0, useColor))
		b.WriteString("\n")
		msg := "✨ No dead code found!\n"
		if useColor {
			msg = colorGreen + msg + colorReset
		}
		b.WriteString(msg)
		return []byte(b.String()), nil
	}

	if f.opts.Only != "" {
		filtered := make([]rules.Finding, 0, len(findings))
		for _, finding := range findings {
			if strings.EqualFold(finding.Confidence, f.opts.Only) {
				filtered = append(filtered, finding)
			}
		}
		findings = filtered
	}

	if f.opts.Deletable {
		filtered := make([]rules.Finding, 0, len(findings))
		for _, finding := range findings {
			if finding.Kind == "unused_file" && finding.Confidence == "safe" {
				filtered = append(filtered, finding)
			}
		}
		findings = filtered
	}

	findings = deduplicateUnusedFiles(findings)
	findings = normalizeClassification(findings)

	b.WriteString(f.header(len(findings), useColor))
	b.WriteString("\n")
	if f.opts.Compact {
		b.WriteString(f.summary(findings, useColor))
		return []byte(b.String()), nil
	}

	grouped := groupByConfidence(findings)

	for _, confidence := range confidenceOrder {
		fileGroups, ok := grouped[confidence]
		if !ok || len(fileGroups) == 0 {
			continue
		}

		total := 0
		for _, ff := range fileGroups {
			total += len(ff)
		}

		label := confidenceLabels[confidence]
		icon := confidenceIcons[confidence]
		color := ""
		reset := ""
		if useColor {
			color = confidenceColors[confidence]
			reset = colorReset
		}

		b.WriteString(fmt.Sprintf("%s%s %s%s (%d)\n", color, icon, label, reset, total))
		b.WriteString("\n")

		files := make([]string, 0, len(fileGroups))
		for file := range fileGroups {
			files = append(files, file)
		}
		sort.Strings(files)

		for _, file := range files {
			ff := fileGroups[file]
			sort.Slice(ff, func(i, j int) bool {
				return ff[i].Kind < ff[j].Kind
			})

			displayFile := file
			if displayFile == "" {
				displayFile = "(no file)"
			}
			if useColor {
				b.WriteString(fmt.Sprintf("  %s%s%s\n", colorCyan, displayFile, colorReset))
			} else {
				b.WriteString(fmt.Sprintf("  %s\n", displayFile))
			}

			for _, finding := range ff {
				kindLabel := kindLabels[finding.Kind]
				if kindLabel == "" {
					kindLabel = finding.Kind
				}

				detail := kindLabel
				if finding.Symbol != "" {
					detail = fmt.Sprintf("%s: %s", kindLabel, finding.Symbol)
				}

				if useColor {
					b.WriteString(fmt.Sprintf("  %s└─%s %s\n", colorDim, colorReset, detail))
				} else {
					b.WriteString(fmt.Sprintf("  └─ %s\n", detail))
				}
			}
			b.WriteString("\n")
		}
	}

	b.WriteString(f.summary(findings, useColor))
	return []byte(b.String()), nil
}

func (f prettyFormatter) header(count int, useColor bool) string {
	var b strings.Builder
	ver := "v" + version.Version
	noun := "issues"
	if count == 1 {
		noun = "issue"
	}
	dur := f.opts.Duration.Round(time.Millisecond)

	if useColor {
		b.WriteString(fmt.Sprintf("%s%sPrune %s%s — %d %s found in %s\n", colorBold, colorCyan, ver, colorReset, count, noun, dur))
	} else {
		b.WriteString(fmt.Sprintf("Prune %s — %d %s found in %s\n", ver, count, noun, dur))
	}
	return b.String()
}

func (f prettyFormatter) summary(findings []rules.Finding, useColor bool) string {
	var b strings.Builder

	typeCounts := map[string]int{}
	confCounts := map[string]int{}
	for _, finding := range findings {
		typeCounts[finding.Kind]++
		confCounts[finding.Confidence]++
	}

	separator := "─────────────────────────────────\n"
	if useColor {
		b.WriteString(colorDim + separator + colorReset)
	} else {
		b.WriteString(separator)
	}

	b.WriteString("Summary\n")
	b.WriteString("\n")

	typeOrder := []struct {
		key   string
		label string
	}{
		{"unused_file", "Files"},
		{"unused_export", "Exports"},
		{"unused_function", "Functions"},
		{"unused_variable", "Variables"},
		{"possible_dynamic_usage", "Dynamic"},
		{"suspicious_dynamic_usage", "Suspicious"},
	}
	for _, t := range typeOrder {
		if c, ok := typeCounts[t.key]; ok && c > 0 {
			b.WriteString(fmt.Sprintf("  %-12s %d\n", t.label, c))
		}
	}

	b.WriteString("\n")

	for _, confidence := range confidenceOrder {
		c := confCounts[confidence]
		if c == 0 {
			continue
		}
		label := confidenceLabels[confidence]
		color := ""
		reset := ""
		if useColor {
			color = confidenceColors[confidence]
			reset = colorReset
		}
		b.WriteString(fmt.Sprintf("  %s%-12s%s %d\n", color, label, reset, c))
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Total        %d\n", len(findings)))
	b.WriteString("\n")

	dur := f.opts.Duration.Round(time.Millisecond)
	if useColor {
		b.WriteString(fmt.Sprintf("%sDone in %s%s\n", colorDim, dur, colorReset))
	} else {
		b.WriteString(fmt.Sprintf("Done in %s\n", dur))
	}

	return b.String()
}

func deduplicateUnusedFiles(findings []rules.Finding) []rules.Finding {
	unusedFiles := map[string]bool{}
	for _, f := range findings {
		if f.Kind == "unused_file" {
			unusedFiles[f.File] = true
		}
	}
	if len(unusedFiles) == 0 {
		return findings
	}

	result := make([]rules.Finding, 0, len(findings))
	for _, f := range findings {
		if unusedFiles[f.File] && f.Kind != "unused_file" {
			continue
		}
		result = append(result, f)
	}
	return result
}

func groupByConfidence(findings []rules.Finding) map[string]map[string][]rules.Finding {
	result := map[string]map[string][]rules.Finding{}
	for _, f := range findings {
		conf := f.Confidence
		if result[conf] == nil {
			result[conf] = map[string][]rules.Finding{}
		}
		result[conf][f.File] = append(result[conf][f.File], f)
	}
	return result
}

func supportsColor() bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
