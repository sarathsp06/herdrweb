package server

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// slashCommand is one discovered slash command surfaced to the composer palette.
type slashCommand struct {
	Cmd    string `json:"cmd"`
	Desc   string `json:"desc"`
	Source string `json:"source"`
}

// slashHomeDir resolves the "home" root for command discovery; tests override it.
var slashHomeDir = os.UserHomeDir

// Discovery limits: total commands returned and recursion depth inside a
// commands directory. Both keep a pathological ~/.claude tree from ballooning
// the response or the walk.
const (
	slashMaxCommands = 200
	slashMaxDepth    = 4
	slashDescMax     = 80
)

// handleSlash returns the slash commands discovered for the requesting pane's
// working directory: user/project .claude/commands markdown files, skills, and
// plugin-cache commands. Discovery is best-effort — missing directories simply
// contribute nothing, and the request never fails.
func (h *Hub) handleSlash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cwd := r.URL.Query().Get("cwd")
	writeJSON(w, map[string][]slashCommand{"commands": discoverSlashCommands(cwd)})
}

// discoverSlashCommands scans the home and project trees and returns commands
// sorted user+project first, then skills, alphabetical within each group,
// deduped by cmd (first wins) and capped at slashMaxCommands.
func discoverSlashCommands(cwd string) []slashCommand {
	var cmds, skills []slashCommand

	home, err := slashHomeDir()
	if err == nil && home != "" {
		cmds = append(cmds, collectCommandDir(filepath.Join(home, ".claude", "commands"), "user")...)
		skills = append(skills, collectSkills(filepath.Join(home, ".claude", "skills"))...)
		cmds = append(cmds, collectPluginCommands(filepath.Join(home, ".claude", "plugins", "cache"))...)
	}

	if filepath.IsAbs(cwd) {
		if fi, err := os.Stat(cwd); err == nil && fi.IsDir() {
			cmds = append(cmds, collectCommandDir(filepath.Join(cwd, ".claude", "commands"), "project")...)
			skills = append(skills, collectSkills(filepath.Join(cwd, ".claude", "skills"))...)
		}
	}

	byCmd := func(s []slashCommand) func(i, j int) bool {
		return func(i, j int) bool { return s[i].Cmd < s[j].Cmd }
	}
	sort.SliceStable(cmds, byCmd(cmds))
	sort.SliceStable(skills, byCmd(skills))

	out := make([]slashCommand, 0, len(cmds)+len(skills))
	seen := make(map[string]bool, len(cmds)+len(skills))
	for _, c := range append(cmds, skills...) {
		if seen[c.Cmd] {
			continue
		}
		seen[c.Cmd] = true
		out = append(out, c)
		if len(out) >= slashMaxCommands {
			break
		}
	}
	return out
}

// collectCommandDir walks a .claude/commands tree (depth-bounded) and maps each
// *.md file to a command: relative path minus .md, path separators turned into
// ':' (foo/bar.md -> /foo:bar).
func collectCommandDir(dir, source string) []slashCommand {
	var out []slashCommand
	walkCommands(dir, nil, 0, func(rel []string, path string) {
		out = append(out, slashCommand{
			Cmd:    "/" + strings.Join(rel, ":"),
			Desc:   mdDescription(path),
			Source: source,
		})
	})
	return out
}

// walkCommands recurses over dir collecting *.md files. rel carries the
// name components accumulated so far; dot-directories are skipped.
func walkCommands(dir string, rel []string, depth int, emit func(rel []string, path string)) {
	if depth > slashMaxDepth {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			if strings.HasPrefix(name, ".") {
				continue
			}
			walkCommands(filepath.Join(dir, name), append(rel, name), depth+1, emit)
			continue
		}
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		emit(append(rel, strings.TrimSuffix(name, ".md")), filepath.Join(dir, name))
	}
}

// collectSkills maps <dir>/*/SKILL.md to a "/dirname" command with the
// frontmatter description.
func collectSkills(dir string) []slashCommand {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []slashCommand
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		path := filepath.Join(dir, e.Name(), "SKILL.md")
		if _, err := os.Stat(path); err != nil {
			continue
		}
		out = append(out, slashCommand{
			Cmd:    "/" + e.Name(),
			Desc:   mdDescription(path),
			Source: "skill",
		})
	}
	return out
}

// collectPluginCommands scans the plugin cache: cache/*/*/*/commands/**/*.md.
func collectPluginCommands(cacheDir string) []slashCommand {
	var out []slashCommand
	// Three fixed levels of nesting before the commands dir.
	level1, err := os.ReadDir(cacheDir)
	if err != nil {
		return nil
	}
	for _, a := range level1 {
		if !a.IsDir() {
			continue
		}
		level2, err := os.ReadDir(filepath.Join(cacheDir, a.Name()))
		if err != nil {
			continue
		}
		for _, b := range level2 {
			if !b.IsDir() {
				continue
			}
			level3, err := os.ReadDir(filepath.Join(cacheDir, a.Name(), b.Name()))
			if err != nil {
				continue
			}
			for _, c := range level3 {
				if !c.IsDir() {
					continue
				}
				dir := filepath.Join(cacheDir, a.Name(), b.Name(), c.Name(), "commands")
				out = append(out, collectCommandDir(dir, "user")...)
			}
		}
	}
	return out
}

// mdDescription extracts a one-line description from a markdown command file:
// the YAML frontmatter "description:" value when present, else the first
// non-empty line outside the frontmatter, truncated to slashDescMax runes.
func mdDescription(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	inFront := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if i == 0 && trimmed == "---" {
			inFront = true
			continue
		}
		if inFront {
			if trimmed == "---" {
				inFront = false
				continue
			}
			if v, ok := strings.CutPrefix(trimmed, "description:"); ok {
				return truncateDesc(strings.Trim(strings.TrimSpace(v), `"'`))
			}
			continue
		}
		if trimmed != "" {
			return truncateDesc(trimmed)
		}
	}
	return ""
}

func truncateDesc(s string) string {
	r := []rune(s)
	if len(r) <= slashDescMax {
		return s
	}
	return string(r[:slashDescMax])
}
