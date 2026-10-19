// Package prompt reads a blocked agent's question and answer options off its
// screen. Herdr reports *that* an agent is blocked but not *what* it asks, so
// the bridge reads the same bottom-of-screen snapshot Herdr classified
// (`pane.read` with source "detection") and parses the picker drawn there.
// It is a pure leaf: text in, protocol.Prompt out, no I/O.
package prompt

import (
	"regexp"
	"strings"

	"github.com/sarathsp06/herdrweb/internal/protocol"
)

var (
	// An option row: optional cursor, then a number ("1.") or a radio/checkbox
	// glyph, then the label.
	optionRe = regexp.MustCompile(`^\s*([❯›>])?\s*(?:(\d+)\.|[○●◉◯□■☐☑])\s+(\S.*?)\s*$`)
	// Box-drawing rules that separate parts of a picker.
	ruleRe        = regexp.MustCompile(`^\s*[├╭╰─┤┌└]`)
	arrowHintRe   = regexp.MustCompile(`↑\s*/\s*↓|↑↓`)
	recommendedRe = regexp.MustCompile(`(?i)\s*\(recommended\)\s*`)
	needsTextRe   = regexp.MustCompile(`(?i)write-in|type your own|tell \w+ what to do|something else`)
	yesNoRe       = regexp.MustCompile(`(?i)\s*[(\[](?:y/n|yes/no)[)\]]\s*$`)
	tabSplitRe    = regexp.MustCompile(`\s{3,}`)
)

// questionLookup is how many lines above the first option hold the question.
const questionLookup = 8

// unbox drops a leading and trailing box border so boxed and plain pickers
// parse the same way.
func unbox(l string) string {
	l = strings.TrimRight(l, " \t")
	l = strings.TrimSuffix(l, "│")
	if t := strings.TrimLeft(l, " "); strings.HasPrefix(t, "│") {
		l = strings.TrimPrefix(t, "│")
	}
	return strings.TrimRight(l, " \t")
}

func indent(l string) int { return len(l) - len(strings.TrimLeft(l, " \t")) }

type row struct {
	line   int
	cursor bool
	num    string
	label  string
	desc   []string
}

// Parse returns the prompt drawn at the bottom of a blocked agent's screen, or
// nil when it finds no question it can offer answers for.
func Parse(text string) *protocol.Prompt {
	lines := strings.Split(text, "\n")
	if p := parsePicker(lines); p != nil {
		return p
	}
	return parseYesNo(lines)
}

func parsePicker(lines []string) *protocol.Prompt {
	// Group option rows into blocks. Blank lines and description lines under an
	// option keep a block open; anything else closes it.
	var blocks [][]row
	var cur []row
	for i, raw := range lines {
		l := unbox(raw)
		if m := optionRe.FindStringSubmatch(l); m != nil {
			cur = append(cur, row{line: i, cursor: m[1] != "", num: m[2], label: m[3]})
			continue
		}
		t := strings.TrimSpace(l)
		if len(cur) > 0 && t == "" {
			continue
		}
		if len(cur) > 0 && !ruleRe.MatchString(raw) && (strings.HasPrefix(t, "↳") || indent(l) >= 4) {
			last := &cur[len(cur)-1]
			last.desc = append(last.desc, strings.TrimSpace(strings.TrimPrefix(t, "↳")))
			continue
		}
		if len(cur) > 0 {
			blocks = append(blocks, cur)
			cur = nil
		}
	}
	if len(cur) > 0 {
		blocks = append(blocks, cur)
	}
	// The live picker is the last block with a real choice in it; earlier
	// blocks are transcript history or a summary of the questions.
	var rows []row
	for i := len(blocks) - 1; i >= 0; i-- {
		if len(blocks[i]) >= 2 {
			rows = blocks[i]
			break
		}
	}
	if rows == nil {
		return nil
	}

	first := rows[0].line
	p := &protocol.Prompt{}
	for j := first - 1; j >= 0 && j >= first-questionLookup; j-- {
		t := strings.TrimSpace(unbox(lines[j]))
		if t == "" || ruleRe.MatchString(lines[j]) || strings.Trim(t, "─ ") == "" {
			continue
		}
		if tabs := tabRow(t); tabs != nil {
			p.Tabs = tabs
			continue
		}
		if p.Question == "" {
			p.Question = strings.TrimPrefix(t, "? ")
		}
	}

	tail := strings.Join(lines[rows[len(rows)-1].line:], "\n")
	cursor, numbered := -1, true
	for i, r := range rows {
		if r.cursor && cursor < 0 {
			cursor = i
		}
		if r.num == "" {
			numbered = false
		}
	}
	// A numbered list in the transcript looks like a menu. A real picker has a
	// cursor on one option, or at least a question above it.
	if cursor < 0 && !strings.HasSuffix(p.Question, "?") {
		return nil
	}
	switch {
	case cursor >= 0 && arrowHintRe.MatchString(tail):
		p.Nav = "arrows"
	case numbered:
		p.Nav = "digits"
	case cursor >= 0:
		p.Nav = "arrows"
	}

	for i, r := range rows {
		label := r.label
		rec := recommendedRe.MatchString(label)
		if rec {
			label = strings.TrimSpace(recommendedRe.ReplaceAllString(label, " "))
		}
		o := protocol.PromptOption{
			Label:       label,
			Desc:        strings.Join(r.desc, " "),
			Recommended: rec,
			NeedsText:   needsTextRe.MatchString(label),
			Keys:        keysFor(p.Nav, i, cursor, r.num),
		}
		p.Options = append(p.Options, o)
	}
	return p
}

// keysFor returns the key presses that pick option i: its digit, or arrow
// moves from where the cursor sits now, then enter.
func keysFor(nav string, i, cursor int, num string) []string {
	switch nav {
	case "digits":
		return []string{num}
	case "arrows":
		if cursor < 0 {
			cursor = 0
		}
		keys := []string{}
		for range max(i-cursor, 0) {
			keys = append(keys, "down")
		}
		for range max(cursor-i, 0) {
			keys = append(keys, "up")
		}
		return append(keys, "enter")
	}
	return []string{}
}

// tabRow recognizes the tab strip of a multi-question form: three or more
// widely spaced labels ending in "Submit".
func tabRow(t string) []string {
	segs := tabSplitRe.Split(t, -1)
	if len(segs) < 3 || !strings.EqualFold(strings.TrimSpace(segs[len(segs)-1]), "submit") {
		return nil
	}
	return segs
}

// parseYesNo handles the plain "Do it? (y/n)" prompt: no picker, two keys.
func parseYesNo(lines []string) *protocol.Prompt {
	seen := 0
	for i := len(lines) - 1; i >= 0 && seen < 6; i-- {
		t := strings.TrimSpace(unbox(lines[i]))
		if t == "" {
			continue
		}
		seen++
		if !yesNoRe.MatchString(t) {
			continue
		}
		q := strings.TrimSpace(yesNoRe.ReplaceAllString(t, ""))
		return &protocol.Prompt{
			Question: strings.TrimSpace(strings.TrimPrefix(q, "?")),
			Nav:      "keys",
			Options: []protocol.PromptOption{
				{Label: "Yes", Keys: []string{"y"}},
				{Label: "No", Keys: []string{"n"}},
			},
		}
	}
	return nil
}

// Truncate shortens s to at most n runes for a notification body.
func Truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return strings.TrimSpace(string(r[:n-1])) + "…"
}
