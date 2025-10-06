package services

import (
	"encoding/json"
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

	// Get domain controller hostname
	dcHostname := s.getDomainControllerHostname()

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

				// Generate overlay URL (hostname.domain.mesh)
				domainService := NewDomainService()
				domainStatus, _ := domainService.GetDomainStatus()
				if domainStatus != nil && domainStatus.Domain != "" && domainStatus.Domain != "PROVISIONED" {
					computer.OverlayURL = fmt.Sprintf("%s.%s.mesh", computer.Name, domainStatus.Domain)
				}
			}
		}

		computers = append(computers, computer)
	}

	// Add domain controller if it's not already in the list and Headscale is enabled
	if headscaleEnabled && dcHostname != "" {
		dcExists := false
		for _, comp := range computers {
			if comp.Name == dcHostname || comp.Name == strings.TrimSuffix(dcHostname, "$") {
				dcExists = true
				break
			}
		}

		if !dcExists {
			dcComputer := models.Computer{
				Name:           strings.TrimSuffix(dcHostname, "$"),
				DNSName:        dcHostname,
				Online:         false,
				ConnectionType: "offline",
			}

			// Check if DC is online via overlay
			if overlayIP := s.getHeadscaleIP(dcHostname); overlayIP != "" {
				dcComputer.OverlayIP = overlayIP
				dcComputer.Online = true
				dcComputer.ConnectionType = "overlay"

				// Generate overlay URL for DC
				domainService := NewDomainService()
				domainStatus, _ := domainService.GetDomainStatus()
				if domainStatus != nil && domainStatus.Domain != "" && domainStatus.Domain != "PROVISIONED" {
					dcComputer.OverlayURL = fmt.Sprintf("%s.%s.mesh", dcComputer.Name, domainStatus.Domain)
				}
			}

			computers = append(computers, dcComputer)
		}
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

	// Parse JSON to find node by hostname and return its IP
	var nodes []map[string]interface{}
	if err := json.Unmarshal(output, &nodes); err != nil {
		return ""
	}

	for _, node := range nodes {
		if nodeHostname, ok := node["name"].(string); ok {
			// Check if hostname matches (with or without $ suffix)
			if nodeHostname == hostname || nodeHostname == hostname+"$" ||
				strings.TrimSuffix(nodeHostname, "$") == hostname {
				if ipv4, ok := node["ipv4"].(string); ok && ipv4 != "" {
					return ipv4
				}
				// Try alternative IP field names
				if ip, ok := node["ip"].(string); ok && ip != "" {
					return ip
				}
			}
		}
	}
	return ""
}

func (s *ComputerService) getDomainControllerHostname() string {
	// Get hostname of the current system
	cmd := exec.Command("hostname")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
