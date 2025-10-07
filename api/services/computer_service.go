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

// GetMachineDetails returns detailed information about a machine from Headscale
func (s *ComputerService) GetMachineDetails(machineId string) (map[string]interface{}, error) {
	// Query Headscale for machine details
	cmd := exec.Command("headscale", "nodes", "list", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to query headscale: %v", err)
	}

	// Parse JSON to find the specific machine
	var nodes []map[string]interface{}
	if err := json.Unmarshal(output, &nodes); err != nil {
		return nil, fmt.Errorf("failed to parse headscale output: %v", err)
	}

	// Find the machine by ID or name
	for _, node := range nodes {
		if nodeID, ok := node["id"].(string); ok && nodeID == machineId {
			return s.formatMachineDetails(node)
		}
		if nodeName, ok := node["name"].(string); ok && nodeName == machineId {
			return s.formatMachineDetails(node)
		}
	}

	return nil, fmt.Errorf("machine not found")
}

// formatMachineDetails formats Headscale node data into our machine details format
func (s *ComputerService) formatMachineDetails(node map[string]interface{}) (map[string]interface{}, error) {
	// Extract basic info
	id := fmt.Sprintf("%.0f", node["id"].(float64))
	name := node["name"].(string)
	
	// Get IP addresses
	var ipv4, ipv6 string
	if addrs, ok := node["ipAddresses"].([]interface{}); ok {
		for _, addr := range addrs {
			if addrStr, ok := addr.(string); ok {
				if strings.Contains(addrStr, ":") {
					ipv6 = addrStr
				} else {
					ipv4 = addrStr
				}
			}
		}
	}

	// Get timestamps
	created := "Unknown"
	if createdAt, ok := node["createdAt"].(string); ok {
		created = createdAt
	}

	lastSeen := "Unknown"
	if lastSeenTime, ok := node["lastSeen"].(string); ok {
		lastSeen = lastSeenTime
	}

	// Get key info
	keyExpiry := "Never"
	if expiry, ok := node["expiry"].(string); ok && expiry != "" {
		keyExpiry = expiry
	}

	nodeKey := "Unknown"
	if key, ok := node["nodeKey"].(string); ok {
		nodeKey = key
	}

	// Determine status
	status := "offline"
	if lastSeen != "Unknown" && lastSeen != "" {
		status = "online"
	}

	return map[string]interface{}{
		"id":             id,
		"name":           name,
		"hostname":       name,
		"creator":        "infrastructure",
		"created":        created,
		"lastSeen":       lastSeen,
		"keyExpiry":      keyExpiry,
		"nodeKey":        nodeKey,
		"tailscaleIPv4":  ipv4,
		"tailscaleIPv6":  ipv6,
		"shortDomain":    name,
		"status":         status,
		"managedBy":      "infrastructure",
	}, nil
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
	// First try to get the actual network IP using ip command
	cmd := exec.Command("ip", "route", "get", "1.1.1.1")
	output, err := cmd.Output()
	if err == nil {
		// Parse output like "1.1.1.1 via 192.168.1.1 dev eth0 src 192.168.1.100"
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "src ") {
				parts := strings.Fields(line)
				for i, part := range parts {
					if part == "src" && i+1 < len(parts) {
						ip := parts[i+1]
						// Validate it's a real IP (not 127.x.x.x)
						if !strings.HasPrefix(ip, "127.") && !strings.HasPrefix(ip, "100.64.") {
							return ip
						}
					}
				}
			}
		}
	}

	// Fallback to hostname resolution
	cmd = exec.Command("host", hostname)
	output, err = cmd.Output()
	if err != nil {
		return ""
	}

	// Parse "hostname has address X.X.X.X"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "has address") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				ip := parts[3]
				// Skip loopback and Tailscale IPs
				if !strings.HasPrefix(ip, "127.") && !strings.HasPrefix(ip, "100.64.") {
					return ip
				}
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
