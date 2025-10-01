package handlers

import (
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

type Group struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Members     []string `json:"members"`
}

type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// ListGroups returns all groups in the domain
func ListGroups(c *gin.Context) {
	// Dev mode: Return dummy data
	if os.Getenv("ENV") == "development" {
		groups := []Group{
			{Name: "Domain Admins", Description: "Domain administrators"},
			{Name: "Domain Users", Description: "All domain users"},
			{Name: "IT Staff", Description: "IT department"},
			{Name: "Finance", Description: "Finance department"},
			{Name: "HR", Description: "Human resources"},
		}
		c.JSON(http.StatusOK, gin.H{
			"groups": groups,
			"count":  len(groups),
		})
		return
	}

	cmd := exec.Command("samba-tool", "group", "list")
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list groups",
			"details": string(output),
		})
		return
	}

	// Parse group list
	groupNames := strings.Split(strings.TrimSpace(string(output)), "\n")
	groups := make([]Group, 0, len(groupNames))

	for _, name := range groupNames {
		if name != "" {
			groups = append(groups, Group{
				Name: name,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
		"count":  len(groups),
	})
}

// CreateGroup creates a new group in the domain
func CreateGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	args := []string{"group", "add", req.Name}
	if req.Description != "" {
		args = append(args, "--description="+req.Description)
	}

	cmd := exec.Command("samba-tool", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create group",
			"details": string(output),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Group created successfully",
		"name":    req.Name,
	})
}

// GetGroup returns details for a specific group
func GetGroup(c *gin.Context) {
	groupName := c.Param("id")

	// Get group members
	cmd := exec.Command("samba-tool", "group", "listmembers", groupName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Group not found",
		})
		return
	}

	members := strings.Split(strings.TrimSpace(string(output)), "\n")

	c.JSON(http.StatusOK, Group{
		Name:    groupName,
		Members: members,
	})
}

// UpdateGroup updates an existing group
func UpdateGroup(c *gin.Context) {
	// TODO: Implement group update logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "Group update not implemented",
	})
}

// DeleteGroup removes a group from the domain
func DeleteGroup(c *gin.Context) {
	groupName := c.Param("id")

	cmd := exec.Command("samba-tool", "group", "delete", groupName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete group",
			"details": string(output),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Group deleted successfully",
		"name":    groupName,
	})
}
