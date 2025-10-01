package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vexa/api/models"
	"github.com/vexa/api/services"
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

// ProvisionDomain provisions a new Samba AD DC domain
func (h *DomainHandler) ProvisionDomain(c *gin.Context) {
	var req models.ProvisionDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	err := h.domainService.ProvisionDomain(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Domain provisioned successfully",
		"domain":  req.Domain,
		"realm":   req.Realm,
	})
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
