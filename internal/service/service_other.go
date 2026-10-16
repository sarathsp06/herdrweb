//go:build !windows

package service

// manageTask is the Windows Task Scheduler backend; Manage only calls it on Windows.
func manageTask(string, ServiceOptions) error {
	panic("service: manageTask called on a non-Windows platform")
}
