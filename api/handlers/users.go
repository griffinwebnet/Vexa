package handlers

import (
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

type User struct {
	Username    string   `json:"username"`
	FullName    string   `json:"full_name"`
	Email       string   `json:"email"`
	Enabled     bool     `json:"enabled"`
	Groups      []string `json:"groups"`
	Description string   `json:"description"`
}

type CreateUserRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Description string `json:"description"`
	Group       string `json:"group"`
	OUPath      string `json:"ou_path"`
}

// ListUsers returns all users in the domain
func ListUsers(c *gin.Context) {
	// Dev mode: Return dummy data
	if os.Getenv("ENV") == "development" {
		users := []User{
			{Username: "jsmith", FullName: "John Smith", Email: "jsmith@example.com", Enabled: true},
			{Username: "mjohnson", FullName: "Mary Johnson", Email: "mjohnson@example.com", Enabled: true},
			{Username: "bwilliams", FullName: "Bob Williams", Email: "bwilliams@example.com", Enabled: true},
			{Username: "administrator", FullName: "Administrator", Email: "admin@example.com", Enabled: true},
		}
		c.JSON(http.StatusOK, gin.H{
			"users": users,
			"count": len(users),
		})
		return
	}

	cmd := exec.Command("samba-tool", "user", "list")
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list users",
			"details": string(output),
		})
		return
	}

	// Parse user list
	usernames := strings.Split(strings.TrimSpace(string(output)), "\n")
	users := make([]User, 0, len(usernames))

	for _, username := range usernames {
		if username != "" {
			users = append(users, User{
				Username: username,
				Enabled:  true, // TODO: Get actual status
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

// CreateUser creates a new user in the domain
func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Build samba-tool command
	args := []string{
		"user", "create",
		req.Username,
		req.Password,
	}

	if req.FullName != "" {
		args = append(args, "--given-name="+req.FullName)
	}
	if req.Email != "" {
		args = append(args, "--mail-address="+req.Email)
	}
	if req.OUPath != "" {
		args = append(args, "--userou="+req.OUPath)
	}

	cmd := exec.Command("samba-tool", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create user",
			"details": string(output),
		})
		return
	}

	// Add user to group if specified
	if req.Group != "" && req.Group != "Domain Users" {
		groupCmd := exec.Command("samba-tool", "group", "addmembers", req.Group, req.Username)
		if err := groupCmd.Run(); err != nil {
			// User created but group add failed - log but don't fail
			c.JSON(http.StatusCreated, gin.H{
				"message":  "User created but failed to add to group",
				"username": req.Username,
				"warning":  "Group membership not set",
			})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "User created successfully",
		"username": req.Username,
	})
}

// GetUser returns details for a specific user
func GetUser(c *gin.Context) {
	username := c.Param("id")

	cmd := exec.Command("samba-tool", "user", "show", username)
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// TODO: Parse user details from output
	c.JSON(http.StatusOK, gin.H{
		"username": username,
		"details":  string(output),
	})
}

// UpdateUser updates an existing user
func UpdateUser(c *gin.Context) {
	username := c.Param("id")

	// TODO: Implement user update logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"message":  "User update not implemented",
		"username": username,
	})
}

// DeleteUser removes a user from the domain
func DeleteUser(c *gin.Context) {
	username := c.Param("id")

	cmd := exec.Command("samba-tool", "user", "delete", username)
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete user",
			"details": string(output),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "User deleted successfully",
		"username": username,
	})
}
