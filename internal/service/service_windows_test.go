//go:build windows

package service

import (
	"encoding/xml"
	"io"
	"strings"
	"testing"
)

// TestGenerateTaskXML proves the scheduled task starts the bridge with the
// exact arguments it was installed with - paths with spaces survive Windows
// command-line quoting and XML escaping - and cannot be killed by Task
// Scheduler's default 72-hour execution limit.
func TestGenerateTaskXML(t *testing.T) {
	opts := ServiceOptions{
		ExecPath:   `C:\Program Files\herdrweb\herdrweb.exe`,
		Addr:       "127.0.0.1:7331",
		Socket:     `C:\Users\A & B\AppData\Roaming\herdr\herdr.sock`,
		Config:     `C:\Users\A & B\AppData\Roaming\herdr\config.toml`,
		LogPath:    `C:\Users\A & B\AppData\Roaming\herdr\herdrweb.log`,
		AllowHosts: "herdr.example.com",
	}
	out, err := GenerateTaskXML(opts, `DESKTOP\a&b`)
	if err != nil {
		t.Fatal(err)
	}

	var task struct {
		UserID    string `xml:"Principals>Principal>UserId"`
		LogonType string `xml:"Principals>Principal>LogonType"`
		Limit     string `xml:"Settings>ExecutionTimeLimit"`
		Command   string `xml:"Actions>Exec>Command"`
		Arguments string `xml:"Actions>Exec>Arguments"`
	}
	// encoding/xml rejects a UTF-16 declaration on a Go string; the body is what matters.
	d := xml.NewDecoder(strings.NewReader(out))
	d.CharsetReader = func(_ string, r io.Reader) (io.Reader, error) { return r, nil }
	if err := d.Decode(&task); err != nil {
		t.Fatalf("task XML does not parse: %v\n%s", err, out)
	}

	if task.Command != opts.ExecPath {
		t.Errorf("Command = %q, want %q", task.Command, opts.ExecPath)
	}
	wantArgs := `-addr 127.0.0.1:7331 -socket "C:\Users\A & B\AppData\Roaming\herdr\herdr.sock" ` +
		`-config "C:\Users\A & B\AppData\Roaming\herdr\config.toml" ` +
		`-log-file "C:\Users\A & B\AppData\Roaming\herdr\herdrweb.log" -allow-host herdr.example.com`
	if task.Arguments != wantArgs {
		t.Errorf("Arguments =\n  %s\nwant\n  %s", task.Arguments, wantArgs)
	}
	if task.UserID != `DESKTOP\a&b` || task.LogonType != "S4U" {
		t.Errorf("principal = %q/%q, want DESKTOP\\a&b/S4U", task.UserID, task.LogonType)
	}
	if task.Limit != "PT0S" {
		t.Errorf("ExecutionTimeLimit = %q, want PT0S (no limit)", task.Limit)
	}
}
