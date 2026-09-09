// Package config reads and writes the Herdr Web settings persisted in the
// Herdr config.toml under a [web] table.
package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

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

type file struct {
	Web Settings `toml:"web"`
}

// DefaultPath returns ~/.config/herdr/config.toml.
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "herdr", "config.toml")
}

// Load reads [web] settings from path, falling back to defaults for missing
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
	var f file
	f.Web = s
	if err := toml.Unmarshal(data, &f); err != nil {
		return s, err
	}
	return f.Web, nil
}

// Save writes the [web] table, preserving any other top-level tables already in
// the file by re-encoding the merged document.
func Save(path string, s Settings) error {
	existing := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		_ = toml.Unmarshal(data, &existing)
	}
	existing["web"] = s
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	fh, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := toml.NewEncoder(fh).Encode(existing); err != nil {
		fh.Close()
		return err
	}
	if err := fh.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
