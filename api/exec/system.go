package exec

import (
	"os/exec"
	"runtime"
)

// System provides an interface for system-level operations
type System struct{}

// NewSystem creates a new System instance
func NewSystem() *System {
	return &System{}
}

// CheckCommandExists checks if a command exists in the system PATH
func (s *System) CheckCommandExists(command string) bool {
	cmd := exec.Command("which", command)
	return cmd.Run() == nil
}

// GetOS returns the operating system
func (s *System) GetOS() string {
	return runtime.GOOS
}

// GetArchitecture returns the system architecture
func (s *System) GetArchitecture() string {
	return runtime.GOARCH
}

// ServiceStatus checks the status of a system service
func (s *System) ServiceStatus(serviceName string) (bool, error) {
	cmd := exec.Command("systemctl", "is-active", serviceName)
	err := cmd.Run()
	return err == nil, err
}

// EnableAndStartService enables and starts a system service
func (s *System) EnableAndStartService(serviceName string) error {
	cmd := exec.Command("systemctl", "enable", "--now", serviceName)
	return cmd.Run()
}

// StopService stops a system service
func (s *System) StopService(serviceName string) error {
	cmd := exec.Command("systemctl", "stop", serviceName)
	return cmd.Run()
}

// RemoveFile removes a file from the filesystem
func (s *System) RemoveFile(path string) error {
	cmd := exec.Command("rm", "-f", path)
	return cmd.Run()
}

// RunCommand executes a command with arguments
func (s *System) RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}
