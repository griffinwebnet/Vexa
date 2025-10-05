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
	// Check domain status first - if no domain, bypass authentication entirely
	domain := services.NewDomainService()
	status, _ := domain.GetDomainStatus()

	// If no domain is provisioned, bypass login and redirect to setup
	if status == nil || !status.Provisioned {
		c.JSON(http.StatusOK, gin.H{
			"requires_setup": true,
			"message":        "Domain not configured. Please run initial setup.",
		})
		return
	}

	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Authenticate user via standard flow (domain first, then PAM fallback)
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

	// Check if user has proper authorization
	if !authResult.IsAdmin && !authResult.IsDomainUser {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User account not authorized",
		})
		return
	}

	// Generate JWT token
	loginResponse, err := h.authService.GenerateToken(req.Username, authResult.IsAdmin, authResult.IsDomainUser)
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
