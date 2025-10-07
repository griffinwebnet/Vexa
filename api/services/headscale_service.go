package services

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	vexaexec "github.com/griffinwebnet/vexa/api/exec"
	"github.com/griffinwebnet/vexa/api/utils"
)

// HeadscaleService handles Headscale-related business logic
type HeadscaleService struct {
	headscaleTool *vexaexec.HeadscaleTool
}

// NewHeadscaleService creates a new HeadscaleService instance
func NewHeadscaleService() *HeadscaleService {
	return &HeadscaleService{
		headscaleTool: vexaexec.NewHeadscaleTool(),
	}
}

// GetStatus returns the current Headscale status
func (s *HeadscaleService) GetStatus() (map[string]interface{}, error) {

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
func (s *HeadscaleService) CreatePreAuthKey(user string, reusable bool, ephemeral bool) (*vexaexec.PreAuthKey, error) {

	// Create actual pre-auth key
	key, err := s.headscaleTool.CreatePreAuthKey(user, reusable, ephemeral)
	if err != nil {
		return nil, fmt.Errorf("failed to create pre-auth key: %v", err)
	}

	return key, nil
}

// GetInfrastructureKey returns the existing infrastructure pre-auth key
func (s *HeadscaleService) GetInfrastructureKey() (string, error) {
	// Get the infrastructure user ID first
	userID, err := s.getInfrastructureUserID()
	if err != nil {
		return "", fmt.Errorf("failed to get infrastructure user ID: %v", err)
	}

	// List all pre-auth keys for the infrastructure user using the user ID
	cmd, cmdErr := utils.SafeCommand("headscale", "preauthkeys", "list", "-u", fmt.Sprintf("%d", userID), "-c", "/etc/headscale/config.yaml", "-o", "json")
	if cmdErr != nil {
		return "", fmt.Errorf("command sanitization failed: %v", cmdErr)
	}
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list pre-auth keys: %v", err)
	}

	// Parse JSON to find a reusable key
	var keys []map[string]interface{}
	if err := json.Unmarshal(output, &keys); err != nil {
		return "", fmt.Errorf("failed to parse pre-auth keys: %v", err)
	}

	// Find the first reusable key
	for _, key := range keys {
		if reusable, ok := key["reusable"].(bool); ok && reusable {
			// Reusable keys don't expire, so just return the key string
			if keyStr, ok := key["key"].(string); ok && keyStr != "" {
				return keyStr, nil
			}
		}
	}

	return "", fmt.Errorf("no reusable infrastructure key found - the Headscale setup may have failed")
}

// getInfrastructureUserID gets the numeric ID of the infrastructure user
func (s *HeadscaleService) getInfrastructureUserID() (int, error) {
	cmd, cmdErr := utils.SafeCommand("headscale", "users", "list", "-c", "/etc/headscale/config.yaml", "-o", "json")
	if cmdErr != nil {
		return 0, fmt.Errorf("command sanitization failed: %v", cmdErr)
	}
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to list users: %v", err)
	}

	// Parse users
	var users []map[string]interface{}
	if err := json.Unmarshal(output, &users); err != nil {
		return 0, fmt.Errorf("failed to parse users: %v", err)
	}

	// Find infrastructure user and get its ID
	for _, user := range users {
		if name, ok := user["name"].(string); ok && name == "infrastructure" {
			if id, ok := user["id"].(float64); ok {
				return int(id), nil
			}
		}
	}

	return 0, fmt.Errorf("infrastructure user not found")
}

// IsEnabled checks if Headscale is available and configured
func (s *HeadscaleService) IsEnabled() bool {
	if os.Getenv("ENV") == "development" {
		return true // Always enabled in development mode
	}

	// Check if Headscale service is running
	cmd, err := utils.SafeCommand("systemctl", "is-active", "headscale")
	if err != nil {
		return false
	}
	return cmd.Run() == nil
}

// GetLoginServerBase returns the Headscale login server base URL without the
// "/mesh" suffix. Preference order:
// 1) HEADSCALE_SERVER_URL env var
// 2) /etc/headscale/config.yaml server_url
// Returns empty string if not found.
func (s *HeadscaleService) GetLoginServerBase() string {
	// Prefer explicit env var
	if url := strings.TrimSpace(os.Getenv("HEADSCALE_SERVER_URL")); url != "" {
		return trimMeshAndSlash(url)
	}

	// Fallback: parse headscale config
	content, err := os.ReadFile("/etc/headscale/config.yaml")
	if err == nil {
		lines := strings.Split(string(content), "\n")
		for _, raw := range lines {
			line := strings.TrimSpace(raw)
			if strings.HasPrefix(line, "server_url:") {
				val := strings.TrimSpace(strings.TrimPrefix(line, "server_url:"))
				// remove optional quotes
				val = strings.Trim(val, "\"'")
				return trimMeshAndSlash(val)
			}
		}
	}

	return ""
}

// GetLoginServerFull returns the Headscale login-server URL suitable for
// passing directly to `tailscale up --login-server`. Preference order:
// 1) HEADSCALE_SERVER_URL env var (expected to include scheme and path)
// 2) /etc/headscale/config.yaml server_url
// Returns empty string if not found.
func (s *HeadscaleService) GetLoginServerFull() string {
	if url := strings.TrimSpace(os.Getenv("HEADSCALE_SERVER_URL")); url != "" {
		return url
	}

	content, err := os.ReadFile("/etc/headscale/config.yaml")
	if err == nil {
		lines := strings.Split(string(content), "\n")
		for _, raw := range lines {
			line := strings.TrimSpace(raw)
			if strings.HasPrefix(line, "server_url:") {
				val := strings.TrimSpace(strings.TrimPrefix(line, "server_url:"))
				val = strings.Trim(val, "\"'")
				return val
			}
		}
	}

	return ""
}

func trimMeshAndSlash(u string) string {
	// remove trailing slash first
	u = strings.TrimSpace(u)
	for strings.HasSuffix(u, "/") {
		u = strings.TrimSuffix(u, "/")
	}
	// then remove trailing /mesh if present
	if strings.HasSuffix(strings.ToLower(u), "/mesh") {
		u = u[:len(u)-len("/mesh")]
	}
	return u
}
