//go:build windows

package service

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"
	"unicode/utf16"

	"github.com/sarathsp06/herdrweb/internal/herdr"
)

// taskName is the Task Scheduler entry that stands in for a systemd/launchd unit.
const taskName = "herdrweb"

// The task runs as the installing user with LogonType S4U ("run whether the
// user is logged on or not", no stored password): the bridge gets no console
// window and no desktop, yet keeps the user's identity, so Herdr's named pipe
// ACL admits it. It starts at logon, never times out (the schtasks default is
// 72 hours) and restarts after a crash.
const taskTemplate = `<?xml version="1.0" encoding="UTF-16"?>
<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <RegistrationInfo>
    <Description>Herdr Web Bridge Daemon</Description>
    <URI>\{{.Name}}</URI>
  </RegistrationInfo>
  <Triggers>
    <LogonTrigger>
      <Enabled>true</Enabled>
      <UserId>{{xml .UserID}}</UserId>
    </LogonTrigger>
  </Triggers>
  <Principals>
    <Principal id="Author">
      <UserId>{{xml .UserID}}</UserId>
      <LogonType>S4U</LogonType>
      <RunLevel>LeastPrivilege</RunLevel>
    </Principal>
  </Principals>
  <Settings>
    <MultipleInstancesPolicy>IgnoreNew</MultipleInstancesPolicy>
    <DisallowStartIfOnBatteries>false</DisallowStartIfOnBatteries>
    <StopIfGoingOnBatteries>false</StopIfGoingOnBatteries>
    <StartWhenAvailable>true</StartWhenAvailable>
    <ExecutionTimeLimit>PT0S</ExecutionTimeLimit>
    <RestartOnFailure>
      <Interval>PT1M</Interval>
      <Count>999</Count>
    </RestartOnFailure>
  </Settings>
  <Actions Context="Author">
    <Exec>
      <Command>{{xml .Command}}</Command>
      <Arguments>{{xml .Arguments}}</Arguments>
    </Exec>
  </Actions>
</Task>
`

// GenerateTaskXML renders the Task Scheduler definition for opts, run as userID.
func GenerateTaskXML(opts ServiceOptions, userID string) (string, error) {
	var args []string
	for _, kv := range [][2]string{
		{"-addr", opts.Addr},
		{"-socket", opts.Socket},
		{"-config", opts.Config},
		{"-log-file", opts.LogPath},
		{"-allow-host", opts.AllowHosts},
	} {
		if kv[1] != "" {
			args = append(args, kv[0], syscall.EscapeArg(kv[1]))
		}
	}
	tmpl, err := template.New("task").Funcs(template.FuncMap{"xml": xmlEscape}).Parse(taskTemplate)
	if err != nil {
		return "", fmt.Errorf("parse task template: %w", err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{
		"Name":      taskName,
		"UserID":    userID,
		"Command":   opts.ExecPath,
		"Arguments": strings.Join(args, " "),
	})
	if err != nil {
		return "", fmt.Errorf("execute task template: %w", err)
	}
	return buf.String(), nil
}

func xmlEscape(s string) (string, error) {
	var buf bytes.Buffer
	err := xml.EscapeText(&buf, []byte(s))
	return buf.String(), err
}

func manageTask(action string, opts ServiceOptions) error {
	switch action {
	case "install":
		return installTask(opts)
	case "uninstall":
		_ = runCmd("schtasks", "/End", "/TN", taskName)
		if err := runCmd("schtasks", "/Delete", "/TN", taskName, "/F"); err != nil {
			return err
		}
		log.Printf("Removed scheduled task %s", taskName)
		return nil
	case "start":
		return runCmdInteractive("schtasks", "/Run", "/TN", taskName)
	case "stop":
		return runCmdInteractive("schtasks", "/End", "/TN", taskName)
	case "status":
		return runCmdInteractive("schtasks", "/Query", "/TN", taskName, "/V", "/FO", "LIST")
	default:
		return fmt.Errorf("unknown service action %q (valid: install, uninstall, start, stop, status)", action)
	}
}

func installTask(opts ServiceOptions) error {
	if opts.ExecPath == "" {
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("get executable path: %w", err)
		}
		opts.ExecPath = execPath
	}
	// No console under S4U: logs must go to a file.
	if opts.LogPath == "" {
		opts.LogPath = filepath.Join(herdr.ConfigDir(), "herdrweb.log")
	}
	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("resolve current user: %w", err)
	}
	content, err := GenerateTaskXML(opts, u.Username)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp("", "herdrweb-task-*.xml")
	if err != nil {
		return fmt.Errorf("create task file: %w", err)
	}
	defer os.Remove(f.Name())
	// schtasks /XML expects UTF-16 matching the declaration.
	units := utf16.Encode([]rune("\ufeff" + content))
	if err := binary.Write(f, binary.LittleEndian, units); err != nil {
		f.Close()
		return fmt.Errorf("write task file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("write task file: %w", err)
	}
	if err := runCmd("schtasks", "/Create", "/TN", taskName, "/XML", f.Name(), "/F"); err != nil {
		return fmt.Errorf("%w (if access was denied, retry from an elevated terminal)", err)
	}
	log.Printf("Installed scheduled task %s (runs at logon as %s, logs to %s)", taskName, u.Username, opts.LogPath)
	log.Println("Start it now with: herdrweb -service start")
	return nil
}
