package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/griffinwebnet/vexa/api/services"
)

// DeploymentHandler handles computer deployment operations
type DeploymentHandler struct {
	headscaleService *services.HeadscaleService
}

// NewDeploymentHandler creates a new DeploymentHandler instance
func NewDeploymentHandler() *DeploymentHandler {
	return &DeploymentHandler{
		headscaleService: services.NewHeadscaleService(),
	}
}

// GetDeploymentScripts returns available deployment options
func (h *DeploymentHandler) GetDeploymentScripts(c *gin.Context) {
	headscaleEnabled := h.headscaleService.IsEnabled()

	scripts := []gin.H{
		{
			"id":          "tailscale-domain",
			"name":        "Domain Join with Tailscale",
			"description": "Download Tailscale, join domain, and connect to Tailnet",
			"icon":        "🔗",
			"enabled":     headscaleEnabled,
			"requirements": []string{
				"Administrator privileges",
				"Network access to domain controller",
				"Headscale server configured",
			},
		},
		{
			"id":          "domain-only",
			"name":        "Domain Join Only",
			"description": "Join computer to domain without Tailscale",
			"icon":        "🏢",
			"enabled":     true, // Always available
			"requirements": []string{
				"Administrator privileges",
				"Network access to domain controller",
			},
		},
		{
			"id":          "tailnet-add",
			"name":        "Add to Tailnet",
			"description": "Add existing domain-joined computer to Tailnet",
			"icon":        "🌐",
			"enabled":     headscaleEnabled,
			"requirements": []string{
				"Administrator privileges",
				"Computer already domain-joined",
				"Headscale server configured",
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"scripts":           scripts,
		"headscale_enabled": headscaleEnabled,
	})
}

// GenerateDeploymentCommand generates a PowerShell command for deployment
func (h *DeploymentHandler) GenerateDeploymentCommand(c *gin.Context) {
	var req struct {
		ScriptType       string `json:"script_type" binding:"required"`
		DomainName       string `json:"domain_name"`
		DomainController string `json:"domain_controller"`
		ComputerName     string `json:"computer_name,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Check if Headscale is enabled for Tailscale options
	headscaleEnabled := h.headscaleService.IsEnabled()
	if (req.ScriptType == "tailscale-domain" || req.ScriptType == "tailnet-add") && !headscaleEnabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Headscale is not enabled. Cannot use Tailscale deployment options.",
		})
		return
	}

	var command string
	var scriptName string
	var authKey string

	// Get existing infrastructure pre-auth key for Tailscale options
	if req.ScriptType == "tailscale-domain" || req.ScriptType == "tailnet-add" {
		// Use the existing infrastructure key instead of creating a new one
		existingKey, err := h.headscaleService.GetInfrastructureKey()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to get infrastructure key: %v", err),
			})
			return
		}
		authKey = existingKey
	}

	switch req.ScriptType {
	case "tailscale-domain":
		scriptName = "domain-join-with-tailscale.ps1"
		command = fmt.Sprintf(`powershell -ExecutionPolicy Bypass -File ".\%s"`, scriptName)

		// Add parameters
		if req.DomainName != "" {
			command += fmt.Sprintf(` -DomainName "%s"`, req.DomainName)
		}
		if req.DomainController != "" {
			command += fmt.Sprintf(` -DomainController "%s"`, req.DomainController)
		}
		if authKey != "" {
			command += fmt.Sprintf(` -TailscaleAuthKey "%s"`, authKey)
		}
		if req.ComputerName != "" {
			command += fmt.Sprintf(` -ComputerName "%s"`, req.ComputerName)
		}

	case "domain-only":
		scriptName = "domain-join-only.ps1"
		command = fmt.Sprintf(`powershell -ExecutionPolicy Bypass -File ".\%s"`, scriptName)

		if req.DomainName != "" {
			command += fmt.Sprintf(` -DomainName "%s"`, req.DomainName)
		}
		if req.ComputerName != "" {
			command += fmt.Sprintf(` -ComputerName "%s"`, req.ComputerName)
		}

	case "tailnet-add":
		scriptName = "tailnet-add.ps1"
		command = fmt.Sprintf(`powershell -ExecutionPolicy Bypass -File ".\%s"`, scriptName)

		if authKey != "" {
			command += fmt.Sprintf(` -TailscaleAuthKey "%s"`, authKey)
		}
		if req.ComputerName != "" {
			command += fmt.Sprintf(` -ComputerName "%s"`, req.ComputerName)
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid script type",
		})
		return
	}

	response := gin.H{
		"command":    command,
		"script_url": fmt.Sprintf("%s/api/deployment/scripts/%s", getBaseURL(c), scriptName),
		"instructions": []string{
			"1. Copy the command above",
			"2. Open PowerShell as Administrator",
			"3. Paste and run the command",
			"4. Follow the on-screen prompts",
		},
	}

	// Include auth key info for Tailscale options (for debugging/admin purposes)
	if authKey != "" {
		response["auth_key_generated"] = true
		response["auth_key_preview"] = authKey[:20] + "..." // Show first 20 chars for verification
	}

	c.JSON(http.StatusOK, response)
}

// ServeDeploymentScript serves the actual PowerShell script files
func (h *DeploymentHandler) ServeDeploymentScript(c *gin.Context) {
	scriptName := c.Param("script")

	// Validate script name for security
	allowedScripts := []string{
		"domain-join-with-tailscale.ps1",
		"domain-join-only.ps1",
		"tailnet-add.ps1",
	}

	isAllowed := false
	for _, allowed := range allowedScripts {
		if scriptName == allowed {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Script not found",
		})
		return
	}

	// Get the script content
	scriptPath := filepath.Join("scripts", "deployment", scriptName)
	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Script file not found",
		})
		return
	}

	// Replace template variables with actual values
	// Resolve login server URL (full) used by tailscale up
	loginServer := h.headscaleService.GetLoginServerFull()
	if loginServer == "" {
		// fall back to API base + /mesh when none configured
		loginServer = getBaseURL(c) + "/mesh"
	}
	processed := strings.ReplaceAll(string(scriptContent), "{{LOGIN_SERVER}}", loginServer)
	processedContent := processed

	// Set appropriate headers
	c.Header("Content-Type", "text/plain")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", scriptName))
	c.String(http.StatusOK, processedContent)
}

// Helper function to get base URL
func getBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	host := c.Request.Host
	if host == "" {
		host = "localhost:8080" // Vexa API port (Headscale is on 8443 and proxied via /mesh/)
	}
	return fmt.Sprintf("%s://%s", scheme, host)
}
