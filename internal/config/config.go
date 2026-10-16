// Package config reads and writes the Herdr Web settings. They live in their
// own file, FileName, beside Herdr's config.toml; Migrate moves them out of the
// [web] table that earlier versions wrote into config.toml itself.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

// FileName is the settings file, kept in the directory of Herdr's config.toml
// alongside the push keys and subscriptions. Herdr reports unknown tables in
// config.toml ("unknown config section [web]"), so the settings stay out of it.
const FileName = "herdrweb-settings.json"

// Settings are the UI-owned preferences.
type Settings struct {
	Theme       string  `toml:"theme" json:"theme"`
	Notify      bool    `toml:"notify" json:"notify"`
	Follow      bool    `toml:"follow" json:"follow"`
	Ansi        bool    `toml:"ansi" json:"ansi"`
	DevCaptions bool    `toml:"dev_captions" json:"devCaptions"`
	FontScale   float64 `toml:"font_scale" json:"fontScale"`
}

// Default settings.
func Default() Settings {
	return Settings{Theme: "herdr-dark", Notify: true, Follow: true, Ansi: true, DevCaptions: false, FontScale: 1}
}

// DefaultPath returns ~/.config/herdr/config.toml.
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "herdr", "config.toml")
}

// Load reads the settings file at path, falling back to defaults for missing
// keys and when the file does not exist.
func Load(path string) (Settings, error) {
	s := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return s, err
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return Default(), err
	}
	return s, nil
}

// Save writes the settings file at path.
func Save(path string, s Settings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(path, append(data, '\n'), 0o644)
}

// Migrate moves the [web] table that earlier versions kept in Herdr's
// config.toml (herdrConfig) into the settings file at settingsPath, then deletes
// the table's lines from config.toml, leaving every other line - comments,
// order, formatting - as written. An existing settings file is newer and wins;
// the table is then only removed. A config.toml that does not parse is left
// untouched. Reports whether config.toml was changed.
func Migrate(herdrConfig, settingsPath string) (bool, error) {
	// Write through a symlink (a dotfiles-managed config.toml) rather than
	// replacing the link with a regular file.
	target, err := filepath.EvalSymlinks(herdrConfig)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	data, err := os.ReadFile(target)
	if err != nil {
		return false, err
	}
	rest, found := cutTable(string(data), "web")
	if !found {
		return false, nil
	}
	f := struct {
		Web Settings `toml:"web"`
	}{Default()}
	if err := toml.Unmarshal(data, &f); err != nil {
		return false, fmt.Errorf("parse %s: %w", herdrConfig, err)
	}
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		if err := Save(settingsPath, f.Web); err != nil {
			return false, err
		}
	} else if err != nil {
		return false, err
	}
	info, err := os.Stat(target)
	if err != nil {
		return false, err
	}
	if err := writeAtomic(target, []byte(rest), info.Mode().Perm()); err != nil {
		return false, err
	}
	return true, nil
}

// tableHeader matches a TOML table or array-of-tables header line and captures
// its (bare or dotted) name.
var tableHeader = regexp.MustCompile(`^\s*\[\[?\s*([A-Za-z0-9_\-]+(?:\s*\.\s*[A-Za-z0-9_\-]+)*)\s*\]\]?\s*(?:#.*)?$`)

// cutTable removes table name and its sub-tables - from the header line up to
// the next other table header - from a TOML document, keeping every other line
// byte for byte. When the table ran to the end, the blank lines left before it
// go too.
func cutTable(doc, name string) (string, bool) {
	lines := strings.SplitAfter(doc, "\n")
	start := -1
	for i, line := range lines {
		m := tableHeader.FindStringSubmatch(strings.TrimRight(line, "\r\n"))
		if m == nil {
			continue
		}
		key := strings.Join(strings.Fields(m[1]), "")
		own := key == name || strings.HasPrefix(key, name+".")
		switch {
		case start < 0 && own:
			start = i
		case start >= 0 && !own:
			return strings.Join(lines[:start], "") + strings.Join(lines[i:], ""), true
		}
	}
	if start < 0 {
		return doc, false
	}
	head := lines[:start]
	for len(head) > 0 && strings.TrimSpace(head[len(head)-1]) == "" {
		head = head[:len(head)-1]
	}
	return strings.Join(head, ""), true
}

// writeAtomic replaces path with data via a temporary file and a rename, so a
// crash never leaves a half-written file.
func writeAtomic(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
