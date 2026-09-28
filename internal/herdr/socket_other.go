//go:build !windows

package herdr

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"time"
)

// ConfigDir is Herdr's per-user directory: ~/.config/herdr.
func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "herdr")
}

func dialSocket(ctx context.Context, path string, timeout time.Duration) (net.Conn, error) {
	d := net.Dialer{Timeout: timeout}
	return d.DialContext(ctx, "unix", path)
}

// Listen serves a Herdr-style socket at path; tests use it for fake servers.
func Listen(path string) (net.Listener, error) {
	return net.Listen("unix", path)
}
