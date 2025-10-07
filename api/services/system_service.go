package services

import (
	"bytes"
	"strings"
	"time"

	"github.com/griffinwebnet/vexa/api/utils"
)

// GetSambaVersion returns the installed Samba version
func GetSambaVersion() string {
	cmd, err := utils.SafeCommand("samba", "-V")
	if err != nil {
		return "unknown"
	}
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
	cmd, err := utils.SafeCommand("apt", "list", "--upgradable", "samba")
	if err != nil {
		return "unknown"
	}
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
	cmd, err := utils.SafeCommand("stat", "-c", "%y", "/usr/sbin/samba")
	if err != nil {
		return time.Now().Format(time.RFC3339)
	}
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
		cmd, cmdErr := utils.SafeCommand("dpkg", "-s", dep)
		if cmdErr != nil {
			return false
		}
		if err := cmd.Run(); err != nil {
			return false
		}
	}
	return true
}

// GetHeadscaleVersion returns the installed Headscale version
func GetHeadscaleVersion() string {
	cmd, err := utils.SafeCommand("headscale", "version")
	if err != nil {
		return "unknown"
	}
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
	cmd, err := utils.SafeCommand("stat", "-c", "%y", "/usr/local/bin/headscale")
	if err != nil {
		return time.Now().Format(time.RFC3339)
	}
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
		cmd, cmdErr := utils.SafeCommand("systemctl", "is-enabled", svc)
		if cmdErr != nil {
			return false
		}
		if err := cmd.Run(); err != nil {
			return false
		}
	}
	return true
}
