//go:build windows

package herdr

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Microsoft/go-winio"
)

// ConfigDir is Herdr's per-user directory: %APPDATA%\herdr on Windows.
func ConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, "AppData", "Roaming")
	}
	return filepath.Join(dir, "herdr")
}

// pipeName maps a Herdr socket path to the named pipe Herdr actually serves.
// On Windows the herdr.sock file only records the server PID; the endpoint is
// the named pipe \\.\pipe\<full socket path>.
func pipeName(path string) string {
	if strings.HasPrefix(path, `\\.\pipe\`) {
		return path
	}
	return `\\.\pipe\` + path
}

func dialSocket(ctx context.Context, path string, timeout time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return winio.DialPipeContext(ctx, pipeName(path))
}

// Listen serves a Herdr-style socket at path; tests use it for fake servers.
func Listen(path string) (net.Listener, error) {
	return winio.ListenPipe(pipeName(path), nil)
}
