package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vexa/api/services"
)

// OverlayHandler handles HTTP requests for overlay networking
type OverlayHandler struct {
	overlayService *services.OverlayService
}

// NewOverlayHandler creates a new OverlayHandler
func NewOverlayHandler() *OverlayHandler {
	return &OverlayHandler{
		overlayService: services.NewOverlayService(),
	}
}

// SetupOverlay configures overlay networking
func (h *OverlayHandler) SetupOverlay(c *gin.Context) {
	var req struct {
		FQDN string `json:"fqdn" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	if err := h.overlayService.SetupOverlay(req.FQDN); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Overlay networking configured successfully",
	})
}

// GetOverlayStatus returns overlay networking status
func (h *OverlayHandler) GetOverlayStatus(c *gin.Context) {
	status, err := h.overlayService.GetOverlayStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// TestFQDN tests if an FQDN is publicly accessible
func (h *OverlayHandler) TestFQDN(c *gin.Context) {
	var req struct {
		FQDN string `json:"fqdn" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	result, err := h.overlayService.TestFQDNAccessibility(req.FQDN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// AddMachine generates scripts for joining a new machine
func (h *OverlayHandler) AddMachine(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	scripts, err := h.overlayService.AddMachine(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Machine added successfully",
		"scripts": scripts,
	})
}
