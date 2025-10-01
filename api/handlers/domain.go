package handlers

import (
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
)

type ProvisionDomainRequest struct {
	Domain        string `json:"domain" binding:"required"`
	Realm         string `json:"realm" binding:"required"`
	AdminPassword string `json:"admin_password" binding:"required"`
	DNSBackend    string `json:"dns_backend"`
	DNSForwarder  string `json:"dns_forwarder"`
}

type DomainStatusResponse struct {
	Provisioned bool   `json:"provisioned"`
	Domain      string `json:"domain,omitempty"`
	Realm       string `json:"realm,omitempty"`
	DCReady     bool   `json:"dc_ready"`
	DNSReady    bool   `json:"dns_ready"`
}

// ProvisionDomain provisions a new Samba AD DC domain
func ProvisionDomain(c *gin.Context) {
	var req ProvisionDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Set defaults
	if req.DNSBackend == "" {
		req.DNSBackend = "SAMBA_INTERNAL"
	}

	// Build samba-tool command
	args := []string{
		"domain", "provision",
		"--realm=" + req.Realm,
		"--domain=" + req.Domain,
		"--adminpass=" + req.AdminPassword,
		"--server-role=dc",
		"--dns-backend=" + req.DNSBackend,
		"--use-rfc2307",
	}

	if req.DNSForwarder != "" {
		args = append(args, "--option=dns forwarder = "+req.DNSForwarder)
	}

	// Execute samba-tool domain provision
	cmd := exec.Command("samba-tool", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Domain provisioning failed",
			"details": string(output),
		})
		return
	}

	// Start Samba service
	startCmd := exec.Command("systemctl", "enable", "--now", "samba-ad-dc")
	if err := startCmd.Run(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to start Samba AD DC service",
			"details": string(output),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Domain provisioned successfully",
		"domain":  req.Domain,
		"realm":   req.Realm,
		"output":  string(output),
	})
}

// DomainStatus returns the current status of the domain controller
func DomainStatus(c *gin.Context) {
	// Dev mode: Return dummy status
	if os.Getenv("ENV") == "development" {
		response := DomainStatusResponse{
			Provisioned: true,
			Domain:      "EXAMPLE",
			Realm:       "example.local",
			DCReady:     true,
			DNSReady:    true,
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// Check if domain is provisioned
	cmd := exec.Command("samba-tool", "domain", "info", "127.0.0.1")
	_, err := cmd.CombinedOutput()

	provisioned := err == nil

	// Check if DC is running
	dcCmd := exec.Command("systemctl", "is-active", "samba-ad-dc")
	dcReady := dcCmd.Run() == nil

	response := DomainStatusResponse{
		Provisioned: provisioned,
		DCReady:     dcReady,
		DNSReady:    dcReady, // DNS is internal to Samba
	}

	if provisioned {
		// TODO: Parse domain info from output
		response.Domain = "UNKNOWN"
		response.Realm = "UNKNOWN"
	}

	c.JSON(http.StatusOK, response)
}

// ConfigureDomain updates domain configuration
func ConfigureDomain(c *gin.Context) {
	// TODO: Implement domain configuration updates
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented yet",
	})
}
