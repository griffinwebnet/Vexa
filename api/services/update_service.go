package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

const (
	// Update channels
	ChannelStable  = "stable"  // Only tagged releases
	ChannelNightly = "nightly" // Latest main branch
)

// UpdateProgress represents the current state of an update
type UpdateProgress struct {
	Status    string `json:"status"`    // current status (downloading, building, etc)
	Progress  int    `json:"progress"`  // progress percentage
	Error     string `json:"error"`     // error message if any
	Completed bool   `json:"completed"` // whether update is complete
}

// StartSystemUpdate starts a system update via the CLI tool
func StartSystemUpdate(buildFromSource bool) error {
	args := []string{"update", "start"}
	if buildFromSource {
		args = append(args, "--build-source")
	}

	cmd := exec.Command("vexa", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start update: %v", err)
	}

	return nil
}

// GetUpdateStatus returns the current update status
func GetUpdateStatus() (*UpdateProgress, error) {
	cmd := exec.Command("vexa", "update", "status", "--json")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get update status: %v", err)
	}

	var progress UpdateProgress
	if err := json.Unmarshal(out.Bytes(), &progress); err != nil {
		return nil, fmt.Errorf("failed to parse update status: %v", err)
	}

	return &progress, nil
}

// GetUpdateLog returns the update log
func GetUpdateLog() (string, error) {
	cmd := exec.Command("vexa", "update", "log")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get update log: %v", err)
	}

	return out.String(), nil
}
