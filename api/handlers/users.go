package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vexa/api/models"
	"github.com/vexa/api/services"
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
