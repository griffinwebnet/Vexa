package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/griffinwebnet/vexa/api/utils"
)

type DomainPolicySettings struct {
	PasswordComplexityEnabled bool `json:"password_complexity_enabled"` // Simple toggle: true = secure, false = insecure
	PasswordExpirationDays    int  `json:"password_expiration_days"`    // 0 = never expires
	PasswordHistoryCount      int  `json:"password_history_count"`      // Number of previous passwords to remember
	MinPasswordLength         int  `json:"min_password_length"`         // Minimum password length
	LockoutThreshold          int  `json:"lockout_threshold"`           // 0 = disabled
	LockoutDuration           int  `json:"lockout_duration"`            // Minutes
}

// parsePasswordSettings parses the output from samba-tool domain passwordsettings show
func parsePasswordSettings(output string) (DomainPolicySettings, error) {
	policies := DomainPolicySettings{
		PasswordExpirationDays: 0, // Default to never expires
		MinPasswordLength:      7, // Samba default
	}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse password complexity (on/off)
		if strings.Contains(line, "Password complexity") {
			re := regexp.MustCompile(`Password complexity:\s*(on|off)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				policies.PasswordComplexityEnabled = (matches[1] == "on")
			}
		}

		// Parse minimum password length
		if strings.Contains(line, "Minimum password length") {
			re := regexp.MustCompile(`Minimum password length:\s*(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.Atoi(matches[1]); err == nil {
					policies.MinPasswordLength = val
				}
			}
		}

		// Parse maximum password age (days)
		if strings.Contains(line, "Maximum password age") {
			re := regexp.MustCompile(`Maximum password age.*:\s*(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.Atoi(matches[1]); err == nil {
					policies.PasswordExpirationDays = val
				}
			}
		}

		// Parse password history length
		if strings.Contains(line, "Password history length") {
			re := regexp.MustCompile(`Password history length:\s*(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.Atoi(matches[1]); err == nil {
					policies.PasswordHistoryCount = val
				}
			}
		}

		// Parse lockout threshold
		if strings.Contains(line, "Account lockout threshold") {
			re := regexp.MustCompile(`Account lockout threshold.*:\s*(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.Atoi(matches[1]); err == nil {
					policies.LockoutThreshold = val
				}
			}
		}

		// Parse lockout duration (minutes)
		if strings.Contains(line, "Account lockout duration") {
			re := regexp.MustCompile(`Account lockout duration.*:\s*(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.Atoi(matches[1]); err == nil {
					policies.LockoutDuration = val
				}
			}
		}
	}

	return policies, nil
}

// GetDomainPolicies returns current domain password and security policies
func GetDomainPolicies(c *gin.Context) {
	// Get current password settings using samba-tool
	cmd, cmdErr := utils.SafeCommand("samba-tool", "domain", "passwordsettings", "show")
	if cmdErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Command sanitization failed",
			"details": cmdErr.Error(),
		})
		return
	}
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve password settings",
			"details": string(output),
		})
		return
	}

	// Parse the actual samba-tool output
	policies, parseErr := parsePasswordSettings(string(output))
	if parseErr != nil {
		utils.Warn("Failed to parse password settings, using defaults: %v", parseErr)
		// Fallback to defaults if parsing fails
		policies = DomainPolicySettings{
			PasswordComplexityEnabled: false, // Default to insecure (off)
			MinPasswordLength:         7,     // Samba minimum
			PasswordExpirationDays:    0,     // Never expires
			PasswordHistoryCount:      0,     // No history
			LockoutThreshold:          0,     // No lockout
			LockoutDuration:           30,    // 30 minutes if enabled
		}
	}

	utils.Info("Retrieved domain policies: complexity=%v, expiration=%d days, history=%d",
		policies.PasswordComplexityEnabled, policies.PasswordExpirationDays, policies.PasswordHistoryCount)

	c.JSON(http.StatusOK, policies)
}

// UpdateDomainPolicies updates domain password and security policies and applies them to ALL users
func UpdateDomainPolicies(c *gin.Context) {
	var req DomainPolicySettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	utils.Info("Updating domain policies: complexity=%v, expiration=%d days, history=%d",
		req.PasswordComplexityEnabled, req.PasswordExpirationDays, req.PasswordHistoryCount)

	// Set password complexity (simple on/off toggle)
	complexityValue := "off"
	if req.PasswordComplexityEnabled {
		complexityValue = "on"
	}

	// Apply domain-wide settings using samba-tool
	commands := [][]string{
		{"samba-tool", "domain", "passwordsettings", "set", "--complexity=" + complexityValue},
		{"samba-tool", "domain", "passwordsettings", "set", "--min-pwd-length=" + strconv.Itoa(req.MinPasswordLength)},
		{"samba-tool", "domain", "passwordsettings", "set", "--max-pwd-age=" + strconv.Itoa(req.PasswordExpirationDays)},
		{"samba-tool", "domain", "passwordsettings", "set", "--history-length=" + strconv.Itoa(req.PasswordHistoryCount)},
	}

	var failedCommands []string
	for _, cmdArgs := range commands {
		utils.Info("Executing: %s", strings.Join(cmdArgs, " "))
		cmd, cmdErr := utils.SafeCommand(cmdArgs[0], cmdArgs[1:]...)
		if cmdErr != nil {
			utils.Error("Command sanitization failed for %v: %v", cmdArgs, cmdErr)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Command sanitization failed",
				"command": cmdArgs,
			})
			return
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			utils.Error("Failed to execute %v: %v, output: %s", cmdArgs, err, string(output))
			failedCommands = append(failedCommands, strings.Join(cmdArgs, " ")+": "+string(output))
		} else {
			utils.Info("Successfully executed: %s", strings.Join(cmdArgs, " "))
		}
	}

	if len(failedCommands) > 0 {
		utils.Error("Some password policy commands failed: %v", failedCommands)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Some password settings failed to apply",
			"details": failedCommands,
		})
		return
	}

	// Apply password expiration to ALL existing users
	if req.PasswordExpirationDays == 0 {
		utils.Info("Setting all user passwords to never expire")
		// Get list of all users
		userListCmd, cmdErr := utils.SafeCommand("samba-tool", "user", "list")
		if cmdErr == nil {
			output, err := userListCmd.CombinedOutput()
			if err == nil {
				users := strings.Split(strings.TrimSpace(string(output)), "\n")
				successCount := 0
				failCount := 0

				for _, username := range users {
					username = strings.TrimSpace(username)
					if username == "" || username == "krbtgt" || username == "Guest" {
						continue // Skip system accounts
					}

					// Set password to never expire for this user
					expiryCmd, expiryErr := utils.SafeCommand("samba-tool", "user", "setexpiry", username, "--noexpiry")
					if expiryErr == nil {
						_, err := expiryCmd.CombinedOutput()
						if err == nil {
							successCount++
						} else {
							failCount++
						}
					}
				}
				utils.Info("Set password never expires for %d users (%d failed)", successCount, failCount)
			}
		}
	}

	utils.Info("Domain password policies updated successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Password policies updated successfully and applied to all users",
	})
}

// GetOUList returns list of organizational units
func GetOUList(c *gin.Context) {
	utils.Info("Fetching organizational units list")

	cmd, err := utils.SafeCommand("samba-tool", "ou", "list")
	if err != nil {
		utils.Error("Command sanitization failed for samba-tool ou list: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Command sanitization failed",
		})
		return
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.Error("Failed to list OUs: %v, output: %s", err, string(output))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list OUs",
			"details": string(output),
		})
		return
	}

	// Parse OU list and build hierarchy
	ouStructure := parseOUList(string(output))
	utils.Info("Successfully fetched %d OUs", len(ouStructure.Children))

	c.JSON(http.StatusOK, ouStructure)
}

// CreateOU creates a new organizational unit
func CreateOU(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		ParentPath  string `json:"parent_path"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Build full OU path
	ouPath := "OU=" + req.Name
	if req.ParentPath != "" {
		ouPath += "," + req.ParentPath
	}

	args := []string{"ou", "create", ouPath}
	if req.Description != "" {
		args = append(args, "--description="+req.Description)
	}

	cmd, cmdErr := utils.SafeCommand("samba-tool", args...)
	if cmdErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Command sanitization failed",
			"details": cmdErr.Error(),
		})
		return
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create OU",
			"details": string(output),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "OU created successfully",
		"path":    ouPath,
	})
}

// DeleteOU removes an organizational unit
func DeleteOU(c *gin.Context) {
	ouPath := c.Param("path")

	cmd, cmdErr := utils.SafeCommand("samba-tool", "ou", "delete", ouPath)
	if cmdErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Command sanitization failed",
			"details": cmdErr.Error(),
		})
		return
	}
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete OU",
			"details": string(output),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OU deleted successfully",
	})
}

// OUStructure represents the hierarchical structure of organizational units
type OUStructure struct {
	Name        string        `json:"name"`
	Path        string        `json:"path"`
	Description string        `json:"description,omitempty"`
	Children    []OUStructure `json:"children"`
}

// parseOUList parses the samba-tool ou list output and builds a hierarchical structure
func parseOUList(output string) OUStructure {
	// Default root structure
	root := OUStructure{
		Name:     "Domain",
		Path:     "root",
		Children: []OUStructure{},
	}

	// Parse the output line by line
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse OU line format: "OU=Name,DC=domain,DC=com"
		if strings.HasPrefix(line, "OU=") {
			ou := parseOULine(line)
			if ou.Name != "" {
				root.Children = append(root.Children, ou)
			}
		}
	}

	// If no OUs found, add default Domain Controllers OU
	if len(root.Children) == 0 {
		root.Children = append(root.Children, OUStructure{
			Name:        "Domain Controllers",
			Path:        "OU=Domain Controllers",
			Description: "Default controllers container",
			Children:    []OUStructure{},
		})
	}

	return root
}

// parseOULine parses a single OU line and extracts the name and path
func parseOULine(line string) OUStructure {
	// Extract OU name from "OU=Name,DC=domain,DC=com"
	parts := strings.Split(line, ",")
	if len(parts) == 0 {
		return OUStructure{}
	}

	ouPart := parts[0] // "OU=Name"
	if !strings.HasPrefix(ouPart, "OU=") {
		return OUStructure{}
	}

	name := strings.TrimPrefix(ouPart, "OU=")
	if name == "" {
		return OUStructure{}
	}

	return OUStructure{
		Name:        name,
		Path:        line, // Full DN path
		Description: "Organizational Unit",
		Children:    []OUStructure{},
	}
}
