package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	vexaexec "github.com/vexa/api/exec"
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
	// List all pre-auth keys for the infrastructure user
	cmd := exec.Command("headscale", "preauthkeys", "list", "-u", "infrastructure", "-c", "/etc/headscale/config.yaml", "-o", "json")
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
			if keyStr, ok := key["key"].(string); ok && keyStr != "" {
				return keyStr, nil
			}
		}
	}

	return "", fmt.Errorf("no reusable infrastructure key found")
}

// IsEnabled checks if Headscale is available and configured
func (s *HeadscaleService) IsEnabled() bool {
	if os.Getenv("ENV") == "development" {
		return true // Always enabled in development mode
	}

	// Check if Headscale service is running
	cmd := exec.Command("systemctl", "is-active", "headscale")
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
