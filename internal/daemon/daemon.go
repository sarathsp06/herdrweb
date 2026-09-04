package daemon

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const EnvHerdrDaemon = "HERDR_DAEMON"

// SetupLogging redirects log output to logFilePath if specified.
// It creates missing parent directories automatically.
func SetupLogging(logFilePath string) (*os.File, error) {
	if strings.TrimSpace(logFilePath) == "" {
		return nil, nil
	}

	dir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create log dir %s: %w", dir, err)
	}

	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file %s: %w", logFilePath, err)
	}

	log.SetOutput(f)
	return f, nil
}

// WritePIDFile writes current process ID to pidFilePath.
// Returns a cleanup function that deletes the PID file on exit.
func WritePIDFile(pidFilePath string) (func(), error) {
	if strings.TrimSpace(pidFilePath) == "" {
		return func() {}, nil
	}

	dir := filepath.Dir(pidFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return func() {}, fmt.Errorf("create pid dir %s: %w", dir, err)
	}

	pid := os.Getpid()
	content := []byte(strconv.Itoa(pid) + "\n")
	if err := os.WriteFile(pidFilePath, content, 0644); err != nil {
		return func() {}, fmt.Errorf("write pid file %s: %w", pidFilePath, err)
	}

	cleanup := func() {
		_ = os.Remove(pidFilePath)
	}
	return cleanup, nil
}

// Daemonize background-spawns the binary if HERDR_DAEMON is not set to "1".
// If it is the parent process, it starts the child and returns (true, pid, nil).
// If it is already the daemonized child, it returns (false, pid, nil).
func Daemonize(logFilePath string) (bool, int, error) {
	if os.Getenv(EnvHerdrDaemon) == "1" {
		return false, os.Getpid(), nil
	}

	execPath, err := os.Executable()
	if err != nil {
		return true, 0, fmt.Errorf("get executable path: %w", err)
	}

	args := os.Args[1:]
	cmd := exec.Command(execPath, args...)
	cmd.Env = append(os.Environ(), EnvHerdrDaemon+"=1")
	cmd.SysProcAttr = sysProcAttr()
	cmd.Stdin = nil

	if strings.TrimSpace(logFilePath) != "" {
		dir := filepath.Dir(logFilePath)
		if err := os.MkdirAll(dir, 0755); err == nil {
			if logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
				cmd.Stdout = logFile
				cmd.Stderr = logFile
			}
		}
	}

	if err := cmd.Start(); err != nil {
		return true, 0, fmt.Errorf("start daemon process: %w", err)
	}

	return true, cmd.Process.Pid, nil
}
