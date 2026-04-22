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

type prettyFormatter struct {
	opts FormatterOptions
}

func (f prettyFormatter) Format(findings []rules.Finding) ([]byte, error) { //nolint:gocyclo
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

		fmt.Fprintf(&b, "%s%s %s%s (%d)\n", color, icon, label, reset, total)
		b.WriteString("\n")

		files := make([]string, 0, len(fileGroups))
		for file := range fileGroups {
			files = append(files, file)
		}
		sort.Strings(files)

		for _, file := range files {
			ff := fileGroups[file]
			type groupedFinding struct {
				Kind   string
				Symbol string
				Reason string
				Lines  []int
			}

			var groups []groupedFinding
			groupMap := map[string]int{}

			for _, finding := range ff {
				key := finding.Kind + "|" + finding.Symbol + "|" + finding.Reason
				if idx, ok := groupMap[key]; ok {
					if finding.Line > 0 {
						groups[idx].Lines = append(groups[idx].Lines, finding.Line)
					}
				} else {
					lines := []int{}
					if finding.Line > 0 {
						lines = append(lines, finding.Line)
					}
					groups = append(groups, groupedFinding{
						Kind:   finding.Kind,
						Symbol: finding.Symbol,
						Reason: finding.Reason,
						Lines:  lines,
					})
					groupMap[key] = len(groups) - 1
				}
			}

			sort.Slice(groups, func(i, j int) bool {
				if groups[i].Kind != groups[j].Kind {
					return groups[i].Kind < groups[j].Kind
				}
				return groups[i].Symbol < groups[j].Symbol
			})

			displayFile := file
			if displayFile == "" {
				displayFile = "(no file)"
			}
			if useColor {
				fmt.Fprintf(&b, "  %s%s%s%s\n", colorCyan, colorBold, displayFile, colorReset)
			} else {
				fmt.Fprintf(&b, "  %s\n", displayFile)
			}

			for _, grp := range groups {
				uniqueLines := []int{}
				seenLines := map[int]bool{}
				for _, l := range grp.Lines {
					if !seenLines[l] {
						seenLines[l] = true
						uniqueLines = append(uniqueLines, l)
					}
				}
				sort.Ints(uniqueLines)

				lineStr := ""
				if len(uniqueLines) == 1 {
					lineStr = fmt.Sprintf("line %d", uniqueLines[0])
				} else if len(uniqueLines) > 1 {
					var strLines []string
					for _, l := range uniqueLines {
						strLines = append(strLines, fmt.Sprint(l))
					}
					lineStr = fmt.Sprintf("lines %s", strings.Join(strLines, ", "))
				}

				kindLabel := kindLabels[grp.Kind]
				if kindLabel == "" {
					kindLabel = grp.Kind
				}

				detail := kindLabel
				if grp.Symbol != "" {
					detail = fmt.Sprintf("%s: %s", kindLabel, grp.Symbol)
				}

				if useColor {
					if lineStr != "" {
						fmt.Fprintf(&b, "  %s└─%s %s %s[%s]%s\n", colorDim, colorReset, detail, colorDim, lineStr, colorReset)
					} else {
						fmt.Fprintf(&b, "  %s└─%s %s\n", colorDim, colorReset, detail)
					}
					if grp.Reason != "" {
						fmt.Fprintf(&b, "     %s│ %s%s\n", colorDim, grp.Reason, colorReset)
					}
				} else {
					if lineStr != "" {
						fmt.Fprintf(&b, "  └─ %s [%s]\n", detail, lineStr)
					} else {
						fmt.Fprintf(&b, "  └─ %s\n", detail)
					}
					if grp.Reason != "" {
						fmt.Fprintf(&b, "     │ %s\n", grp.Reason)
					}
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
		fmt.Fprintf(&b, "%s%sPrune %s%s — %d %s found in %s\n", colorBold, colorCyan, ver, colorReset, count, noun, dur)
	} else {
		fmt.Fprintf(&b, "Prune %s — %d %s found in %s\n", ver, count, noun, dur)
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
		b.WriteString(colorBold + "Summary\n" + colorReset)
		b.WriteString(colorDim + separator + colorReset)
	} else {
		b.WriteString(separator)
		b.WriteString("Summary\n")
		b.WriteString(separator)
	}

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
			fmt.Fprintf(&b, "  %-12s %d\n", t.label, c)
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
		fmt.Fprintf(&b, "  %s%-12s%s %d\n", color, label, reset, c)
	}

	b.WriteString("\n")
	fmt.Fprintf(&b, "  Total        %d\n", len(findings))

	if useColor {
		b.WriteString(colorDim + separator + colorReset)
	} else {
		b.WriteString(separator)
	}

	dur := f.opts.Duration.Round(time.Millisecond)
	if useColor {
		fmt.Fprintf(&b, "%sDone in %s%s\n", colorDim, dur, colorReset)
	} else {
		fmt.Fprintf(&b, "Done in %s\n", dur)
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
