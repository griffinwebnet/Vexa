package handlers

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
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

// ResetUserPassword generates a new random password and resets the user's password
func ResetUserPassword(c *gin.Context) {
	username := c.Param("id")

	// Generate random password
	password := generatePassword()

	// Reset password using samba-tool
	cmd := exec.Command("samba-tool", "user", "setpassword", username, "--newpassword="+password)
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

	cmd := exec.Command("samba-tool", "user", "disable", username)
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

	cmd := exec.Command("samba-tool", "user", "enable", username)
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

// generatePassword creates a random password like "SummerLilypad&216"
func generatePassword() string {
	// Random adjective
	adjIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(adjectives))))
	adj := adjectives[adjIdx.Int64()]

	// Random noun
	nounIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(nouns))))
	noun := nouns[nounIdx.Int64()]

	// Random symbol
	symIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(symbols))))
	symbol := symbols[symIdx.Int64()]

	// Random 3-digit number
	num, _ := rand.Int(rand.Reader, big.NewInt(900))
	number := num.Int64() + 100 // Ensures 3 digits (100-999)

	// Combine: Adjective + Noun + Symbol + Number
	password := strings.Join([]string{adj, noun, symbol, string(rune(number/100 + '0')), string(rune((number/10)%10 + '0')), string(rune(number%10 + '0'))}, "")

	return password
}
