package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vexa/api/models"
	"github.com/vexa/api/utils"
)

// AuthService handles authentication-related business logic
type AuthService struct {
	devMode bool
}

// NewAuthService creates a new AuthService instance
func NewAuthService(devMode bool) *AuthService {
	return &AuthService{
		devMode: devMode,
	}
}

// Authenticate validates user credentials
func (s *AuthService) Authenticate(req models.LoginRequest) (*models.AuthResult, error) {
	authenticated, isAdmin := s.authenticateUser(req.Username, req.Password)

	return &models.AuthResult{
		Authenticated: authenticated,
		IsAdmin:       isAdmin,
		Error:         nil,
	}, nil
}

// GenerateToken generates a JWT token for the authenticated user
func (s *AuthService) GenerateToken(username string, isAdmin bool) (*models.LoginResponse, error) {
	expiresAt := time.Now().Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"username": username,
		"user_id":  username, // TODO: Get actual user ID from system
		"is_admin": isAdmin,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(utils.GetJWTSecret()))
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt,
		Username:  username,
		IsAdmin:   isAdmin,
	}, nil
}

// authenticateUser validates credentials against PAM, SAMBA, or dev mode
func (s *AuthService) authenticateUser(username, password string) (bool, bool) {
	fmt.Printf("DEBUG: Attempting authentication for user: %s\n", username)

	// Only allow test credentials in dev mode
	if s.devMode {
		// Development mode: allow test credentials
		if username == "admin" && password == "admin" {
			fmt.Printf("DEBUG: Dev mode authentication successful for admin\n")
			return true, true
		}
		if username == "user" && password == "user" {
			fmt.Printf("DEBUG: Dev mode authentication successful for user\n")
			return true, false
		}
	}

	// Production mode: Try SAMBA domain authentication first, then PAM
	fmt.Printf("DEBUG: Trying SAMBA authentication for %s\n", username)
	if utils.AuthenticateSAMBA(username, password) {
		// SAMBA authentication successful
		fmt.Printf("DEBUG: SAMBA authentication successful for %s\n", username)
		// TODO: Check if user is in Domain Admins group for admin status
		return true, true
	}
	fmt.Printf("DEBUG: SAMBA authentication failed for %s\n", username)

	fmt.Printf("DEBUG: Trying PAM authentication for %s\n", username)
	if utils.AuthenticatePAM(username, password) {
		// PAM authentication successful for system users
		fmt.Printf("DEBUG: PAM authentication successful for %s\n", username)
		// TODO: Implement proper admin group checking for PAM users
		return true, true
	}
	fmt.Printf("DEBUG: PAM authentication failed for %s\n", username)

	// Authentication failed
	fmt.Printf("DEBUG: All authentication methods failed for %s\n", username)
	return false, false
}
