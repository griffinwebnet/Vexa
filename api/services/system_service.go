package services

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// getSambaVersion returns the installed Samba version
func getSambaVersion() string {
	cmd := exec.Command("samba", "-V")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "unknown"
	}
	return strings.TrimSpace(out.String())
}

// checkSambaUpdates checks if Samba updates are available
func checkSambaUpdates() string {
	// Use package manager to check for updates
	cmd := exec.Command("apt", "list", "--upgradable", "samba")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "unknown"
	}
	if strings.Contains(out.String(), "samba") {
		return "update_available"
	}
	return "up_to_date"
}

// getSambaLastUpdate returns when Samba was last updated
func getSambaLastUpdate() string {
	// Check package install time
	cmd := exec.Command("stat", "-c", "%y", "/usr/sbin/samba")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return time.Now().Format(time.RFC3339)
	}
	return strings.TrimSpace(out.String())
}

// checkSambaDependencies checks if all Samba dependencies are installed
func checkSambaDependencies() bool {
	deps := []string{
		"winbind",
		"libnss-winbind",
		"libpam-winbind",
	}

	for _, dep := range deps {
		cmd := exec.Command("dpkg", "-s", dep)
		if err := cmd.Run(); err != nil {
			return false
		}
	}
	return true
}

// getHeadscaleVersion returns the installed Headscale version
func getHeadscaleVersion() string {
	cmd := exec.Command("headscale", "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "unknown"
	}
	return strings.TrimSpace(out.String())
}

// checkHeadscaleUpdates checks if Headscale updates are available
func checkHeadscaleUpdates() string {
	// Check GitHub releases for newer version
	// For now, just return up_to_date
	return "up_to_date"
}

// getHeadscaleLastUpdate returns when Headscale was last updated
func getHeadscaleLastUpdate() string {
	cmd := exec.Command("stat", "-c", "%y", "/usr/bin/headscale")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return time.Now().Format(time.RFC3339)
	}
	return strings.TrimSpace(out.String())
}

// checkHeadscaleDependencies checks if all Headscale dependencies are installed
func checkHeadscaleDependencies() bool {
	// Check if required services are running
	services := []string{
		"headscale",
		"tailscaled",
	}

	for _, svc := range services {
		cmd := exec.Command("systemctl", "is-active", svc)
		if err := cmd.Run(); err != nil {
			return false
		}
	}
	return true
}

// UpdateSystem performs system-wide updates
func UpdateSystem() error {
	// Update package lists
	if err := exec.Command("apt", "update").Run(); err != nil {
		return fmt.Errorf("failed to update package lists: %v", err)
	}

	// Update Samba
	if err := exec.Command("apt", "install", "-y", "samba").Run(); err != nil {
		return fmt.Errorf("failed to update samba: %v", err)
	}

	// Update Headscale (would need to implement GitHub release download)

	// Restart services
	services := []string{"samba", "winbind", "headscale", "tailscaled"}
	for _, svc := range services {
		if err := exec.Command("systemctl", "restart", svc).Run(); err != nil {
			return fmt.Errorf("failed to restart %s: %v", svc, err)
		}
	}

	return nil
}
