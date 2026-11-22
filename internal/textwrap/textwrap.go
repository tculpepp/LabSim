// Package textwrap word-wraps text for dialog boxes and panels.
package textwrap

import "strings"

// Wrap splits text into lines no wider than max (as measured), breaking at
// spaces. Explicit newlines are kept. A word wider than max gets its own line.
func Wrap(text string, max float64, measure func(string) float64) []string {
	var out []string
	for _, para := range strings.Split(text, "\n") {
		words := strings.Fields(para)
		if len(words) == 0 {
			out = append(out, "")
			continue
		}
		line := words[0]
		for _, w := range words[1:] {
			if cand := line + " " + w; measure(cand) <= max {
				line = cand
			} else {
				out = append(out, line)
				line = w
			}
		}
		out = append(out, line)
	}
	return out
}

// Pages groups lines into pages of at most n lines. Always returns at least one page.
func Pages(lines []string, n int) [][]string {
	if len(lines) == 0 {
		return [][]string{{}}
	}
	var out [][]string
	for len(lines) > n {
		out = append(out, lines[:n])
		lines = lines[n:]
	}
	return append(out, lines)
}
