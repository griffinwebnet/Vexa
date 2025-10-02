package exec

import (
	"encoding/json"
	"os/exec"
)

// HeadscaleTool provides an interface for executing headscale commands
type HeadscaleTool struct{}

// NewHeadscaleTool creates a new HeadscaleTool instance
func NewHeadscaleTool() *HeadscaleTool {
	return &HeadscaleTool{}
}

// Run executes a headscale command with the given arguments
func (h *HeadscaleTool) Run(args ...string) (string, error) {
	cmd := exec.Command("headscale", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// PreAuthKey represents a pre-auth key response from headscale
type PreAuthKey struct {
	User       string   `json:"user"`
	ID         string   `json:"id"`
	Key        string   `json:"key"`
	Reusable   bool     `json:"reusable"`
	Ephemeral  bool     `json:"ephemeral"`
	Used       bool     `json:"used"`
	Expiration string   `json:"expiration"`
	CreatedAt  string   `json:"created_at"`
	ACLTags    []string `json:"acl_tags"`
}

// CreatePreAuthKey creates a new pre-auth key
func (h *HeadscaleTool) CreatePreAuthKey(user string, reusable bool, ephemeral bool) (*PreAuthKey, error) {
	args := []string{"preauthkeys", "create", "--user", user, "--output", "json"}

	if reusable {
		args = append(args, "--reusable")
	}
	if ephemeral {
		args = append(args, "--ephemeral")
	}

	output, err := h.Run(args...)
	if err != nil {
		return nil, err
	}

	var key PreAuthKey
	err = json.Unmarshal([]byte(output), &key)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

// ListPreAuthKeys lists all pre-auth keys for a user
func (h *HeadscaleTool) ListPreAuthKeys(user string) ([]PreAuthKey, error) {
	args := []string{"preauthkeys", "list", "--user", user, "--output", "json"}

	output, err := h.Run(args...)
	if err != nil {
		return nil, err
	}

	var keys []PreAuthKey
	err = json.Unmarshal([]byte(output), &keys)
	if err != nil {
		return nil, err
	}

	return keys, nil
}

// GetStatus returns the current headscale status
func (h *HeadscaleTool) GetStatus() (string, error) {
	return h.Run("status")
}

// ListUsers lists all users in headscale
func (h *HeadscaleTool) ListUsers() (string, error) {
	return h.Run("users", "list")
}

// IsEnabled checks if headscale is available and configured
func (h *HeadscaleTool) IsEnabled() bool {
	_, err := h.GetStatus()
	return err == nil
}
