package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vexa/api/services"
)

// SystemHandler handles HTTP requests for system operations
type SystemHandler struct {
	systemService *services.SystemService
}

// NewSystemHandler creates a new SystemHandler instance
func NewSystemHandler() *SystemHandler {
	return &SystemHandler{
		systemService: services.NewSystemService(),
	}
}

// SystemStatus returns system information
func (h *SystemHandler) SystemStatus(c *gin.Context) {
	response := h.systemService.GetSystemStatus()
	c.JSON(http.StatusOK, response)
}

// GetServiceStatus checks the status of a system service
func (h *SystemHandler) GetServiceStatus(c *gin.Context) {
	serviceName := c.Param("name")

	status, err := h.systemService.GetServiceStatus(serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}
