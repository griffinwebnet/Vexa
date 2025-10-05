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
	authenticated, isAdmin, isDomainUser := s.authenticateUser(req.Username, req.Password)

	return &models.AuthResult{
		Authenticated: authenticated,
		IsAdmin:       isAdmin,
		IsDomainUser:  isDomainUser,
		Error:         nil,
	}, nil
}

// GenerateToken generates a JWT token for the authenticated user
func (s *AuthService) GenerateToken(username string, isAdmin bool, isDomainUser bool) (*models.LoginResponse, error) {
	fmt.Printf("DEBUG: GenerateToken called for user: %s, isAdmin: %v, isDomainUser: %v\n", username, isAdmin, isDomainUser)

	expiresAt := time.Now().Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"username":       username,
		"user_id":        username, // TODO: Get actual user ID from system
		"is_admin":       isAdmin,
		"is_domain_user": isDomainUser,
		"exp":            expiresAt.Unix(),
		"iat":            time.Now().Unix(),
	}

	fmt.Printf("DEBUG: JWT claims: %+v\n", claims)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(utils.GetJWTSecret()))
	if err != nil {
		fmt.Printf("DEBUG: Failed to sign JWT token: %v\n", err)
		return nil, err
	}

	response := &models.LoginResponse{
		Token:        tokenString,
		ExpiresAt:    expiresAt,
		Username:     username,
		IsAdmin:      isAdmin,
		IsDomainUser: isDomainUser,
	}

	fmt.Printf("DEBUG: Generated login response: %+v\n", response)
	return response, nil
}

// authenticateUser validates credentials against SAMBA domain first, then PAM fallback
// Returns: (authenticated, isAdmin, isDomainUser)
func (s *AuthService) authenticateUser(username, password string) (bool, bool, bool) {
	fmt.Printf("DEBUG: Attempting authentication for user: %s\n", username)

	// Only allow test credentials in dev mode
	if s.devMode {
		// Development mode: allow test credentials
		if username == "admin" && password == "admin" {
			fmt.Printf("DEBUG: Dev mode authentication successful for admin\n")
			return true, true, false // Local admin, not domain user
		}
		if username == "user" && password == "user" {
			fmt.Printf("DEBUG: Dev mode authentication successful for user\n")
			return true, false, false // Local user, not domain user
		}
	}

	// Production mode: Try SAMBA domain authentication first
	fmt.Printf("DEBUG: Trying SAMBA authentication for %s\n", username)
	if utils.AuthenticateSAMBA(username, password) {
		// SAMBA authentication successful - this is a domain user
		fmt.Printf("DEBUG: SAMBA authentication successful for %s\n", username)

		// Check if user is in Domain Admins group for admin status
		isAdmin := utils.CheckDomainAdminStatus(username)
		fmt.Printf("DEBUG: User %s domain admin status: %v\n", username, isAdmin)
		return true, isAdmin, true // Domain user
	}
	fmt.Printf("DEBUG: SAMBA authentication failed for %s\n", username)

	// Fallback to PAM for local system users
	fmt.Printf("DEBUG: Trying PAM authentication for %s\n", username)
	if utils.AuthenticatePAM(username, password) {
		// PAM authentication successful for system users
		fmt.Printf("DEBUG: PAM authentication successful for %s\n", username)
		// Determine admin via sudoers groups/root
		isAdmin := utils.CheckLocalAdminStatus(username)
		fmt.Printf("DEBUG: Local admin (sudoers) status for %s: %v\n", username, isAdmin)

		// Local users are only authorized if they are admins (root or sudoers)
		if isAdmin {
			return true, true, false // Local admin user
		} else {
			// Local non-admin users are not authorized
			fmt.Printf("DEBUG: Local user %s is not authorized (not admin/sudoer)\n", username)
			return false, false, false
		}
	}
	fmt.Printf("DEBUG: PAM authentication failed for %s\n", username)

	// Authentication failed
	fmt.Printf("DEBUG: All authentication methods failed for %s\n", username)
	return false, false, false
}
