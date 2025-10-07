package services

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/griffinwebnet/vexa/api/models"
	"github.com/griffinwebnet/vexa/api/utils"
)

// AuthService handles authentication-related business logic
type AuthService struct {
}

// NewAuthService creates a new AuthService instance
func NewAuthService() *AuthService {
	return &AuthService{}
}

// Authenticate validates user credentials
func (s *AuthService) Authenticate(req models.LoginRequest) (*models.AuthResult, error) {
	utils.Info("Authentication attempt for user: %s", req.Username)

	authenticated, isAdmin, isDomainUser := s.authenticateUser(req.Username, req.Password)

	if authenticated {
		utils.Info("Authentication successful for user: %s (admin: %v, domain: %v)", req.Username, isAdmin, isDomainUser)
	} else {
		utils.Warn("Authentication failed for user: %s", req.Username)
	}

	return &models.AuthResult{
		Authenticated: authenticated,
		IsAdmin:       isAdmin,
		IsDomainUser:  isDomainUser,
		Error:         nil,
	}, nil
}

// GenerateToken generates a JWT token for the authenticated user
func (s *AuthService) GenerateToken(username string, isAdmin bool, isDomainUser bool) (*models.LoginResponse, error) {
	utils.Info("Generating JWT token for user: %s", username)

	expiresAt := time.Now().Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"username":       username,
		"user_id":        username, // TODO: Get actual user ID from system
		"is_admin":       isAdmin,
		"is_domain_user": isDomainUser,
		"exp":            expiresAt.Unix(),
		"iat":            time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(utils.GetJWTSecret()))
	if err != nil {
		utils.Error("Failed to generate JWT token for user %s: %v", username, err)
		return nil, err
	}

	utils.Info("JWT token generated successfully for user: %s (expires: %v)", username, expiresAt)

	response := &models.LoginResponse{
		Token:        tokenString,
		ExpiresAt:    expiresAt,
		Username:     username,
		IsAdmin:      isAdmin,
		IsDomainUser: isDomainUser,
	}

	return response, nil
}

// authenticateUser validates credentials against SAMBA domain first, then PAM fallback
// Returns: (authenticated, isAdmin, isDomainUser)
func (s *AuthService) authenticateUser(username, password string) (bool, bool, bool) {
	utils.Debug("Attempting authentication for user: %s", username)

	// Try SAMBA domain authentication first
	if utils.AuthenticateSAMBA(username, password) {
		// SAMBA authentication successful - this is a domain user
		utils.Info("SAMBA authentication successful for user: %s", username)
		// Check if user is in Domain Admins group for admin status
		isAdmin := utils.CheckDomainAdminStatus(username)
		utils.Debug("User %s domain admin status: %v", username, isAdmin)
		return true, isAdmin, true // Domain user
	}

	// Fallback to PAM for local system users
	if utils.AuthenticatePAM(username, password) {
		// PAM authentication successful for system users
		utils.Info("PAM authentication successful for user: %s", username)
		// Determine admin via sudoers groups/root
		isAdmin := utils.CheckLocalAdminStatus(username)
		utils.Debug("Local admin status for %s: %v", username, isAdmin)

		// Local users are only authorized if they are admins (root or sudoers)
		if isAdmin {
			return true, true, false // Local admin user
		} else {
			// Local non-admin users are not authorized
			utils.Warn("Local user %s is not authorized (not admin/sudoer)", username)
			return false, false, false
		}
	}

	// Authentication failed
	utils.Warn("All authentication methods failed for user: %s", username)
	return false, false, false
}
