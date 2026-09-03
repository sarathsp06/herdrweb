// Package protocol defines the normalized UI model the bridge sends to browsers
// and the mapping from Herdr's flat session snapshot onto it. These Go types are
// the canonical source; web/src/lib/protocol mirrors them.
package protocol

// Status is the 5-value agent lifecycle enum.
type Status string

const (
	Working Status = "working"
	Blocked Status = "blocked"
	Done    Status = "done"
	Idle    Status = "idle"
	Unknown Status = "unknown"
)

// Pane is a normalized terminal/agent pane.
type Pane struct {
	ID     string   `json:"id"`
	Label  string   `json:"label"`
	Sub    string   `json:"sub"`
	Status Status   `json:"status"`
	Agent  bool     `json:"agent"`
	Tail   []string `json:"tail"`
}

// Tab groups panes.
type Tab struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Panes []Pane `json:"panes"`
}

// Space is a Herdr workspace, normalized for the UI (README calls it a "space").
type Space struct {
	ID       string  `json:"id"`
	Label    string  `json:"label"`
	Cwd      string  `json:"cwd"`
	Branch   string  `json:"branch"`
	Worktree *string `json:"worktree"`
	Tabs     []Tab   `json:"tabs"`
}

// Focus mirrors the focused workspace/tab/pane.
type Focus struct {
	SpaceID string `json:"spaceId,omitempty"`
	TabID   string `json:"tabId,omitempty"`
	PaneID  string `json:"paneId,omitempty"`
}

// Snapshot is the normalized bootstrap the bridge sends as {type:"snapshot"}.
type Snapshot struct {
	Type   string  `json:"type"` // always "snapshot"
	Spaces []Space `json:"spaces"`
	Focus  Focus   `json:"focus"`
}

// ---- Raw Herdr snapshot shapes (subset we consume) ----

type HerdrSnapshot struct {
	Workspaces         []HerdrWorkspace `json:"workspaces"`
	Tabs               []HerdrTab       `json:"tabs"`
	Panes              []HerdrPane      `json:"panes"`
	FocusedWorkspaceID string           `json:"focused_workspace_id"`
	FocusedTabID       string           `json:"focused_tab_id"`
	FocusedPaneID      string           `json:"focused_pane_id"`
}

type HerdrWorkspace struct {
	WorkspaceID string `json:"workspace_id"`
	Label       string `json:"label"`
	ActiveTabID string `json:"active_tab_id"`
	// Optional provenance fields (present only for git repos / worktrees).
	Branch   string  `json:"branch"`
	Cwd      string  `json:"cwd"`
	Worktree *string `json:"worktree"`
}

type HerdrTab struct {
	TabID       string `json:"tab_id"`
	WorkspaceID string `json:"workspace_id"`
	Label       string `json:"label"`
}

type HerdrPane struct {
	PaneID                string `json:"pane_id"`
	TabID                 string `json:"tab_id"`
	WorkspaceID           string `json:"workspace_id"`
	Agent                 string `json:"agent"`
	AgentStatus           string `json:"agent_status"`
	Cwd                   string `json:"cwd"`
	TerminalTitleStripped string `json:"terminal_title_stripped"`
}

// Normalize maps a raw Herdr snapshot onto the nested UI model.
func Normalize(hs *HerdrSnapshot) Snapshot {
	tabsByWs := map[string][]HerdrTab{}
	for _, t := range hs.Tabs {
		tabsByWs[t.WorkspaceID] = append(tabsByWs[t.WorkspaceID], t)
	}
	panesByTab := map[string][]HerdrPane{}
	for _, p := range hs.Panes {
		panesByTab[p.TabID] = append(panesByTab[p.TabID], p)
	}

	spaces := make([]Space, 0, len(hs.Workspaces))
	for _, ws := range hs.Workspaces {
		sp := Space{ID: ws.WorkspaceID, Label: ws.Label, Branch: ws.Branch, Cwd: ws.Cwd, Worktree: ws.Worktree}
		for _, t := range tabsByWs[ws.WorkspaceID] {
			tab := Tab{ID: t.TabID, Label: t.Label}
			for _, p := range panesByTab[t.TabID] {
				tab.Panes = append(tab.Panes, normalizePane(p))
			}
			sp.Tabs = append(sp.Tabs, tab)
			// Derive space cwd from the first pane if the workspace lacks one.
			if sp.Cwd == "" {
				for _, p := range panesByTab[t.TabID] {
					if p.Cwd != "" {
						sp.Cwd = p.Cwd
						break
					}
				}
			}
		}
		spaces = append(spaces, sp)
	}
	return Snapshot{
		Type:   "snapshot",
		Spaces: spaces,
		Focus:  Focus{SpaceID: hs.FocusedWorkspaceID, TabID: hs.FocusedTabID, PaneID: hs.FocusedPaneID},
	}
}

func normalizePane(p HerdrPane) Pane {
	agent := p.Agent != ""
	label := p.Agent
	if label == "" {
		label = p.TerminalTitleStripped
	}
	if label == "" {
		label = p.PaneID
	}
	sub := label
	if agent && p.Cwd != "" {
		sub = p.Agent + " · " + base(p.Cwd)
	} else if p.TerminalTitleStripped != "" {
		sub = p.TerminalTitleStripped
	}
	return Pane{
		ID:     p.PaneID,
		Label:  label,
		Sub:    sub,
		Status: normStatus(p.AgentStatus, agent),
		Agent:  agent,
		Tail:   []string{},
	}
}

func normStatus(s string, agent bool) Status {
	switch Status(s) {
	case Working, Blocked, Done, Idle, Unknown:
		return Status(s)
	}
	if agent {
		return Unknown
	}
	return Idle
}

func base(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
