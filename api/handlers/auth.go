package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vexa/api/models"
	"github.com/vexa/api/services"
	"github.com/vexa/api/utils"
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

	// Check domain status first
	domain := services.NewDomainService()
	status, _ := domain.GetDomainStatus()

	// If no domain exists, only allow local admin users
	if status == nil || !status.Provisioned {
		// Try PAM authentication for local admin users only
		if utils.AuthenticatePAM(req.Username, req.Password) {
			isAdmin := utils.CheckLocalAdminStatus(req.Username)
			if isAdmin {
				// Local admin user - allow them to proceed to setup
				loginResponse, err := h.authService.GenerateToken(req.Username, true, false)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Failed to generate token",
					})
					return
				}
				c.JSON(http.StatusOK, loginResponse)
				return
			}
		}

		// Not a local admin or authentication failed
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "System not configured. Only local administrators can perform initial setup.",
		})
		return
	}

	// Domain exists - try domain authentication first, then PAM fallback
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
