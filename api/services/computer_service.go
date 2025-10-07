package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/griffinwebnet/vexa/api/config"
	sambaExec "github.com/griffinwebnet/vexa/api/exec"
	"github.com/griffinwebnet/vexa/api/models"
	"github.com/griffinwebnet/vexa/api/utils"
)

// ComputerService handles computer-related business logic
type ComputerService struct {
	sambaTool *sambaExec.SambaTool
	config    *config.Config
}

// NewComputerService creates a new ComputerService instance
func NewComputerService() *ComputerService {
	return &ComputerService{
		sambaTool: sambaExec.NewSambaTool(),
		config:    config.LoadConfig(),
	}
}

// ListComputers returns all computers/devices in the domain with connection status
func (s *ComputerService) ListComputers() ([]models.Computer, error) {

	// Get domain computers from Samba
	output, err := s.sambaTool.Run("computer", "list")
	if err != nil {
		return nil, fmt.Errorf("failed to list computers: %s", output)
	}

	computerNames := strings.Split(strings.TrimSpace(output), "\n")
	domainComputers := make(map[string]bool) // Track which computers are domain-joined

	// Build list of domain computers
	for _, name := range computerNames {
		if name != "" {
			cleanName := strings.TrimSuffix(name, "$")
			domainComputers[cleanName] = true
		}
	}

	// Get Tailscale nodes if Headscale is enabled
	var tailscaleNodes map[string]map[string]interface{}
	if s.isHeadscaleEnabled() {
		tailscaleNodes = s.getTailscaleNodes()
	}

	computers := make([]models.Computer, 0)

	// Process domain computers
	for _, name := range computerNames {
		if name == "" {
			continue
		}

		cleanName := strings.TrimSuffix(name, "$")
		computer := models.Computer{
			Name:           cleanName,
			DNSName:        name,
			Online:         false,
			ConnectionType: "offline",
		}

		// Check local connectivity
		if s.pingComputer(cleanName) {
			computer.Online = true
			computer.ConnectionType = "local"
			if ip := s.getComputerIP(cleanName); ip != "" {
				computer.IPAddress = ip
			}
		}

		// Check Tailscale connectivity
		if tailscaleNodes != nil {
			// Try exact match first
			if node, exists := tailscaleNodes[cleanName]; exists {
				// Found in Tailscale - get overlay IP and URL
				if ip := s.extractIPFromNode(node); ip != "" {
					computer.OverlayIP = ip
					computer.Online = true
					computer.ConnectionType = "overlay"

					// Generate overlay URL
					domainService := NewDomainService()
					domainStatus, _ := domainService.GetDomainStatus()
					if domainStatus != nil && domainStatus.Domain != "" && domainStatus.Domain != "PROVISIONED" {
						computer.OverlayURL = fmt.Sprintf("%s.%s.mesh", cleanName, domainStatus.Domain)
					}
				}
			} else {
				// Try to find a Tailscale node that might be the same machine
				// Check if this is the domain controller and look for server names
				dcHostname := s.getDomainControllerHostname()
				if cleanName == strings.TrimSuffix(dcHostname, "$") {
					// This is the DC, look for server names from configuration
					serverNames := s.config.GetAllServerNames()
					for _, serverName := range serverNames {
						if node, exists := tailscaleNodes[serverName]; exists {
							if ip := s.extractIPFromNode(node); ip != "" {
								computer.OverlayIP = ip
								computer.Online = true
								computer.ConnectionType = "overlay"

								// Generate overlay URL
								domainService := NewDomainService()
								domainStatus, _ := domainService.GetDomainStatus()
								if domainStatus != nil && domainStatus.Domain != "" && domainStatus.Domain != "PROVISIONED" {
									computer.OverlayURL = fmt.Sprintf("%s.%s.mesh", cleanName, domainStatus.Domain)
								}
								break
							}
						}
					}
				}
			}
		}

		computers = append(computers, computer)
	}

	// Add Tailscale-only nodes (not domain-joined) and the server itself
	if tailscaleNodes != nil {
		dcHostname := s.getDomainControllerHostname()
		dcCleanName := strings.TrimSuffix(dcHostname, "$")

		for nodeName, node := range tailscaleNodes {
			// Skip if already in domain computers list
			if domainComputers[nodeName] {
				continue
			}

			// Check if this is the server itself using dynamic configuration
			isServer := s.config.IsServerName(nodeName) || nodeName == dcCleanName

			// This is either a Tailscale-only node (like a NAS, etc.) or the server itself
			computer := models.Computer{
				Name:           nodeName,
				DNSName:        nodeName,
				Online:         true,
				ConnectionType: "overlay",
			}

			if ip := s.extractIPFromNode(node); ip != "" {
				computer.OverlayIP = ip
			}

			// Generate overlay URL
			domainService := NewDomainService()
			domainStatus, _ := domainService.GetDomainStatus()
			if domainStatus != nil && domainStatus.Domain != "" && domainStatus.Domain != "PROVISIONED" {
				computer.OverlayURL = fmt.Sprintf("%s.%s.mesh", nodeName, domainStatus.Domain)
			}

			// Add special indicator for the server
			if isServer {
				computer.Name = nodeName + " (Server)"
			}

			computers = append(computers, computer)
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
	fmt.Printf("DEBUG: Looking for machine: %s\n", machineId)

	// Query Headscale for machine details
	cmd, cmdErr := utils.SafeCommand("headscale", "nodes", "list", "--output", "json")
	if cmdErr != nil {
		return nil, fmt.Errorf("command sanitization failed: %v", cmdErr)
	}
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to query headscale: %v", err)
	}

	// Parse JSON to find the specific machine
	var nodes []map[string]interface{}
	if err := json.Unmarshal(output, &nodes); err != nil {
		return nil, fmt.Errorf("failed to parse headscale output: %v", err)
	}

	fmt.Printf("DEBUG: Found %d nodes in Headscale\n", len(nodes))

	// Find the machine by ID or name
	for i, node := range nodes {
		nodeName := "unknown"
		nodeID := "unknown"

		if name, ok := node["name"].(string); ok {
			nodeName = name
		}
		if id, ok := node["id"].(float64); ok {
			nodeID = fmt.Sprintf("%.0f", id)
		}

		fmt.Printf("DEBUG: Node %d: ID=%s, Name=%s\n", i, nodeID, nodeName)

		// Check by ID (convert float64 to string)
		if nodeID, ok := node["id"].(float64); ok {
			if fmt.Sprintf("%.0f", nodeID) == machineId {
				fmt.Printf("DEBUG: Found machine by ID match\n")
				return s.formatMachineDetails(node)
			}
		}

		// Check by name (exact match)
		if nodeName, ok := node["name"].(string); ok && nodeName == machineId {
			fmt.Printf("DEBUG: Found machine by exact name match\n")
			return s.formatMachineDetails(node)
		}

		// Check by name with $ suffix (for domain computers)
		if nodeName, ok := node["name"].(string); ok && nodeName == machineId+"$" {
			fmt.Printf("DEBUG: Found machine by name with $ suffix match\n")
			return s.formatMachineDetails(node)
		}

		// Check by name without $ suffix
		if nodeName, ok := node["name"].(string); ok && strings.TrimSuffix(nodeName, "$") == machineId {
			fmt.Printf("DEBUG: Found machine by name without $ suffix match\n")
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
	if addrs, ok := node["ip_addresses"].([]interface{}); ok {
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
		"id":            id,
		"name":          name,
		"hostname":      name,
		"creator":       "infrastructure",
		"created":       created,
		"lastSeen":      lastSeen,
		"keyExpiry":     keyExpiry,
		"nodeKey":       nodeKey,
		"tailscaleIPv4": ipv4,
		"tailscaleIPv6": ipv6,
		"shortDomain":   name,
		"status":        status,
		"managedBy":     "infrastructure",
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
	cmd, err := utils.SafeCommand("systemctl", "is-active", "headscale")
	if err != nil {
		return false
	}
	return cmd.Run() == nil
}

func (s *ComputerService) pingComputer(hostname string) bool {
	cmd, err := utils.SafeCommand("ping", "-c", "1", "-W", "1", hostname)
	if err != nil {
		return false
	}
	return cmd.Run() == nil
}

func (s *ComputerService) getComputerIP(hostname string) string {
	// First try to get the actual network IP using ip command
	cmd, cmdErr := utils.SafeCommand("ip", "route", "get", "1.1.1.1")
	if cmdErr != nil {
		return ""
	}
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
	cmd, cmdErr = utils.SafeCommand("host", hostname)
	if cmdErr != nil {
		return ""
	}
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
	cmd, cmdErr := utils.SafeCommand("headscale", "nodes", "list", "--output", "json")
	if cmdErr != nil {
		return ""
	}
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

				// Try to get IP from ip_addresses array (Headscale uses underscore)
				if ipAddresses, ok := node["ip_addresses"].([]interface{}); ok {
					for _, addr := range ipAddresses {
						if addrStr, ok := addr.(string); ok {
							// Return first IPv4 address (not IPv6)
							if !strings.Contains(addrStr, ":") {
								return addrStr
							}
						}
					}
				}

				// Fallback to direct IP fields
				if ipv4, ok := node["ipv4"].(string); ok && ipv4 != "" {
					return ipv4
				}
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
	cmd, cmdErr := utils.SafeCommand("hostname")
	if cmdErr != nil {
		return ""
	}
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// getTailscaleNodes returns a map of node names to their data from Headscale
func (s *ComputerService) getTailscaleNodes() map[string]map[string]interface{} {
	cmd, cmdErr := utils.SafeCommand("headscale", "nodes", "list", "--output", "json")
	if cmdErr != nil {
		return nil
	}
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var nodes []map[string]interface{}
	if err := json.Unmarshal(output, &nodes); err != nil {
		return nil
	}

	nodeMap := make(map[string]map[string]interface{})
	for _, node := range nodes {
		if nodeName, ok := node["name"].(string); ok {
			cleanName := strings.TrimSuffix(nodeName, "$")
			nodeMap[cleanName] = node
		}
	}
	return nodeMap
}

// extractIPFromNode extracts the first IPv4 address from a Headscale node
func (s *ComputerService) extractIPFromNode(node map[string]interface{}) string {
	if ipAddresses, ok := node["ip_addresses"].([]interface{}); ok {
		for _, addr := range ipAddresses {
			if addrStr, ok := addr.(string); ok {
				// Return first IPv4 address (not IPv6)
				if !strings.Contains(addrStr, ":") {
					return addrStr
				}
			}
		}
	}
	return ""
}
