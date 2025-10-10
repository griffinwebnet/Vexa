package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/griffinwebnet/vexa/api/utils"
)

type PasswordPolicy struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSymbols   bool `json:"require_symbols"`
}

type DomainPolicySettings struct {
	PasswordComplexity     PasswordPolicy `json:"password_complexity"`
	PasswordExpirationDays int            `json:"password_expiration_days"`
	PasswordHistoryCount   int            `json:"password_history_count"`
	MinPasswordAge         int            `json:"min_password_age"`
	LockoutThreshold       int            `json:"lockout_threshold"`
	LockoutDuration        int            `json:"lockout_duration"`
}

// parsePasswordSettings parses the output from samba-tool domain passwordsettings show
func parsePasswordSettings(output string) (DomainPolicySettings, error) {
	policies := DomainPolicySettings{}
	
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
				complexityOn := matches[1] == "on"
				policies.PasswordComplexity.RequireUppercase = complexityOn
				policies.PasswordComplexity.RequireLowercase = complexityOn
				policies.PasswordComplexity.RequireNumbers = complexityOn
				policies.PasswordComplexity.RequireSymbols = complexityOn
			}
		}
		
		// Parse minimum password length
		if strings.Contains(line, "Minimum password length") {
			re := regexp.MustCompile(`Minimum password length:\s*(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.Atoi(matches[1]); err == nil {
					policies.PasswordComplexity.MinLength = val
				}
			}
		}
		
		// Parse maximum password age
		if strings.Contains(line, "Maximum password age") {
			re := regexp.MustCompile(`Maximum password age:\s*(\d+)`)
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
		
		// Parse minimum password age
		if strings.Contains(line, "Minimum password age") {
			re := regexp.MustCompile(`Minimum password age:\s*(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.Atoi(matches[1]); err == nil {
					policies.MinPasswordAge = val
				}
			}
		}
		
		// Parse lockout threshold
		if strings.Contains(line, "Lockout threshold") {
			re := regexp.MustCompile(`Lockout threshold:\s*(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.Atoi(matches[1]); err == nil {
					policies.LockoutThreshold = val
				}
			}
		}
		
		// Parse lockout duration
		if strings.Contains(line, "Lockout duration") {
			re := regexp.MustCompile(`Lockout duration:\s*(\d+)`)
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
		utils.Info("Failed to parse password settings, using defaults: %v", parseErr)
		// Fallback to defaults if parsing fails
		policies = DomainPolicySettings{
			PasswordComplexity: PasswordPolicy{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireNumbers:   true,
				RequireSymbols:   true,
			},
			PasswordExpirationDays: 365,
			PasswordHistoryCount:   3,
			MinPasswordAge:         1,
			LockoutThreshold:       5,
			LockoutDuration:        30,
		}
	}

	c.JSON(http.StatusOK, policies)
}

// UpdateDomainPolicies updates domain password and security policies
func UpdateDomainPolicies(c *gin.Context) {
	var req DomainPolicySettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Set password complexity
	complexityValue := "off"
	if req.PasswordComplexity.RequireUppercase &&
		req.PasswordComplexity.RequireLowercase &&
		req.PasswordComplexity.RequireNumbers &&
		req.PasswordComplexity.RequireSymbols {
		complexityValue = "on"
	}

	// Apply settings using samba-tool
	commands := [][]string{
		{"samba-tool", "domain", "passwordsettings", "set", "--complexity=" + complexityValue},
		{"samba-tool", "domain", "passwordsettings", "set", "--min-pwd-length=" + strconv.Itoa(req.PasswordComplexity.MinLength)},
		{"samba-tool", "domain", "passwordsettings", "set", "--max-pwd-age=" + strconv.Itoa(req.PasswordExpirationDays)},
		{"samba-tool", "domain", "passwordsettings", "set", "--history-length=" + strconv.Itoa(req.PasswordHistoryCount)},
		{"samba-tool", "domain", "passwordsettings", "set", "--min-pwd-age=" + strconv.Itoa(req.MinPasswordAge)},
	}

	var failedCommands []string
	for _, cmdArgs := range commands {
		utils.Info("Executing: %s %v", cmdArgs[0], cmdArgs[1:])
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
			utils.Info("Successfully executed: %s %v", cmdArgs[0], cmdArgs[1:])
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

	c.JSON(http.StatusOK, gin.H{
		"message": "Password policies updated successfully",
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
