package handlers

import (
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

type Computer struct {
	Name            string `json:"name"`
	DNSName         string `json:"dns_name"`
	OperatingSystem string `json:"operating_system"`
	LastLogon       string `json:"last_logon"`
	ConnectionType  string `json:"connection_type"` // "local", "overlay", "offline"
	Online          bool   `json:"online"`
	IPAddress       string `json:"ip_address,omitempty"`
	OverlayIP       string `json:"overlay_ip,omitempty"`
}

// ListComputers returns all computers/devices in the domain with connection status
func ListComputers(c *gin.Context) {
	// Dev mode: Return dummy data
	if os.Getenv("ENV") == "development" {
		computers := []Computer{
			{
				Name:            "DESKTOP-IT-01",
				DNSName:         "DESKTOP-IT-01.example.local",
				OperatingSystem: "Windows 11 Pro",
				Online:          true,
				ConnectionType:  "local",
				IPAddress:       "192.168.1.105",
			},
			{
				Name:            "LAPTOP-SALES-03",
				DNSName:         "LAPTOP-SALES-03.example.local",
				OperatingSystem: "Windows 10 Pro",
				Online:          true,
				ConnectionType:  "overlay",
				IPAddress:       "192.168.1.147",
				OverlayIP:       "100.64.0.12",
			},
			{
				Name:            "SERVER-FILE-01",
				DNSName:         "SERVER-FILE-01.example.local",
				OperatingSystem: "Windows Server 2022",
				Online:          true,
				ConnectionType:  "local",
				IPAddress:       "192.168.1.50",
			},
			{
				Name:            "LAPTOP-REMOTE-05",
				DNSName:         "LAPTOP-REMOTE-05.example.local",
				OperatingSystem: "Windows 11 Pro",
				Online:          true,
				ConnectionType:  "overlay",
				OverlayIP:       "100.64.0.25",
			},
			{
				Name:            "WORKSTATION-HR-02",
				DNSName:         "WORKSTATION-HR-02.example.local",
				OperatingSystem: "Windows 10 Pro",
				Online:          false,
				ConnectionType:  "offline",
			},
		}
		c.JSON(http.StatusOK, gin.H{
			"computers": computers,
			"count":     len(computers),
		})
		return
	}

	// Get computers from Samba
	cmd := exec.Command("samba-tool", "computer", "list")
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list computers",
			"details": string(output),
		})
		return
	}

	// Parse computer list
	computerNames := strings.Split(strings.TrimSpace(string(output)), "\n")
	computers := make([]Computer, 0)

	// Check if Headscale is enabled
	headscaleEnabled := isHeadscaleEnabled()

	for _, name := range computerNames {
		if name == "" {
			continue
		}

		computer := Computer{
			Name:           strings.TrimSuffix(name, "$"),
			DNSName:        name,
			Online:         false,
			ConnectionType: "offline",
		}

		// Try to ping the computer to check if online
		pingCmd := exec.Command("ping", "-c", "1", "-W", "1", computer.Name)
		if pingCmd.Run() == nil {
			computer.Online = true
			computer.ConnectionType = "local"

			// Try to get IP address
			if ip := getComputerIP(computer.Name); ip != "" {
				computer.IPAddress = ip
			}
		}

		// If Headscale is enabled, check overlay connection
		if headscaleEnabled {
			if overlayIP := getHeadscaleIP(computer.Name); overlayIP != "" {
				computer.OverlayIP = overlayIP
				computer.Online = true
				computer.ConnectionType = "overlay"
			}
		}

		computers = append(computers, computer)
	}

	c.JSON(http.StatusOK, gin.H{
		"computers": computers,
		"count":     len(computers),
	})
}

// GetComputer returns details for a specific computer
func GetComputer(c *gin.Context) {
	computerName := c.Param("id")

	// TODO: Get detailed computer info
	c.JSON(http.StatusOK, gin.H{
		"name":    computerName,
		"details": "Computer details not yet implemented",
	})
}

// DeleteComputer removes a computer from the domain
func DeleteComputer(c *gin.Context) {
	computerName := c.Param("id")

	cmd := exec.Command("samba-tool", "computer", "delete", computerName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete computer",
			"details": string(output),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Computer deleted successfully",
		"name":    computerName,
	})
}

// Helper functions

func isHeadscaleEnabled() bool {
	cmd := exec.Command("systemctl", "is-active", "headscale")
	return cmd.Run() == nil
}

func getComputerIP(hostname string) string {
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

func getHeadscaleIP(hostname string) string {
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
