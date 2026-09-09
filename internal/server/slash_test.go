package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeSlashFile creates path (and parents) with content.
func writeSlashFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// setSlashHome points discovery at a temp home for the duration of the test.
func setSlashHome(t *testing.T, home string) {
	t.Helper()
	prev := slashHomeDir
	slashHomeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { slashHomeDir = prev })
}

func slashRequest(t *testing.T, cwd string) []slashCommand {
	t.Helper()
	h := &Hub{}
	url := "/api/slash"
	if cwd != "" {
		url += "?cwd=" + cwd
	}
	rec := httptest.NewRecorder()
	h.handleSlash(rec, httptest.NewRequest(http.MethodGet, url, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	var body struct {
		Commands []slashCommand `json:"commands"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return body.Commands
}

func TestSlashDiscovery(t *testing.T) {
	home := t.TempDir()
	setSlashHome(t, home)

	// User commands: flat, nested, and frontmatter description.
	writeSlashFile(t, filepath.Join(home, ".claude", "commands", "deploy.md"),
		"---\ndescription: ship it to prod\n---\nbody\n")
	writeSlashFile(t, filepath.Join(home, ".claude", "commands", "foo", "bar.md"),
		"\nFirst line of body used as description\nmore\n")
	// Dot-directory inside commands must be skipped.
	writeSlashFile(t, filepath.Join(home, ".claude", "commands", ".hidden", "nope.md"), "hidden\n")
	// Skill.
	writeSlashFile(t, filepath.Join(home, ".claude", "skills", "prune", "SKILL.md"),
		"---\nname: prune\ndescription: "+strings.Repeat("x", 100)+"\n---\n")
	// Plugin cache command.
	writeSlashFile(t, filepath.Join(home, ".claude", "plugins", "cache", "mkt", "pkg", "v1", "commands", "plug.md"),
		"plugin command\n")

	// Project commands, one colliding with a user command (user wins).
	proj := t.TempDir()
	writeSlashFile(t, filepath.Join(proj, ".claude", "commands", "deploy.md"),
		"project deploy that must lose the dedupe\n")
	writeSlashFile(t, filepath.Join(proj, ".claude", "commands", "local.md"),
		"---\ndescription: project-only command\n---\n")

	tests := []struct {
		name string
		cwd  string
		want map[string]slashCommand // cmd -> expected entry
		miss []string                // cmds that must be absent
	}{
		{
			name: "home plus project",
			cwd:  proj,
			want: map[string]slashCommand{
				"/deploy":  {Cmd: "/deploy", Desc: "ship it to prod", Source: "user"},
				"/foo:bar": {Cmd: "/foo:bar", Desc: "First line of body used as description", Source: "user"},
				"/plug":    {Cmd: "/plug", Desc: "plugin command", Source: "user"},
				"/local":   {Cmd: "/local", Desc: "project-only command", Source: "project"},
				"/prune":   {Cmd: "/prune", Desc: strings.Repeat("x", 80), Source: "skill"},
			},
			miss: []string{"/.hidden:nope", "/hidden:nope"},
		},
		{
			name: "garbage cwd yields only home results",
			cwd:  filepath.Join(proj, "does-not-exist"),
			want: map[string]slashCommand{
				"/deploy": {Cmd: "/deploy", Desc: "ship it to prod", Source: "user"},
			},
			miss: []string{"/local"},
		},
		{
			name: "relative cwd ignored",
			cwd:  "not/absolute",
			miss: []string{"/local"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slashRequest(t, tt.cwd)
			byCmd := make(map[string]slashCommand, len(got))
			for _, c := range got {
				if _, dup := byCmd[c.Cmd]; dup {
					t.Errorf("duplicate cmd %q in response", c.Cmd)
				}
				byCmd[c.Cmd] = c
			}
			for cmd, want := range tt.want {
				if byCmd[cmd] != want {
					t.Errorf("%s = %+v, want %+v", cmd, byCmd[cmd], want)
				}
			}
			for _, cmd := range tt.miss {
				if _, ok := byCmd[cmd]; ok {
					t.Errorf("%s present, want absent", cmd)
				}
			}
		})
	}
}

func TestSlashOrderingAndDedupe(t *testing.T) {
	home := t.TempDir()
	setSlashHome(t, home)

	// A skill sharing a name with a command: the command wins the dedupe, and
	// commands sort before skills.
	writeSlashFile(t, filepath.Join(home, ".claude", "commands", "zeta.md"), "command zeta\n")
	writeSlashFile(t, filepath.Join(home, ".claude", "commands", "alpha.md"), "command alpha\n")
	writeSlashFile(t, filepath.Join(home, ".claude", "skills", "zeta", "SKILL.md"),
		"---\ndescription: skill zeta\n---\n")
	writeSlashFile(t, filepath.Join(home, ".claude", "skills", "beta", "SKILL.md"),
		"---\ndescription: skill beta\n---\n")

	got := slashRequest(t, "")
	var cmds []string
	for _, c := range got {
		cmds = append(cmds, c.Cmd+":"+c.Source)
	}
	want := []string{"/alpha:user", "/zeta:user", "/beta:skill"}
	if len(cmds) != len(want) {
		t.Fatalf("commands = %v, want %v", cmds, want)
	}
	for i := range want {
		if cmds[i] != want[i] {
			t.Fatalf("commands = %v, want %v", cmds, want)
		}
	}
}

func TestSlashRejectsPost(t *testing.T) {
	setSlashHome(t, t.TempDir())
	h := &Hub{}
	rec := httptest.NewRecorder()
	h.handleSlash(rec, httptest.NewRequest(http.MethodPost, "/api/slash", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}
