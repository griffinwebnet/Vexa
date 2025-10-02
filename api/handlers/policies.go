package handlers

import (
	"net/http"
	"os/exec"
	"strconv"

	"github.com/gin-gonic/gin"
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

// GetDomainPolicies returns current domain password and security policies
func GetDomainPolicies(c *gin.Context) {
	// Get current password settings using samba-tool
	cmd := exec.Command("samba-tool", "domain", "passwordsettings", "show")
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve password settings",
			"details": string(output),
		})
		return
	}

	// TODO: Parse output and return structured data
	// For now, return defaults
	policies := DomainPolicySettings{
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

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		if err := cmd.Run(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to apply password settings",
				"command": cmdArgs,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password policies updated successfully",
	})
}

// GetOUList returns list of organizational units
func GetOUList(c *gin.Context) {
	cmd := exec.Command("samba-tool", "ou", "list")
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list OUs",
			"details": string(output),
		})
		return
	}

	// TODO: Parse OU list and build hierarchy
	c.JSON(http.StatusOK, gin.H{
		"ous": string(output),
	})
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

	cmd := exec.Command("samba-tool", "ou", "create", ouPath)
	if req.Description != "" {
		cmd.Args = append(cmd.Args, "--description="+req.Description)
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

	cmd := exec.Command("samba-tool", "ou", "delete", ouPath)
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
