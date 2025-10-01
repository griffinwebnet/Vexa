package services

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vexa/api/models"
	"github.com/vexa/api/utils"
)

// AuthService handles authentication-related business logic
type AuthService struct{}

// NewAuthService creates a new AuthService instance
func NewAuthService() *AuthService {
	return &AuthService{}
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

// authenticateUser validates credentials against PAM or dev mode
func (s *AuthService) authenticateUser(username, password string) (bool, bool) {
	// TODO: Implement actual PAM authentication
	// For now, use simple dev authentication
	if username == "admin" && password == "admin" {
		return true, true
	}
	if username == "user" && password == "user" {
		return true, false
	}
	return false, false
}
