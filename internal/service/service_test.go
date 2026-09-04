package service

import (
	"strings"
	"testing"
)

func TestGenerateSystemdUnit(t *testing.T) {
	opts := ServiceOptions{
		ExecPath: "/usr/local/bin/herdr-bridge",
		Addr:     "127.0.0.1:7331",
		Socket:   "/home/user/.config/herdr/herdr.sock",
		Config:   "/home/user/.config/herdr/config.toml",
		LogPath:  "/home/user/.config/herdr/herdr-bridge.log",
	}

	unit, err := GenerateSystemdUnit(opts)
	if err != nil {
		t.Fatalf("GenerateSystemdUnit failed: %v", err)
	}

	expectedParts := []string{
		"[Unit]",
		"Description=Herdr Web Bridge Daemon",
		"ExecStart=/usr/local/bin/herdr-bridge -addr 127.0.0.1:7331 -socket /home/user/.config/herdr/herdr.sock -config /home/user/.config/herdr/config.toml -log-file /home/user/.config/herdr/herdr-bridge.log",
		"Restart=on-failure",
		"WantedBy=default.target",
	}

	for _, part := range expectedParts {
		if !strings.Contains(unit, part) {
			t.Errorf("expected unit file to contain %q, unit:\n%s", part, unit)
		}
	}
}

func TestGenerateLaunchdPlist(t *testing.T) {
	opts := ServiceOptions{
		ExecPath: "/usr/local/bin/herdr-bridge",
		Addr:     "127.0.0.1:7331",
		LogPath:  "/tmp/herdr-bridge.log",
	}

	plist, err := GenerateLaunchdPlist(opts)
	if err != nil {
		t.Fatalf("GenerateLaunchdPlist failed: %v", err)
	}

	expectedParts := []string{
		"<key>Label</key>",
		"<string>com.herdrweb.herdr-bridge</string>",
		"<string>/usr/local/bin/herdr-bridge</string>",
		"<string>-addr</string>",
		"<string>127.0.0.1:7331</string>",
		"<key>StandardOutPath</key>",
		"<string>/tmp/herdr-bridge.log</string>",
	}

	for _, part := range expectedParts {
		if !strings.Contains(plist, part) {
			t.Errorf("expected plist file to contain %q, plist:\n%s", part, plist)
		}
	}
}

func TestManageInvalidAction(t *testing.T) {
	err := Manage("invalid-action", ServiceOptions{})
	if err == nil {
		t.Fatalf("expected error for invalid service action, got nil")
	}
	if !strings.Contains(err.Error(), "unknown service action") {
		t.Errorf("expected error message to contain 'unknown service action', got: %v", err)
	}
}
