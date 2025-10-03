package services

import (
	"bytes"
	"os/exec"
	"strings"
	"time"
)

// GetSambaVersion returns the installed Samba version
func GetSambaVersion() string {
	cmd := exec.Command("samba", "-V")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "unknown"
	}
	return strings.TrimSpace(out.String())
}

// CheckSambaUpdates checks if Samba updates are available
func CheckSambaUpdates() string {
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

// GetSambaLastUpdate returns when Samba was last updated
func GetSambaLastUpdate() string {
	// Check package install time
	cmd := exec.Command("stat", "-c", "%y", "/usr/sbin/samba")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return time.Now().Format(time.RFC3339)
	}
	return strings.TrimSpace(out.String())
}

// CheckSambaDependencies checks if all Samba dependencies are installed
func CheckSambaDependencies() bool {
	deps := []string{
		"winbind",
		"krb5-user",
	}

	for _, dep := range deps {
		cmd := exec.Command("dpkg", "-s", dep)
		if err := cmd.Run(); err != nil {
			return false
		}
	}
	return true
}

// GetHeadscaleVersion returns the installed Headscale version
func GetHeadscaleVersion() string {
	cmd := exec.Command("headscale", "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "unknown"
	}
	return strings.TrimSpace(out.String())
}

// CheckHeadscaleUpdates checks if Headscale updates are available
func CheckHeadscaleUpdates() string {
	// For now, just return up_to_date
	return "up_to_date"
}

// GetHeadscaleLastUpdate returns when Headscale was last updated
func GetHeadscaleLastUpdate() string {
	cmd := exec.Command("stat", "-c", "%y", "/usr/local/bin/headscale")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return time.Now().Format(time.RFC3339)
	}
	return strings.TrimSpace(out.String())
}

// CheckHeadscaleDependencies checks if all Headscale dependencies are installed
func CheckHeadscaleDependencies() bool {
	// Check if required services are available
	services := []string{
		"headscale",
		"tailscaled",
	}

	for _, svc := range services {
		cmd := exec.Command("systemctl", "is-enabled", svc)
		if err := cmd.Run(); err != nil {
			return false
		}
	}
	return true
}
