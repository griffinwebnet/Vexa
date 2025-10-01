package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vexa/api/models"
	"github.com/vexa/api/services"
)

// GroupHandler handles HTTP requests for group operations
type GroupHandler struct {
	groupService *services.GroupService
}

// NewGroupHandler creates a new GroupHandler instance
func NewGroupHandler() *GroupHandler {
	return &GroupHandler{
		groupService: services.NewGroupService(),
	}
}

// ListGroups returns all groups in the domain
func (h *GroupHandler) ListGroups(c *gin.Context) {
	groups, err := h.groupService.ListGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
		"count":  len(groups),
	})
}

// CreateGroup creates a new group in the domain
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req models.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	err := h.groupService.CreateGroup(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Group created successfully",
		"name":    req.Name,
	})
}

// GetGroup returns details for a specific group
func (h *GroupHandler) GetGroup(c *gin.Context) {
	groupName := c.Param("id")

	group, err := h.groupService.GetGroup(groupName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Group not found",
		})
		return
	}

	c.JSON(http.StatusOK, group)
}

// UpdateGroup updates an existing group
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	groupName := c.Param("id")

	var req models.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	err := h.groupService.UpdateGroup(groupName, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Group updated successfully",
		"name":    groupName,
	})
}

// DeleteGroup removes a group from the domain
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	groupName := c.Param("id")

	err := h.groupService.DeleteGroup(groupName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Group deleted successfully",
		"name":    groupName,
	})
}

// AddGroupMembers adds members to a group
func (h *GroupHandler) AddGroupMembers(c *gin.Context) {
	groupName := c.Param("id")

	var req models.AddGroupMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	err := h.groupService.AddGroupMembers(groupName, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Members added to group successfully",
		"name":    groupName,
	})
}

// RemoveGroupMembers removes members from a group
func (h *GroupHandler) RemoveGroupMembers(c *gin.Context) {
	groupName := c.Param("id")

	var req models.RemoveGroupMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	err := h.groupService.RemoveGroupMembers(groupName, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Members removed from group successfully",
		"name":    groupName,
	})
}
