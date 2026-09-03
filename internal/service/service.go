package service

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

// ServiceOptions holds configurations for service file generation.
type ServiceOptions struct {
	ExecPath string
	Addr     string
	Socket   string
	Config   string
	LogPath  string
}

const systemdTemplate = `[Unit]
Description=Herdr Web Bridge Daemon
Documentation=https://github.com/sarathsp06/herdrweb
After=network.target

[Service]
Type=simple
ExecStart={{.ExecPath}}{{if .Addr}} -addr {{.Addr}}{{end}}{{if .Socket}} -socket {{.Socket}}{{end}}{{if .Config}} -config {{.Config}}{{end}}{{if .LogPath}} -log-file {{.LogPath}}{{end}}
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`

const launchdTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.herdrweb.herdr-bridge</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.ExecPath}}</string>
{{if .Addr}}        <string>-addr</string>
        <string>{{.Addr}}</string>
{{end}}{{if .Socket}}        <string>-socket</string>
        <string>{{.Socket}}</string>
{{end}}{{if .Config}}        <string>-config</string>
        <string>{{.Config}}</string>
{{end}}{{if .LogPath}}        <string>-log-file</string>
        <string>{{.LogPath}}</string>
{{end}}    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
{{if .LogPath}}    <key>StandardOutPath</key>
    <string>{{.LogPath}}</string>
    <key>StandardErrorPath</key>
    <string>{{.LogPath}}</string>
{{end}}</dict>
</plist>
`

// GenerateSystemdUnit generates a systemd service unit file content.
func GenerateSystemdUnit(opts ServiceOptions) (string, error) {
	tmpl, err := template.New("systemd").Parse(systemdTemplate)
	if err != nil {
		return "", fmt.Errorf("parse systemd template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, opts); err != nil {
		return "", fmt.Errorf("execute systemd template: %w", err)
	}
	return buf.String(), nil
}

// GenerateLaunchdPlist generates a macOS launchd plist file content.
func GenerateLaunchdPlist(opts ServiceOptions) (string, error) {
	tmpl, err := template.New("launchd").Parse(launchdTemplate)
	if err != nil {
		return "", fmt.Errorf("parse launchd template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, opts); err != nil {
		return "", fmt.Errorf("execute launchd template: %w", err)
	}
	return buf.String(), nil
}

// Manage handles the service action: install, uninstall, start, stop, status.
func Manage(action string, opts ServiceOptions) error {
	action = strings.ToLower(strings.TrimSpace(action))
	switch action {
	case "install":
		return installService(opts)
	case "uninstall":
		return uninstallService()
	case "start":
		return startService()
	case "stop":
		return stopService()
	case "status":
		return statusService()
	default:
		return fmt.Errorf("unknown service action %q (valid: install, uninstall, start, stop, status)", action)
	}
}

func installService(opts ServiceOptions) error {
	if opts.ExecPath == "" {
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("get executable path: %w", err)
		}
		opts.ExecPath = execPath
	}

	if opts.LogPath == "" {
		home, _ := os.UserHomeDir()
		opts.LogPath = filepath.Join(home, ".config", "herdr", "herdr-bridge.log")
	}

	switch runtime.GOOS {
	case "linux":
		unitContent, err := GenerateSystemdUnit(opts)
		if err != nil {
			return err
		}
		unitPath, isUser, err := getSystemdUnitPath()
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(unitPath), 0755); err != nil {
			return fmt.Errorf("create systemd directory: %w", err)
		}
		if err := os.WriteFile(unitPath, []byte(unitContent), 0644); err != nil {
			return fmt.Errorf("write systemd unit: %w", err)
		}

		log.Printf("Installed systemd unit at %s", unitPath)

		cmdArgs := []string{}
		if isUser {
			cmdArgs = []string{"--user"}
		}

		runCmd("systemctl", append(cmdArgs, "daemon-reload")...)
		runCmd("systemctl", append(cmdArgs, "enable", "herdr-bridge")...)
		log.Println("Service enabled. Start with: herdr-bridge -service start (or systemctl --user start herdr-bridge)")
		return nil

	case "darwin":
		plistContent, err := GenerateLaunchdPlist(opts)
		if err != nil {
			return err
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get user home dir: %w", err)
		}
		plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.herdrweb.herdr-bridge.plist")

		if err := os.MkdirAll(filepath.Dir(plistPath), 0755); err != nil {
			return fmt.Errorf("create launchd directory: %w", err)
		}
		if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
			return fmt.Errorf("write launchd plist: %w", err)
		}

		log.Printf("Installed launchd plist at %s", plistPath)
		runCmd("launchctl", "load", plistPath)
		log.Println("Service loaded. Manage with: herdr-bridge -service start|stop|status")
		return nil

	default:
		return fmt.Errorf("service installation is not supported on OS %q", runtime.GOOS)
	}
}

func uninstallService() error {
	switch runtime.GOOS {
	case "linux":
		unitPath, isUser, err := getSystemdUnitPath()
		if err != nil {
			return err
		}
		cmdArgs := []string{}
		if isUser {
			cmdArgs = []string{"--user"}
		}
		_ = runCmd("systemctl", append(cmdArgs, "stop", "herdr-bridge")...)
		_ = runCmd("systemctl", append(cmdArgs, "disable", "herdr-bridge")...)

		if err := os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
			log.Printf("warning: could not remove %s: %v", unitPath, err)
		} else {
			log.Printf("Removed %s", unitPath)
		}
		_ = runCmd("systemctl", append(cmdArgs, "daemon-reload")...)
		return nil

	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get user home dir: %w", err)
		}
		plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.herdrweb.herdr-bridge.plist")
		_ = runCmd("launchctl", "unload", plistPath)

		if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
			log.Printf("warning: could not remove %s: %v", plistPath, err)
		} else {
			log.Printf("Removed %s", plistPath)
		}
		return nil

	default:
		return fmt.Errorf("service uninstallation is not supported on OS %q", runtime.GOOS)
	}
}

func startService() error {
	switch runtime.GOOS {
	case "linux":
		_, isUser, _ := getSystemdUnitPath()
		cmdArgs := []string{}
		if isUser {
			cmdArgs = []string{"--user"}
		}
		return runCmdInteractive("systemctl", append(cmdArgs, "start", "herdr-bridge")...)
	case "darwin":
		return runCmdInteractive("launchctl", "start", "com.herdrweb.herdr-bridge")
	default:
		return fmt.Errorf("service start is not supported on OS %q", runtime.GOOS)
	}
}

func stopService() error {
	switch runtime.GOOS {
	case "linux":
		_, isUser, _ := getSystemdUnitPath()
		cmdArgs := []string{}
		if isUser {
			cmdArgs = []string{"--user"}
		}
		return runCmdInteractive("systemctl", append(cmdArgs, "stop", "herdr-bridge")...)
	case "darwin":
		return runCmdInteractive("launchctl", "stop", "com.herdrweb.herdr-bridge")
	default:
		return fmt.Errorf("service stop is not supported on OS %q", runtime.GOOS)
	}
}

func statusService() error {
	switch runtime.GOOS {
	case "linux":
		_, isUser, _ := getSystemdUnitPath()
		cmdArgs := []string{}
		if isUser {
			cmdArgs = []string{"--user"}
		}
		return runCmdInteractive("systemctl", append(cmdArgs, "status", "herdr-bridge")...)
	case "darwin":
		return runCmdInteractive("launchctl", "list", "com.herdrweb.herdr-bridge")
	default:
		return fmt.Errorf("service status is not supported on OS %q", runtime.GOOS)
	}
}

func getSystemdUnitPath() (string, bool, error) {
	if os.Geteuid() == 0 {
		return "/etc/systemd/system/herdr-bridge.service", false, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", true, fmt.Errorf("get user home dir: %w", err)
	}
	return filepath.Join(home, ".config", "systemd", "user", "herdr-bridge.service"), true, nil
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v failed: %w (%s)", name, args, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func runCmdInteractive(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
