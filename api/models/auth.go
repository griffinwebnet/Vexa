package models

import "time"

// LoginRequest represents the login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token        string    `json:"token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Username     string    `json:"username"`
	IsAdmin      bool      `json:"is_admin"`
	IsDomainUser bool      `json:"is_domain_user"`
}

// AuthResult represents the result of authentication
type AuthResult struct {
	Authenticated bool
	IsAdmin       bool
	IsDomainUser  bool
	Error         error
}
