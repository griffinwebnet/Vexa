package services

import (
	"fmt"
	exec "os/exec"
	"strings"

	sambaExec "github.com/vexa/api/exec"
	"github.com/vexa/api/models"
)

// ComputerService handles computer-related business logic
type ComputerService struct {
	sambaTool *sambaExec.SambaTool
}

// NewComputerService creates a new ComputerService instance
func NewComputerService() *ComputerService {
	return &ComputerService{
		sambaTool: sambaExec.NewSambaTool(),
	}
}

// ListComputers returns all computers/devices in the domain with connection status
func (s *ComputerService) ListComputers() ([]models.Computer, error) {

	output, err := s.sambaTool.Run("computer", "list")
	if err != nil {
		return nil, fmt.Errorf("failed to list computers: %s", output)
	}

	computerNames := strings.Split(strings.TrimSpace(output), "\n")
	computers := make([]models.Computer, 0)

	// Check if Headscale is enabled
	headscaleEnabled := s.isHeadscaleEnabled()

	for _, name := range computerNames {
		if name == "" {
			continue
		}

		computer := models.Computer{
			Name:           strings.TrimSuffix(name, "$"),
			DNSName:        name,
			Online:         false,
			ConnectionType: "offline",
		}

		// Try to ping the computer to check if online
		if s.pingComputer(computer.Name) {
			computer.Online = true
			computer.ConnectionType = "local"

			// Try to get IP address
			if ip := s.getComputerIP(computer.Name); ip != "" {
				computer.IPAddress = ip
			}
		}

		// If Headscale is enabled, check overlay connection
		if headscaleEnabled {
			if overlayIP := s.getHeadscaleIP(computer.Name); overlayIP != "" {
				computer.OverlayIP = overlayIP
				computer.Online = true
				computer.ConnectionType = "overlay"
			}
		}

		computers = append(computers, computer)
	}

	return computers, nil
}

// GetComputer returns details for a specific computer
func (s *ComputerService) GetComputer(computerName string) (*models.Computer, error) {
	// TODO: Get detailed computer info from Samba
	computer := &models.Computer{
		Name: computerName,
	}
	return computer, nil
}

// DeleteComputer removes a computer from the domain
func (s *ComputerService) DeleteComputer(computerName string) error {
	output, err := s.sambaTool.Run("computer", "delete", computerName)
	if err != nil {
		return fmt.Errorf("failed to delete computer: %s", output)
	}
	return nil
}

// Helper functions

func (s *ComputerService) isHeadscaleEnabled() bool {
	cmd := exec.Command("systemctl", "is-active", "headscale")
	return cmd.Run() == nil
}

func (s *ComputerService) pingComputer(hostname string) bool {
	cmd := exec.Command("ping", "-c", "1", "-W", "1", hostname)
	return cmd.Run() == nil
}

func (s *ComputerService) getComputerIP(hostname string) string {
	cmd := exec.Command("host", hostname)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse "hostname has address X.X.X.X"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "has address") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				return parts[3]
			}
		}
	}
	return ""
}

func (s *ComputerService) getHeadscaleIP(hostname string) string {
	// Query Headscale for this node
	cmd := exec.Command("headscale", "nodes", "list", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// TODO: Parse JSON to find node by hostname and return its IP
	// For now, return empty - this requires proper JSON parsing
	_ = output
	return ""
}
