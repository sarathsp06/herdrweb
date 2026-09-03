package protocol

import "testing"

func sampleHerdr() *HerdrSnapshot {
	wt := "~/.herdr/worktrees/api/billing"
	return &HerdrSnapshot{
		FocusedWorkspaceID: "w1", FocusedTabID: "w1:t1", FocusedPaneID: "w1:p2",
		Workspaces: []HerdrWorkspace{
			{WorkspaceID: "w1", Label: "hedr-web", ActiveTabID: "w1:t1", Branch: "feat/chat-ui"},
			{WorkspaceID: "w3", Label: "api/billing", ActiveTabID: "w3:t1", Branch: "feat/billing", Worktree: &wt},
		},
		Tabs: []HerdrTab{
			{TabID: "w1:t1", WorkspaceID: "w1", Label: "agents"},
			{TabID: "w3:t1", WorkspaceID: "w3", Label: "agents"},
		},
		Panes: []HerdrPane{
			{PaneID: "w1:p1", TabID: "w1:t1", WorkspaceID: "w1", Agent: "claude", AgentStatus: "working", Cwd: "/home/u/code/hedr-web"},
			{PaneID: "w1:p2", TabID: "w1:t1", WorkspaceID: "w1", Agent: "codex", AgentStatus: "blocked", Cwd: "/home/u/code/hedr-web"},
			{PaneID: "w1:p3", TabID: "w1:t1", WorkspaceID: "w1", Agent: "", AgentStatus: "", TerminalTitleStripped: "vite dev", Cwd: "/home/u/code/hedr-web"},
			{PaneID: "w3:p1", TabID: "w3:t1", WorkspaceID: "w3", Agent: "claude", AgentStatus: "blocked", Cwd: "/home/u/wt/billing"},
		},
	}
}

func TestNormalizeNesting(t *testing.T) {
	s := Normalize(sampleHerdr())
	if len(s.Spaces) != 2 {
		t.Fatalf("want 2 spaces, got %d", len(s.Spaces))
	}
	if s.Type != "snapshot" {
		t.Fatalf("type = %q", s.Type)
	}
	if s.Focus.PaneID != "w1:p2" {
		t.Fatalf("focus pane = %q", s.Focus.PaneID)
	}
	w1 := s.Spaces[0]
	if w1.Branch != "feat/chat-ui" || w1.Cwd == "" {
		t.Fatalf("space cwd/branch not derived: %+v", w1)
	}
	if len(w1.Tabs) != 1 || len(w1.Tabs[0].Panes) != 3 {
		t.Fatalf("tab/pane nesting wrong: %+v", w1.Tabs)
	}
}

func TestNormalizeAgentFlagAndStatus(t *testing.T) {
	s := Normalize(sampleHerdr())
	panes := s.Spaces[0].Tabs[0].Panes
	byID := map[string]Pane{}
	for _, p := range panes {
		byID[p.ID] = p
	}
	if !byID["w1:p2"].Agent || byID["w1:p2"].Status != Blocked {
		t.Fatalf("codex pane wrong: %+v", byID["w1:p2"])
	}
	if byID["w1:p3"].Agent {
		t.Fatalf("dev pane should not be an agent")
	}
	if byID["w1:p3"].Status != Idle {
		t.Fatalf("non-agent pane should default idle, got %s", byID["w1:p3"].Status)
	}
}

func TestNormalizeWorktree(t *testing.T) {
	s := Normalize(sampleHerdr())
	if s.Spaces[1].Worktree == nil {
		t.Fatalf("worktree provenance lost")
	}
}
