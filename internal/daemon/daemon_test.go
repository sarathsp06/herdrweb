package daemon

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestSetupLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "logs", "test.log")

	f, err := SetupLogging(logPath)
	if err != nil {
		t.Fatalf("SetupLogging failed: %v", err)
	}
	if f == nil {
		t.Fatalf("expected log file handle, got nil")
	}
	defer f.Close()

	log.Println("test log entry")
	_ = f.Sync()

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file failed: %v", err)
	}

	if !strings.Contains(string(data), "test log entry") {
		t.Errorf("expected log file to contain 'test log entry', got: %s", string(data))
	}
}

func TestWritePIDFile(t *testing.T) {
	tmpDir := t.TempDir()
	pidPath := filepath.Join(tmpDir, "pids", "test.pid")

	cleanup, err := WritePIDFile(pidPath)
	if err != nil {
		t.Fatalf("WritePIDFile failed: %v", err)
	}

	data, err := os.ReadFile(pidPath)
	if err != nil {
		t.Fatalf("read PID file failed: %v", err)
	}

	pidStr := strings.TrimSpace(string(data))
	pidVal, err := strconv.Atoi(pidStr)
	if err != nil {
		t.Fatalf("invalid PID in file: %s", pidStr)
	}

	if pidVal != os.Getpid() {
		t.Errorf("expected PID %d, got %d", os.Getpid(), pidVal)
	}

	cleanup()

	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Errorf("expected PID file to be removed after cleanup, but it still exists")
	}
}

func TestDaemonizeChildMode(t *testing.T) {
	t.Setenv(EnvHerdrDaemon, "1")

	isParent, pid, err := Daemonize("")
	if err != nil {
		t.Fatalf("Daemonize failed: %v", err)
	}

	if isParent {
		t.Errorf("expected isParent to be false in child mode")
	}

	if pid != os.Getpid() {
		t.Errorf("expected pid %d, got %d", os.Getpid(), pid)
	}
}
