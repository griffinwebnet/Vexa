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

	target := "//localhost/netlogon"
	cmd := exec.CommandContext(ctx, "smbclient", target, "-U", username+"%"+password, "-c", "exit")

	err := cmd.Run()
	return err == nil
}
