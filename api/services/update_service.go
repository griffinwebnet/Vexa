package services

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/griffinwebnet/vexa/api/utils"
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

// StartSystemUpdate starts a system update via update.sh script
func StartSystemUpdate(buildFromSource bool) error {
	utils.Info("Starting system update via update.sh, buildFromSource: %v", buildFromSource)

	args := []string{"./update.sh"}
	if buildFromSource {
		args = append(args, "--nightly", "--fast")
	} else {
		args = append(args, "--nightly")
	}

	cmd, err := utils.SafeCommand("bash", args...)
	if err != nil {
		utils.Error("Command sanitization failed for update.sh: %v", err)
		return fmt.Errorf("command sanitization failed: %v", err)
	}

	// Run in background and capture output
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Start(); err != nil {
		utils.Error("Failed to start update.sh: %v", err)
		return fmt.Errorf("failed to start update: %v", err)
	}

	utils.Info("System update started successfully, PID: %d", cmd.Process.Pid)
	return nil
}

// GetUpdateStatus returns the current update status by checking if update.sh is running
func GetUpdateStatus() (*UpdateProgress, error) {
	utils.Info("Getting update status")

	// Check if update.sh is currently running
	cmd, err := utils.SafeCommand("pgrep", "-f", "update.sh")
	if err != nil {
		utils.Error("Command sanitization failed for pgrep: %v", err)
		return nil, fmt.Errorf("command sanitization failed: %v", err)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()

	progress := &UpdateProgress{
		Status:    "idle",
		Progress:  0,
		Completed: true,
	}

	if err != nil {
		// pgrep returns non-zero exit code when no processes found
		// This means update.sh is not running
		utils.Info("No update.sh process found, system is idle")
	} else {
		// update.sh is running
		progress.Status = "updating"
		progress.Progress = 50 // We don't know exact progress, so estimate
		progress.Completed = false
		utils.Info("Update.sh process found, update in progress")
	}

	return progress, nil
}

// GetUpdateLog returns the update log from the most recent update.sh run
func GetUpdateLog() (string, error) {
	utils.Info("Getting update log")

	// Try to get the log from the most recent update.sh run
	// This could be from a log file or by checking the last few lines of system logs
	cmd, err := utils.SafeCommand("tail", "-n", "100", "/var/log/vexa/vexa.log")
	if err != nil {
		utils.Error("Command sanitization failed for tail: %v", err)
		return "", fmt.Errorf("command sanitization failed: %v", err)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		utils.Error("Failed to get update log: %v", err)
		return "", fmt.Errorf("failed to get update log: %v", err)
	}

	logContent := out.String()

	// Filter for update-related log entries
	lines := strings.Split(logContent, "\n")
	var updateLines []string
	for _, line := range lines {
		if strings.Contains(line, "update") || strings.Contains(line, "Update") {
			updateLines = append(updateLines, line)
		}
	}

	filteredLog := strings.Join(updateLines, "\n")
	utils.Info("Update log retrieved successfully, length: %d characters", len(filteredLog))
	return filteredLog, nil
}
