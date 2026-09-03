package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingReturnsDefault(t *testing.T) {
	s, err := Load(filepath.Join(t.TempDir(), "nope.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if s != Default() {
		t.Fatalf("want default, got %+v", s)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.toml")
	want := Settings{Theme: "gruvbox", Notify: false, Follow: true, Ansi: false, DevCaptions: true}
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

func TestSavePreservesOtherTables(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(p, []byte("[keys]\nprefix = \"ctrl+b\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Save(p, Settings{Theme: "ash"}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(p)
	if !contains(string(data), "prefix") || !contains(string(data), "ash") {
		t.Fatalf("lost data on save: %s", data)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
