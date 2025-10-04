package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vexa/api/models"
	"github.com/vexa/api/services"
)

// AuthHandler handles HTTP requests for authentication operations
type AuthHandler struct {
	authService *services.AuthService
	vexaAdmin   *services.VexaAdminService
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(devMode bool) *AuthHandler {
	return &AuthHandler{
		authService: services.NewAuthService(devMode),
		vexaAdmin:   services.NewVexaAdminService(),
	}
}

// Login authenticates user against PAM or Active Directory
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// If system is unprovisioned and Vexa admin is initialized, allow Vexa login only
	domain := services.NewDomainService()
	status, _ := domain.GetDomainStatus()

	if status != nil && !status.Provisioned && h.vexaAdmin.IsInitialized() {
		if req.Username == "vexa" && h.vexaAdmin.Verify(req.Username, req.Password) {
			// Generate token as admin
			loginResponse, err := h.authService.GenerateToken(req.Username, true)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
				return
			}
			c.JSON(http.StatusOK, loginResponse)
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Authenticate user via standard flow
	authResult, err := h.authService.Authenticate(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Authentication failed",
		})
		return
	}

	if !authResult.Authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid username or password",
		})
		return
	}

	// Generate JWT token
	loginResponse, err := h.authService.GenerateToken(req.Username, authResult.IsAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, loginResponse)
}

// BootstrapStatus returns whether the Vexa admin password is set
func (h *AuthHandler) BootstrapStatus(c *gin.Context) {
	initialized := services.NewVexaAdminService().IsInitialized()
	c.JSON(http.StatusOK, gin.H{"initialized": initialized})
}

// BootstrapAdmin sets the initial Vexa admin password
func (h *AuthHandler) BootstrapAdmin(c *gin.Context) {
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	domain := services.NewDomainService()
	status, _ := domain.GetDomainStatus()
	if status != nil && status.Provisioned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already provisioned"})
		return
	}
	if err := services.NewVexaAdminService().SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Vexa admin initialized"})
}
