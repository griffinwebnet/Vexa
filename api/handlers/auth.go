package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vexa/api/utils"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Username  string    `json:"username"`
	IsAdmin   bool      `json:"is_admin"`
}

// Login authenticates user against PAM or Active Directory
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Authenticate via PAM
	authenticated, isAdmin := authenticateUser(req.Username, req.Password)
	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid username or password",
		})
		return
	}

	// Generate JWT token
	expiresAt := time.Now().Add(24 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.Username,
		"user_id":  req.Username, // TODO: Get actual user ID from system
		"is_admin": isAdmin,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(utils.GetJWTSecret()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt,
		Username:  req.Username,
		IsAdmin:   isAdmin,
	})
}

// authenticateUser validates credentials against PAM or dev mode
func authenticateUser(username, password string) (bool, bool) {
	// Development mode bypass (for Windows/testing)
	if os.Getenv("ENV") == "development" {
		// Allow any login in dev mode
		// In production on Linux, this will use real PAM auth
		if username != "" && password != "" {
			// Mock admin user for testing
			isAdmin := username == "admin" || username == "root"
			return true, isAdmin
		}
		return false, false
	}

	// Production mode: Use PAM authentication (Linux only)
	// Note: This will only compile/work on Linux systems
	isAdmin := utils.IsUserAdmin(username)
	return utils.AuthenticatePAM(username, password), isAdmin
}

