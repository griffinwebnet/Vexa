package handlers

import (
	"crypto/rand"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/griffinwebnet/vexa/api/utils"
)

// Word lists for password generation
var adjectives = []string{
	"Summer", "Winter", "Autumn", "Spring", "Crystal", "Golden", "Silver",
	"Mighty", "Swift", "Brave", "Noble", "Wise", "Azure", "Crimson",
	"Emerald", "Violet", "Amber", "Sapphire", "Ruby", "Diamond",
}

var nouns = []string{
	"Lilypad", "Mountain", "River", "Ocean", "Forest", "Meadow", "Valley",
	"Phoenix", "Dragon", "Eagle", "Tiger", "Falcon", "Panther", "Wolf",
	"Thunder", "Lightning", "Storm", "Breeze", "Sunrise", "Sunset",
}

var symbols = []string{"!", "@", "#", "$", "%", "&", "*"}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

// ChangePassword changes a user's password
func ChangePassword(c *gin.Context) {
	username := c.GetString("username") // From JWT auth middleware

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Verify current password
	cmd, cmdErr := utils.SafeCommand("smbclient", "-L", "localhost", "-U", username+"%"+req.CurrentPassword)
	if cmdErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Command sanitization failed",
		})
		return
	}
	if err := cmd.Run(); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Current password is incorrect",
		})
		return
	}

	// Change password
	cmd, cmdErr = utils.SafeCommand("samba-tool", "user", "setpassword", username, "--newpassword="+req.NewPassword)
	if cmdErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Command sanitization failed",
		})
		return
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to change password",
			"details": string(output),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// UpdateProfile updates a user's profile information
func UpdateProfile(c *gin.Context) {
	username := c.GetString("username") // From JWT auth middleware

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Update display name
	if req.FullName != "" {
		cmd, cmdErr := utils.SafeCommand("samba-tool", "user", "edit", username, "--editor=/bin/echo", "--full-name="+req.FullName)
		if cmdErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Command sanitization failed",
			})
			return
		}
		if err := cmd.Run(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update display name",
			})
			return
		}
	}

	// Update email
	if req.Email != "" {
		cmd, cmdErr := utils.SafeCommand("samba-tool", "user", "edit", username, "--editor=/bin/echo", "--mail-address="+req.Email)
		if cmdErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Command sanitization failed",
			})
			return
		}
		if err := cmd.Run(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update email",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
	})
}

// ResetUserPassword generates a new random password and resets the user's password
func ResetUserPassword(c *gin.Context) {
	username := c.Param("id")

	// Generate random password
	password := generatePassword()

	// Reset password using samba-tool
	cmd, cmdErr := utils.SafeCommand("samba-tool", "user", "setpassword", username, "--newpassword="+password)
	if cmdErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Command sanitization failed",
		})
		return
	}
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to reset password",
			"details": string(output),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Password reset successfully",
		"username": username,
		"password": password,
	})
}

// DisableUser disables a user account
func DisableUser(c *gin.Context) {
	username := c.Param("id")

	cmd, cmdErr := utils.SafeCommand("samba-tool", "user", "disable", username)
	if cmdErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Command sanitization failed",
		})
		return
	}
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to disable user",
			"details": string(output),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "User disabled successfully",
		"username": username,
	})
}

// EnableUser enables a user account
func EnableUser(c *gin.Context) {
	username := c.Param("id")

	cmd, cmdErr := utils.SafeCommand("samba-tool", "user", "enable", username)
	if cmdErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Command sanitization failed",
		})
		return
	}
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to enable user",
			"details": string(output),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "User enabled successfully",
		"username": username,
	})
}

// generatePassword generates a secure random password
func generatePassword() string {
	// Get random adjective
	adjIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(adjectives))))
	adj := adjectives[adjIndex.Int64()]

	// Get random noun
	nounIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(nouns))))
	noun := nouns[nounIndex.Int64()]

	// Get random symbol
	symIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(symbols))))
	sym := symbols[symIndex.Int64()]

	// Get random number between 100-999
	num, _ := rand.Int(rand.Reader, big.NewInt(900))
	num = num.Add(num, big.NewInt(100))

	return adj + noun + sym + num.String()
}
