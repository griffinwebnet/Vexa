package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/griffinwebnet/vexa/api/models"
	"github.com/griffinwebnet/vexa/api/services"
)

// DomainHandler handles HTTP requests for domain operations
type DomainHandler struct {
	domainService *services.DomainService
}

// NewDomainHandler creates a new DomainHandler instance
func NewDomainHandler() *DomainHandler {
	return &DomainHandler{
		domainService: services.NewDomainService(),
	}
}

// DomainStatus returns the current status of the domain controller
func (h *DomainHandler) DomainStatus(c *gin.Context) {
	response, err := h.domainService.GetDomainStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetDomainInfo returns detailed domain information
func (h *DomainHandler) GetDomainInfo(c *gin.Context) {
	server := c.DefaultQuery("server", "127.0.0.1")

	info, err := h.domainService.GetDomainInfo(server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, info)
}

// ConfigureDomain updates domain configuration
func (h *DomainHandler) ConfigureDomain(c *gin.Context) {
	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	err := h.domainService.ConfigureDomain(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Domain configuration updated successfully",
	})
}

// ProvisionDomainWithOutput provisions a new domain with streaming CLI output
func (h *DomainHandler) ProvisionDomainWithOutput(c *gin.Context) {
	var req models.ProvisionDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Set up Server-Sent Events
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// Create a channel for CLI output
	outputChan := make(chan string, 100)

	// Start provisioning in a goroutine
	go func() {
		defer close(outputChan)
		err := h.domainService.ProvisionDomainWithOutput(req, outputChan)
		if err != nil {
			outputChan <- "ERROR: " + err.Error()
		}
	}()

	// Stream output to client
	for output := range outputChan {
		c.SSEvent("message", gin.H{
			"type":      "output",
			"content":   output,
			"timestamp": time.Now().Unix(),
		})
		c.Writer.Flush()
	}

	// Send completion event
	c.SSEvent("message", gin.H{
		"type":      "complete",
		"content":   "Domain provisioning completed",
		"timestamp": time.Now().Unix(),
	})
	c.Writer.Flush()
}
