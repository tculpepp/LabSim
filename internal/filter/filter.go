// Package filter blocks bad language in names players type (S-2).
//
// It normalizes leetspeak and case, catches stretched letters ("fuuuck")
// and spaced-out words ("s h i t"), and uses an allow-list so innocent
// words like "peacock" or "class" are not blocked.
package filter

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// MaxLen is the longest name allowed, in characters.
const MaxLen = 26

var (
	ErrEmpty    = errors.New("name is empty")
	ErrTooLong  = errors.New("name is too long")
	ErrBadChars = errors.New("name has characters that are not allowed")
	ErrBlocked  = errors.New("name has words that are not allowed")
)

var leet = strings.NewReplacer(
	"0", "o", "1", "i", "3", "e", "4", "a", "5", "s", "7", "t", "8", "b", "9", "g",
	"@", "a", "$", "s", "!", "i", "|", "i", "+", "t",
)

var (
	strongRe []*regexp.Regexp
	wordRe   []*regexp.Regexp
	allow    = map[string]bool{}
)

// stretch turns "shit" into "s+h+i+t+" so repeated letters still match.
func stretch(root string) string {
	var b strings.Builder
	for _, r := range root {
		b.WriteString(regexp.QuoteMeta(string(r)) + "+")
	}
	return b.String()
}

func init() {
	for _, r := range strongRoots {
		strongRe = append(strongRe, regexp.MustCompile(stretch(r)))
	}
	for _, r := range wordRoots {
		wordRe = append(wordRe, regexp.MustCompile(`^(?:`+stretch(r)+`)(?:s|es|er|ers|ing|y|ed|head|face|hole)?$`))
	}
	for _, w := range allowWords {
		allow[w] = true
	}
}

// Clean trims and collapses spaces.
func Clean(s string) string { return strings.Join(strings.Fields(s), " ") }

// Check returns nil if name is OK to show, or why it is not.
func Check(name string) error {
	name = Clean(name)
	if name == "" {
		return ErrEmpty
	}
	if blocked(name) {
		return ErrBlocked
	}
	if len([]rune(name)) > MaxLen {
		return ErrTooLong
	}
	for _, r := range name {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '.' || r == '-' || r == '\'') {
			return ErrBadChars
		}
	}
	return nil
}

func blocked(name string) bool {
	norm := leet.Replace(strings.ToLower(name))
	tokens := strings.FieldsFunc(norm, func(r rune) bool { return !unicode.IsLetter(r) })
	for _, t := range candidates(tokens) {
		if isBad(t) {
			return true
		}
	}
	return false
}

// candidates is every token plus runs of short tokens joined together,
// so "s h i t" and "sh it" are checked as "shit".
func candidates(tokens []string) []string {
	out := append([]string(nil), tokens...)
	var run []string
	flush := func() {
		if len(run) > 1 {
			out = append(out, strings.Join(run, ""))
		}
		run = run[:0]
	}
	for _, t := range tokens {
		if len([]rune(t)) <= 2 {
			run = append(run, t)
		} else {
			flush()
		}
	}
	flush()
	return out
}

func isBad(tok string) bool {
	for _, re := range wordRe {
		if re.MatchString(tok) {
			return true
		}
	}
	if allow[tok] {
		return false
	}
	for _, re := range strongRe {
		if re.MatchString(tok) {
			return true
		}
	}
	return false
}
