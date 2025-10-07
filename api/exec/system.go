package exec

import (
	"runtime"

	"github.com/griffinwebnet/vexa/api/utils"
)

// System provides an interface for system-level operations
type System struct{}

// NewSystem creates a new System instance
func NewSystem() *System {
	return &System{}
}

// CheckCommandExists checks if a command exists in the system PATH
func (s *System) CheckCommandExists(command string) bool {
	cmd, err := utils.SafeCommand("which", command)
	if err != nil {
		return false
	}
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
	cmd, cmdErr := utils.SafeCommand("systemctl", "is-active", serviceName)
	if cmdErr != nil {
		return false, cmdErr
	}
	err := cmd.Run()
	return err == nil, err
}

// EnableAndStartService enables and starts a system service
func (s *System) EnableAndStartService(serviceName string) error {
	cmd, err := utils.SafeCommand("systemctl", "enable", "--now", serviceName)
	if err != nil {
		return err
	}
	return cmd.Run()
}

// StopService stops a system service
func (s *System) StopService(serviceName string) error {
	cmd, err := utils.SafeCommand("systemctl", "stop", serviceName)
	if err != nil {
		return err
	}
	return cmd.Run()
}

// RemoveFile removes a file from the filesystem
func (s *System) RemoveFile(path string) error {
	cmd, err := utils.SafeCommand("rm", "-f", path)
	if err != nil {
		return err
	}
	return cmd.Run()
}

// RunCommand executes a command with arguments
func (s *System) RunCommand(name string, args ...string) error {
	cmd, err := utils.SafeCommand(name, args...)
	if err != nil {
		return err
	}
	return cmd.Run()
}

// FileExists checks if a file exists on the filesystem
func (s *System) FileExists(path string) bool {
	cmd, err := utils.SafeCommand("test", "-f", path)
	if err != nil {
		return false
	}
	return cmd.Run() == nil
}
