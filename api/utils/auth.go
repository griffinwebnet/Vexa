package utils

import (
	"context"
	"fmt"
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
	fmt.Printf("DEBUG: Attempting SAMBA authentication for user: %s\n", username)

	// Use smbclient to authenticate against the domain
	// This requires Samba client tools
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get domain name dynamically
	domainName := getDomainName()
	fmt.Printf("DEBUG: Using domain name: %s\n", domainName)

	// Try multiple authentication methods for domain users
	targets := []string{
		"//localhost/netlogon",
		"//localhost/ipc$",
		"//localhost/c$",
	}

	// Try without domain prefix first
	for _, target := range targets {
		fmt.Printf("DEBUG: Trying smbclient %s with user %s\n", target, username)
		cmd := exec.CommandContext(ctx, "smbclient", target, "-U", username+"%"+password, "-c", "exit")
		err := cmd.Run()
		if err == nil {
			fmt.Printf("DEBUG: SAMBA authentication successful for %s\n", username)
			return true
		}
		fmt.Printf("DEBUG: smbclient failed for %s: %v\n", target, err)
	}

	// Try with domain prefix if username doesn't already have it
	if !strings.Contains(username, "\\") && !strings.Contains(username, "@") && domainName != "" {
		fmt.Printf("DEBUG: Trying with domain prefix %s\\%s\n", domainName, username)
		cmd := exec.CommandContext(ctx, "smbclient", "//localhost/netlogon", "-U", domainName+"\\"+username+"%"+password, "-c", "exit")
		err := cmd.Run()
		if err == nil {
			fmt.Printf("DEBUG: SAMBA authentication successful for %s\\%s\n", domainName, username)
			return true
		}
		fmt.Printf("DEBUG: smbclient with domain prefix failed: %v\n", err)
	}

	// Try with realm format
	if domainName != "" {
		realm := strings.ToLower(domainName) + ".local"
		fmt.Printf("DEBUG: Trying with realm format %s@%s\n", username, realm)
		cmd := exec.CommandContext(ctx, "smbclient", "//localhost/netlogon", "-U", username+"@"+realm+"%"+password, "-c", "exit")
		err := cmd.Run()
		if err == nil {
			fmt.Printf("DEBUG: SAMBA authentication successful for %s@%s\n", username, realm)
			return true
		}
		fmt.Printf("DEBUG: smbclient with realm format failed: %v\n", err)
	}

	// Try Kerberos authentication as a last resort
	if domainName != "" {
		fmt.Printf("DEBUG: Trying Kerberos authentication for %s\n", username)
		if tryKerberosAuth(username, password, domainName) {
			fmt.Printf("DEBUG: Kerberos authentication successful for %s\n", username)
			return true
		}
	}

	fmt.Printf("DEBUG: All SAMBA authentication attempts failed for %s\n", username)
	return false
}

// getDomainName retrieves the current domain name
func getDomainName() string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "samba-tool", "domain", "info", "127.0.0.1")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("DEBUG: Failed to get domain info: %v\n", err)
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "NetBIOS Domain:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				domain := strings.TrimSpace(parts[1])
				fmt.Printf("DEBUG: Found domain name: %s\n", domain)
				return domain
			}
		}
	}

	fmt.Printf("DEBUG: Could not parse domain name from output: %s\n", string(output))
	return ""
}

// tryKerberosAuth attempts Kerberos authentication
func tryKerberosAuth(username, password, domainName string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	realm := strings.ToUpper(domainName) + ".LOCAL"

	// Try kinit with password
	cmd := exec.CommandContext(ctx, "kinit", username+"@"+realm)
	cmd.Stdin = strings.NewReader(password + "\n")

	err := cmd.Run()
	if err == nil {
		// Clean up the ticket
		exec.Command("kdestroy").Run()
		return true
	}

	fmt.Printf("DEBUG: kinit failed: %v\n", err)
	return false
}
