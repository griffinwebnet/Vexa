package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vexa/api/services"
)

// DNSHandler handles HTTP requests for DNS operations
type DNSHandler struct {
	dnsService *services.DNSService
}

// NewDNSHandler creates a new DNSHandler instance
func NewDNSHandler() *DNSHandler {
	return &DNSHandler{
		dnsService: services.NewDNSService(),
	}
}

// DNSStatus returns the current DNS server status and configuration
func (h *DNSHandler) DNSStatus(c *gin.Context) {
	status, err := h.dnsService.GetDNSStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// UpdateDNSForwarders updates the DNS forwarder configuration
func (h *DNSHandler) UpdateDNSForwarders(c *gin.Context) {
	var req struct {
		Primary   string `json:"primary" binding:"required"`
		Secondary string `json:"secondary" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	err := h.dnsService.UpdateDNSForwarders(req.Primary, req.Secondary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "DNS forwarders updated successfully",
		"forwarders": []string{req.Primary, req.Secondary},
	})
}

// DNS handlers for managing DNS zones and records

func ListDNSZones(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "DNS zone listing not implemented yet",
	})
}

func ListDNSRecords(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "DNS record listing not implemented yet",
	})
}

func CreateDNSRecord(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "DNS record creation not implemented yet",
	})
}

func DeleteDNSRecord(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "DNS record deletion not implemented yet",
	})
}
