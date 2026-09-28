//go:build windows

package daemon

import "syscall"

// Windows process-creation flags (winbase.h): run the child without a console
// and outside the parent's Ctrl+C group, the equivalent of Setsid.
const (
	detachedProcess       = 0x00000008
	createNewProcessGroup = 0x00000200
)

func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: detachedProcess | createNewProcessGroup}
}
