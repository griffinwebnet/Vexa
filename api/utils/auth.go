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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Method 1: Try simple smbclient first (most reliable for basic auth)
	fmt.Printf("DEBUG: Trying smbclient //localhost/ipc$ with user %s\n", username)
	cmd := exec.CommandContext(ctx, "smbclient", "//localhost/ipc$", "-U", username+"%"+password, "-c", "exit")
	output, err := cmd.CombinedOutput()
	if err == nil {
		fmt.Printf("DEBUG: smbclient authentication successful for %s\n", username)
		return true
	} else {
		fmt.Printf("DEBUG: smbclient failed: %v, output: %s\n", err, string(output))
	}

	// Method 2: Try with netlogon share
	fmt.Printf("DEBUG: Trying smbclient //localhost/netlogon with user %s\n", username)
	cmd2 := exec.CommandContext(ctx, "smbclient", "//localhost/netlogon", "-U", username+"%"+password, "-c", "exit")
	output2, err := cmd2.CombinedOutput()
	if err == nil {
		fmt.Printf("DEBUG: smbclient netlogon authentication successful for %s\n", username)
		return true
	} else {
		fmt.Printf("DEBUG: smbclient netlogon failed: %v, output: %s\n", err, string(output2))
	}

	// Method 3: Get domain name and try domain-prefixed authentication
	domainName := getDomainName()
	fmt.Printf("DEBUG: Detected domain name: %s\n", domainName)

	if domainName != "" {
		// Try with domain prefix
		fmt.Printf("DEBUG: Trying smbclient with domain prefix %s\\%s\n", domainName, username)
		cmd3 := exec.CommandContext(ctx, "smbclient", "//localhost/ipc$", "-U", domainName+"\\"+username+"%"+password, "-c", "exit")
		output3, err := cmd3.CombinedOutput()
		if err == nil {
			fmt.Printf("DEBUG: smbclient domain auth successful for %s\\%s\n", domainName, username)
			return true
		} else {
			fmt.Printf("DEBUG: smbclient domain auth failed: %v, output: %s\n", err, string(output3))
		}

		// Try with realm format
		realm := strings.ToLower(domainName) + ".local"
		fmt.Printf("DEBUG: Trying smbclient with realm format %s@%s\n", username, realm)
		cmd4 := exec.CommandContext(ctx, "smbclient", "//localhost/ipc$", "-U", username+"@"+realm+"%"+password, "-c", "exit")
		if err := cmd4.Run(); err == nil {
			fmt.Printf("DEBUG: smbclient realm auth successful for %s@%s\n", username, realm)
			return true
		} else {
			fmt.Printf("DEBUG: smbclient realm auth failed: %v\n", err)
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

// CheckDomainAdminStatus checks if a user is in the Domain Admins group
func CheckDomainAdminStatus(username string) bool {
	fmt.Printf("DEBUG: Checking admin status for user: %s\n", username)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get domain name
	domainName := getDomainName()
	if domainName == "" {
		fmt.Printf("DEBUG: No domain name detected, cannot check admin status\n")
		return false
	}

	// Try multiple formats for the username
	userFormats := []string{
		username,
		domainName + "\\" + username,
		username + "@" + strings.ToLower(domainName) + ".local",
	}

	for _, userFormat := range userFormats {
		fmt.Printf("DEBUG: Checking admin status for user format: %s\n", userFormat)

		// Use samba-tool to check group membership
		cmd := exec.CommandContext(ctx, "samba-tool", "group", "listmembers", "Domain Admins")
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("DEBUG: Failed to list Domain Admins: %v\n", err)
			continue
		}

		members := strings.Split(string(output), "\n")
		for _, member := range members {
			member = strings.TrimSpace(member)
			if member == userFormat || member == username {
				fmt.Printf("DEBUG: User %s found in Domain Admins group\n", username)
				return true
			}
		}
	}

	fmt.Printf("DEBUG: User %s is NOT in Domain Admins group\n", username)
	return false
}

// VerifyCurrentPassword verifies a user's current password directly against Samba
func VerifyCurrentPassword(username, password string) bool {
	fmt.Printf("DEBUG: Verifying current password for user: %s\n", username)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get domain name
	domainName := getDomainName()
	if domainName == "" {
		fmt.Printf("DEBUG: No domain name detected for password verification\n")
		return false
	}

	// Try with domain prefix first
	fmt.Printf("DEBUG: Trying password verification with domain prefix %s\\%s\n", domainName, username)
	cmd := exec.CommandContext(ctx, "smbclient", "//localhost/ipc$", "-U", domainName+"\\"+username+"%"+password, "-c", "exit")
	output, err := cmd.CombinedOutput()
	if err == nil {
		fmt.Printf("DEBUG: Password verification successful for %s\\%s\n", domainName, username)
		return true
	} else {
		fmt.Printf("DEBUG: Password verification failed for %s\\%s: %v, output: %s\n", domainName, username, err, string(output))
	}

	// Try without domain prefix
	fmt.Printf("DEBUG: Trying password verification without domain prefix for %s\n", username)
	cmd2 := exec.CommandContext(ctx, "smbclient", "//localhost/ipc$", "-U", username+"%"+password, "-c", "exit")
	output2, err2 := cmd2.CombinedOutput()
	if err2 == nil {
		fmt.Printf("DEBUG: Password verification successful for %s\n", username)
		return true
	} else {
		fmt.Printf("DEBUG: Password verification failed for %s: %v, output: %s\n", username, err2, string(output2))
	}

	fmt.Printf("DEBUG: Password verification failed for user %s\n", username)
	return false
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

// tryWbinfoAuth attempts authentication using wbinfo
func tryWbinfoAuth(username, password, domainName string) bool {
	if domainName == "" {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try with domain prefix
	fullUser := domainName + "\\" + username
	fmt.Printf("DEBUG: Trying wbinfo authentication for %s\n", fullUser)

	cmd := exec.CommandContext(ctx, "wbinfo", "--authenticate", fullUser+"%"+password)
	err := cmd.Run()
	if err == nil {
		return true
	}

	fmt.Printf("DEBUG: wbinfo authentication failed: %v\n", err)
	return false
}

// tryNtlmAuth attempts authentication using ntlm_auth
func tryNtlmAuth(username, password, domainName string) bool {
	if domainName == "" {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try ntlm_auth
	fmt.Printf("DEBUG: Trying ntlm_auth for %s\n", username)

	cmd := exec.CommandContext(ctx, "ntlm_auth", "--username="+username, "--password="+password, "--domain="+domainName)
	output, err := cmd.Output()

	if err == nil && strings.Contains(string(output), "NT_STATUS_OK") {
		fmt.Printf("DEBUG: ntlm_auth successful\n")
		return true
	}

	fmt.Printf("DEBUG: ntlm_auth failed: %v, output: %s\n", err, string(output))
	return false
}

// trySmbclientAuth attempts authentication using smbclient
func trySmbclientAuth(username, password, domainName string, ctx context.Context) bool {
	// Try multiple authentication methods for domain users
	targets := []string{
		"//localhost/netlogon",
		"//localhost/ipc$",
	}

	// Try without domain prefix first
	for _, target := range targets {
		fmt.Printf("DEBUG: Trying smbclient %s with user %s\n", target, username)
		cmd := exec.CommandContext(ctx, "smbclient", target, "-U", username+"%"+password, "-c", "exit")
		err := cmd.Run()
		if err == nil {
			fmt.Printf("DEBUG: smbclient authentication successful for %s\n", username)
			return true
		}
		fmt.Printf("DEBUG: smbclient failed for %s: %v\n", target, err)
	}

	// Try with domain prefix if username doesn't already have it
	if !strings.Contains(username, "\\") && !strings.Contains(username, "@") && domainName != "" {
		fmt.Printf("DEBUG: Trying smbclient with domain prefix %s\\%s\n", domainName, username)
		cmd := exec.CommandContext(ctx, "smbclient", "//localhost/netlogon", "-U", domainName+"\\"+username+"%"+password, "-c", "exit")
		err := cmd.Run()
		if err == nil {
			fmt.Printf("DEBUG: smbclient authentication successful for %s\\%s\n", domainName, username)
			return true
		}
		fmt.Printf("DEBUG: smbclient with domain prefix failed: %v\n", err)
	}

	// Try with realm format
	if domainName != "" {
		realm := strings.ToLower(domainName) + ".local"
		fmt.Printf("DEBUG: Trying smbclient with realm format %s@%s\n", username, realm)
		cmd := exec.CommandContext(ctx, "smbclient", "//localhost/netlogon", "-U", username+"@"+realm+"%"+password, "-c", "exit")
		err := cmd.Run()
		if err == nil {
			fmt.Printf("DEBUG: smbclient authentication successful for %s@%s\n", username, realm)
			return true
		}
		fmt.Printf("DEBUG: smbclient with realm format failed: %v\n", err)
	}

	return false
}
