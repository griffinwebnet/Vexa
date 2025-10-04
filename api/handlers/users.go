package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vexa/api/models"
	"github.com/vexa/api/services"
	"github.com/vexa/api/utils"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: services.NewUserService(),
	}
}

// ListUsers returns all users in the domain
func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.userService.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

// CreateUser creates a new user in the domain
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	err := h.userService.CreateUser(req)
	if err != nil {
		// Check if it's a warning (user created but group add failed)
		if strings.Contains(err.Error(), "user created but failed to add to group") {
			c.JSON(http.StatusCreated, gin.H{
				"message":  "User created but failed to add to group",
				"username": req.Username,
				"warning":  "Group membership not set",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "User created successfully",
		"username": req.Username,
	})
}

// GetUser returns details for a specific user
func (h *UserHandler) GetUser(c *gin.Context) {
	username := c.Param("id")

	user, err := h.userService.GetUser(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser updates an existing user
func (h *UserHandler) UpdateUser(c *gin.Context) {
	username := c.Param("id")

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	err := h.userService.UpdateUser(username, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "User updated successfully",
		"username": username,
	})
}

// DeleteUser removes a user from the domain
func (h *UserHandler) DeleteUser(c *gin.Context) {
	username := c.Param("id")

	err := h.userService.DeleteUser(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "User deleted successfully",
		"username": username,
	})
}

// DisableUser disables a user account
func (h *UserHandler) DisableUser(c *gin.Context) {
	username := c.Param("id")

	err := h.userService.DisableUser(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "User disabled successfully",
		"username": username,
	})
}

// EnableUser enables a user account
func (h *UserHandler) EnableUser(c *gin.Context) {
	username := c.Param("id")

	err := h.userService.EnableUser(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "User enabled successfully",
		"username": username,
	})
}

// ChangePassword allows users to change their own password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	fmt.Printf("DEBUG: ChangePassword endpoint called\n")

	// Get username from JWT token
	claims, exists := c.Get("claims")
	if !exists {
		fmt.Printf("DEBUG: No authentication claims found\n")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No authentication claims"})
		return
	}

	jwtClaims := claims.(jwt.MapClaims)
	username := jwtClaims["username"].(string)

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Verify current password directly against Samba (not through auth service)
	// This avoids auth service logout issues during password verification
	if !utils.VerifyCurrentPassword(username, req.CurrentPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Change password
	err := h.userService.ChangeUserPassword(username, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// UpdateProfile allows users to update their own profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Get username from JWT token
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No authentication claims"})
		return
	}

	jwtClaims := claims.(jwt.MapClaims)
	username := jwtClaims["username"].(string)

	var req struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Update profile
	err := h.userService.UpdateUserProfile(username, req.FullName, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}
