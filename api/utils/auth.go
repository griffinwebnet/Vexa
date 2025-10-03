package utils

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// AuthenticatePAM authenticates against system PAM authentication
func AuthenticatePAM(username, password string) bool {
	// Use pamtester to test PAM authentication
	// This requires pamtester to be installed: sudo apt-get install libpam0g-dev pamtester
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pamtester", "login", username, "authenticate")
	cmd.Stdin = strings.NewReader(password + "\n")

	err := cmd.Run()
	return err == nil
}

// AuthenticateSAMBA authenticates against SAMBA/Active Directory
func AuthenticateSAMBA(username, password string) bool {
	// Use smbclient to authenticate against the domain
	// This requires Samba client tools
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try multiple authentication methods for domain users
	targets := []string{
		"//localhost/netlogon",
		"//localhost/ipc$",
		"//localhost/c$",
	}

	for _, target := range targets {
		cmd := exec.CommandContext(ctx, "smbclient", target, "-U", username+"%"+password, "-c", "exit")
		err := cmd.Run()
		if err == nil {
			return true
		}
	}

	// Also try with domain prefix if username doesn't already have it
	if !strings.Contains(username, "\\") && !strings.Contains(username, "@") {
		// Try with domain prefix
		cmd := exec.CommandContext(ctx, "smbclient", "//localhost/netlogon", "-U", "VFW5788\\"+username+"%"+password, "-c", "exit")
		err := cmd.Run()
		if err == nil {
			return true
		}
	}

	return false
}
