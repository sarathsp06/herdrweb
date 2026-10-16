package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingReturnsDefault(t *testing.T) {
	s, err := Load(filepath.Join(t.TempDir(), FileName))
	if err != nil {
		t.Fatal(err)
	}
	if s != Default() {
		t.Fatalf("want default, got %+v", s)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), FileName)
	want := Settings{Theme: "gruvbox", Notify: false, Follow: true, Ansi: false, DevCaptions: true, FontScale: 1.15}
	if err := Save(p, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("round-trip mismatch: want %+v got %+v", want, got)
	}
}

func TestLoadFillsMissingKeysWithDefaults(t *testing.T) {
	p := filepath.Join(t.TempDir(), FileName)
	write(t, p, `{"theme":"paper","notify":false}`)
	got, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	want := Default()
	want.Theme, want.Notify = "paper", false
	if got != want {
		t.Fatalf("want %+v got %+v", want, got)
	}
}

// migrate runs Migrate on a config.toml holding doc and returns what is left of
// config.toml, the settings file's path and whether Migrate reported a change.
func migrate(t *testing.T, doc string) (string, string, bool) {
	t.Helper()
	dir := t.TempDir()
	cfg, settings := filepath.Join(dir, "config.toml"), filepath.Join(dir, FileName)
	write(t, cfg, doc)
	moved, err := Migrate(cfg, settings)
	if err != nil {
		t.Fatal(err)
	}
	return read(t, cfg), settings, moved
}

func TestMigrateMovesWebTableAndKeepsTheRestVerbatim(t *testing.T) {
	doc := "# my herdr config\n" +
		"onboarding = false\n" +
		"\n" +
		"[keys]\n" +
		"  remote_image_paste = \"f8\"   # F8 pastes\n" +
		"\n" +
		"[web]\n" +
		"  theme = \"gruvbox\"\n" +
		"  notify = false\n" +
		"  font_scale = 0.9\n" +
		"\n" +
		"[web.extra]\n" +
		"  x = 1\n" +
		"\n" +
		"[[keys.command]]\n" +
		"  key = \"prefix+e\"\n"
	rest, settings, moved := migrate(t, doc)
	want := "# my herdr config\n" +
		"onboarding = false\n" +
		"\n" +
		"[keys]\n" +
		"  remote_image_paste = \"f8\"   # F8 pastes\n" +
		"\n" +
		"[[keys.command]]\n" +
		"  key = \"prefix+e\"\n"
	if !moved || rest != want {
		t.Fatalf("moved=%v, config.toml is now:\n%q\nwant:\n%q", moved, rest, want)
	}
	got, err := Load(settings)
	if err != nil {
		t.Fatal(err)
	}
	wantS := Default()
	wantS.Theme, wantS.Notify, wantS.FontScale = "gruvbox", false, 0.9
	if got != wantS {
		t.Fatalf("settings: want %+v got %+v", wantS, got)
	}
}

func TestMigrateWebTableAtEndOfCRLFFile(t *testing.T) {
	rest, _, moved := migrate(t, "[ui]\r\n  toast = \"system\"\r\n\r\n[web]\r\n  theme = \"paper\"\r\n")
	if want := "[ui]\r\n  toast = \"system\"\r\n"; !moved || rest != want {
		t.Fatalf("moved=%v, config.toml is now %q, want %q", moved, rest, want)
	}
}

func TestMigrateKeepsNewerSettingsFile(t *testing.T) {
	dir := t.TempDir()
	cfg, settings := filepath.Join(dir, "config.toml"), filepath.Join(dir, FileName)
	write(t, cfg, "[web]\ntheme = \"gruvbox\"\n")
	if err := Save(settings, Settings{Theme: "paper"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Migrate(cfg, settings); err != nil {
		t.Fatal(err)
	}
	if s, _ := Load(settings); s.Theme != "paper" {
		t.Fatalf("existing settings overwritten: %+v", s)
	}
	if rest := read(t, cfg); rest != "" {
		t.Fatalf("[web] left in config.toml: %q", rest)
	}
}

func TestMigrateLeavesConfigAloneWithoutWebTable(t *testing.T) {
	doc := "[keys]\nprefix = \"ctrl+b\"\n"
	rest, settings, moved := migrate(t, doc)
	if moved || rest != doc {
		t.Fatalf("moved=%v, config.toml changed to %q", moved, rest)
	}
	if _, err := os.Stat(settings); !os.IsNotExist(err) {
		t.Fatalf("settings file created without a [web] table: %v", err)
	}
}

func TestMigrateLeavesUnparsableConfigAlone(t *testing.T) {
	dir := t.TempDir()
	cfg, settings := filepath.Join(dir, "config.toml"), filepath.Join(dir, FileName)
	doc := "[keys]\nprefix = \n\n[web]\ntheme = \"paper\"\n"
	write(t, cfg, doc)
	if _, err := Migrate(cfg, settings); err == nil {
		t.Fatal("want a parse error")
	}
	if rest := read(t, cfg); rest != doc {
		t.Fatalf("unparsable config.toml changed to %q", rest)
	}
	if _, err := os.Stat(settings); !os.IsNotExist(err) {
		t.Fatalf("settings file written from an unparsable config: %v", err)
	}
}

func TestMigrateWritesThroughSymlink(t *testing.T) {
	dir := t.TempDir()
	real, link := filepath.Join(dir, "dotfiles-config.toml"), filepath.Join(dir, "config.toml")
	write(t, real, "[ui]\nx = 1\n\n[web]\ntheme = \"paper\"\n")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := Migrate(link, filepath.Join(dir, FileName)); err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Lstat(link); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("config.toml is no longer a symlink: %v", err)
	}
	if rest := read(t, real); rest != "[ui]\nx = 1\n" {
		t.Fatalf("link target is now %q", rest)
	}
}

func write(t *testing.T, path, s string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
