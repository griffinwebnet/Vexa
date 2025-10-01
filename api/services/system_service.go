package services

import (
	"github.com/vexa/api/exec"
	"github.com/vexa/api/models"
)

// SystemService handles system-related business logic
type SystemService struct {
	system *exec.System
}

// NewSystemService creates a new SystemService instance
func NewSystemService() *SystemService {
	return &SystemService{
		system: exec.NewSystem(),
	}
}

// GetSystemStatus returns system information
func (s *SystemService) GetSystemStatus() *models.SystemStatusResponse {
	sambaInstalled := s.system.CheckCommandExists("samba-tool")
	bindInstalled := s.system.CheckCommandExists("named")

	return &models.SystemStatusResponse{
		OS:             s.system.GetOS(),
		Architecture:   s.system.GetArchitecture(),
		SambaInstalled: sambaInstalled,
		BindInstalled:  bindInstalled,
		APIVersion:     "1.0.0",
	}
}

// GetServiceStatus checks the status of a system service
func (s *SystemService) GetServiceStatus(serviceName string) (*models.ServiceStatus, error) {
	active, err := s.system.ServiceStatus(serviceName)
	if err != nil {
		return nil, err
	}

	status := "inactive"
	if active {
		status = "active"
	}

	return &models.ServiceStatus{
		Name:   serviceName,
		Active: active,
		Status: status,
	}, nil
}
