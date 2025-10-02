package services

import (
	"fmt"
	"os"
	"time"

	"github.com/vexa/api/exec"
)

// HeadscaleService handles Headscale-related business logic
type HeadscaleService struct {
	headscaleTool *exec.HeadscaleTool
}

// NewHeadscaleService creates a new HeadscaleService instance
func NewHeadscaleService() *HeadscaleService {
	return &HeadscaleService{
		headscaleTool: exec.NewHeadscaleTool(),
	}
}

// GetStatus returns the current Headscale status
func (s *HeadscaleService) GetStatus() (map[string]interface{}, error) {
	if os.Getenv("ENV") == "development" {
		// Return dummy data for development
		return map[string]interface{}{
			"enabled": true,
			"status":  "running",
			"users":   []string{"admin", "user"},
			"nodes":   5,
		}, nil
	}

	// Check if Headscale is enabled
	if !s.headscaleTool.IsEnabled() {
		return map[string]interface{}{
			"enabled": false,
			"status":  "not_available",
		}, nil
	}

	// Get actual status
	status, err := s.headscaleTool.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get headscale status: %v", err)
	}

	users, err := s.headscaleTool.ListUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to list headscale users: %v", err)
	}

	return map[string]interface{}{
		"enabled":       true,
		"status":        "running",
		"status_output": status,
		"users_output":  users,
	}, nil
}

// CreatePreAuthKey creates a new pre-auth key for deployment
func (s *HeadscaleService) CreatePreAuthKey(user string, reusable bool, ephemeral bool) (*exec.PreAuthKey, error) {
	if os.Getenv("ENV") == "development" {
		// Return dummy pre-auth key for development
		return &exec.PreAuthKey{
			User:       user,
			ID:         "dummy-key-id",
			Key:        "tskey-auth-dummy123456789abcdef",
			Reusable:   reusable,
			Ephemeral:  ephemeral,
			Used:       false,
			Expiration: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			CreatedAt:  time.Now().Format(time.RFC3339),
			ACLTags:    []string{},
		}, nil
	}

	// Create actual pre-auth key
	key, err := s.headscaleTool.CreatePreAuthKey(user, reusable, ephemeral)
	if err != nil {
		return nil, fmt.Errorf("failed to create pre-auth key: %v", err)
	}

	return key, nil
}

// IsEnabled checks if Headscale is available and configured
func (s *HeadscaleService) IsEnabled() bool {
	if os.Getenv("ENV") == "development" {
		return true // Always enabled in development mode
	}

	return s.headscaleTool.IsEnabled()
}
