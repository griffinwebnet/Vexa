package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/griffinwebnet/vexa/api/services"
)

// ComputerHandler handles HTTP requests for computer operations
type ComputerHandler struct {
	computerService *services.ComputerService
}

// NewComputerHandler creates a new ComputerHandler instance
func NewComputerHandler() *ComputerHandler {
	return &ComputerHandler{
		computerService: services.NewComputerService(),
	}
}

// ListComputers returns all computers/devices in the domain with connection status
func (h *ComputerHandler) ListComputers(c *gin.Context) {
	computers, err := h.computerService.ListComputers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"computers": computers,
		"count":     len(computers),
	})
}

// GetComputer returns details for a specific computer
func (h *ComputerHandler) GetComputer(c *gin.Context) {
	computerName := c.Param("id")

	computer, err := h.computerService.GetComputer(computerName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Computer not found",
		})
		return
	}

	c.JSON(http.StatusOK, computer)
}

// GetMachineDetails returns detailed information about a specific machine
func (h *ComputerHandler) GetMachineDetails(c *gin.Context) {
	machineId := c.Param("id")

	details, err := h.computerService.GetMachineDetails(machineId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Machine not found",
		})
		return
	}

	c.JSON(http.StatusOK, details)
}

// DeleteComputer removes a computer from the domain
func (h *ComputerHandler) DeleteComputer(c *gin.Context) {
	computerName := c.Param("id")

	err := h.computerService.DeleteComputer(computerName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Computer deleted successfully",
		"name":    computerName,
	})
}
