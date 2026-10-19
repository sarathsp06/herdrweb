package prompt

import (
	"os"
	"reflect"
	"testing"

	"github.com/sarathsp06/herdrweb/internal/protocol"
)

// The testdata screens are `pane.read --source detection` captures of real
// blocked agents (Claude Code, Antigravity, omp, Codex), with their content
// replaced but their layout kept byte for byte where it matters.
func parseFile(t *testing.T, name string) *protocol.Prompt {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return Parse(string(b))
}

func labels(p *protocol.Prompt) []string {
	out := []string{}
	for _, o := range p.Options {
		out = append(out, o.Label)
	}
	return out
}

func keys(p *protocol.Prompt) [][]string {
	out := [][]string{}
	for _, o := range p.Options {
		out = append(out, o.Keys)
	}
	return out
}

func TestClaudePermissionUsesDigits(t *testing.T) {
	p := parseFile(t, "claude_permission.txt")
	if p == nil {
		t.Fatal("no prompt parsed")
	}
	if p.Question != "Do you want to proceed?" || p.Nav != "digits" {
		t.Fatalf("question/nav = %q/%q", p.Question, p.Nav)
	}
	if want := [][]string{{"1"}, {"2"}, {"3"}}; !reflect.DeepEqual(keys(p), want) {
		t.Fatalf("keys = %v", keys(p))
	}
	if !p.Options[2].NeedsText || p.Options[0].NeedsText {
		t.Fatalf("needsText wrong: %+v", p.Options)
	}
}

func TestNumberedPickerWithArrowHintUsesArrows(t *testing.T) {
	// Antigravity numbers its options but selects with ↑/↓ + enter.
	p := parseFile(t, "agy_question.txt")
	if p == nil || p.Nav != "arrows" {
		t.Fatalf("want arrows, got %+v", p)
	}
	if p.Question != "Question 1/1: What would you like to work on next?" {
		t.Fatalf("question = %q", p.Question)
	}
	if !p.Options[0].Recommended || p.Options[0].Label != "Confirm the prompt and wait behavior works" {
		t.Fatalf("recommended option = %+v", p.Options[0])
	}
	if want := []string{"down", "down", "down", "enter"}; !reflect.DeepEqual(p.Options[3].Keys, want) {
		t.Fatalf("write-in keys = %v", p.Options[3].Keys)
	}
	if !p.Options[3].NeedsText {
		t.Fatal("write-in should need text")
	}
}

func TestBoxedPickerIgnoresSummaryAndStatusLine(t *testing.T) {
	p := parseFile(t, "omp_single.txt")
	if p == nil {
		t.Fatal("no prompt parsed")
	}
	if want := []string{"Option A", "Option B", "Other (type your own)"}; !reflect.DeepEqual(labels(p), want) {
		t.Fatalf("labels = %v", labels(p))
	}
	if p.Question != "Test question: which option do you pick?" || p.Nav != "arrows" {
		t.Fatalf("question/nav = %q/%q", p.Question, p.Nav)
	}
	if want := [][]string{{"enter"}, {"down", "enter"}, {"down", "down", "enter"}}; !reflect.DeepEqual(keys(p), want) {
		t.Fatalf("keys = %v", keys(p))
	}
}

func TestMultiQuestionFormKeepsDescriptionsAndTabs(t *testing.T) {
	p := parseFile(t, "omp_multi.txt")
	if p == nil {
		t.Fatal("no prompt parsed")
	}
	if p.Question != "Where should generated reports be stored?" {
		t.Fatalf("question = %q", p.Question)
	}
	if want := []string{"Storage", "Retention", "Export format", "Cache", "Submit"}; !reflect.DeepEqual(p.Tabs, want) {
		t.Fatalf("tabs = %v", p.Tabs)
	}
	if len(p.Options) != 4 {
		t.Fatalf("want 4 options, got %v", labels(p))
	}
	if p.Options[1].Desc != "Store the rendered file next to the report row." {
		t.Fatalf("desc = %q", p.Options[1].Desc)
	}
	if !p.Options[0].Recommended || p.Options[0].Label != "Object storage, per workspace" {
		t.Fatalf("first option = %+v", p.Options[0])
	}
}

func TestYesNoPrompt(t *testing.T) {
	p := parseFile(t, "codex_yes_no.txt")
	if p == nil {
		t.Fatal("no prompt parsed")
	}
	if p.Question != "Apply patch to src/routes/+layout.svelte?" || p.Nav != "keys" {
		t.Fatalf("question/nav = %q/%q", p.Question, p.Nav)
	}
	if want := [][]string{{"y"}, {"n"}}; !reflect.DeepEqual(keys(p), want) {
		t.Fatalf("keys = %v", keys(p))
	}
}

func TestTranscriptListIsNotAPicker(t *testing.T) {
	if p := parseFile(t, "working_list.txt"); p != nil {
		t.Fatalf("a numbered list in the transcript parsed as %+v", p)
	}
}

func TestEmptyScreen(t *testing.T) {
	if p := Parse(""); p != nil {
		t.Fatalf("got %+v", p)
	}
}

func TestTruncate(t *testing.T) {
	if got := Truncate("abcdef", 4); got != "abc…" {
		t.Fatalf("got %q", got)
	}
	if got := Truncate("abc", 4); got != "abc" {
		t.Fatalf("got %q", got)
	}
}
