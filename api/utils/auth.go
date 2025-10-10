package utils

import (
	"context"
	"strings"
	"time"
)

// AuthenticatePAM authenticates against system PAM authentication
func AuthenticatePAM(username, password string) bool {
	// Use pamtester to test PAM authentication
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd, err := SafeCommandContext(ctx, "pamtester", "login", username, "authenticate")
	if err != nil {
		Error("Command sanitization failed for pamtester: %v", err)
		return false
	}
	cmd.Stdin = strings.NewReader(password + "\n")

	runErr := cmd.Run()
	return runErr == nil
}

// AuthenticateSAMBA authenticates against SAMBA/Active Directory
func AuthenticateSAMBA(username, password string) bool {
	Debug("Attempting SAMBA authentication for user: %s", username)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Method 1: Try simple smbclient first (most reliable for basic auth)
	// method replaces earlier fallback smb method deleted as of 0.3.122
	cmd, err := SafeCommandContext(ctx, "smbclient", "//localhost/ipc$", "-U", username+"%"+password, "-c", "exit")
	if err != nil {
		Error("Command sanitization failed for smbclient: %v", err)
		return false
	}
	output, err := cmd.CombinedOutput()
	if err == nil {
		Debug("smbclient authentication successful for %s", username)
		return true
	} else {
		Debug("smbclient failed: %v, output: %s", err, string(output))
	}

	// Method 2: Try with netlogon share
	// may be deleted later. I've yet to see it dump its debug string to the
	// log so the primary funciton seems to be sufficient.
	Debug("Trying smbclient //localhost/netlogon with user %s", username)
	cmd2, err := SafeCommandContext(ctx, "smbclient", "//localhost/netlogon", "-U", username+"%"+password, "-c", "exit")
	if err != nil {
		Error("Command sanitization failed for smbclient netlogon: %v", err)
		return false
	}
	output2, err2 := cmd2.CombinedOutput()
	if err2 == nil {
		Debug("smbclient netlogon authentication successful for %s", username)
		return true
	} else {
		Debug("smbclient netlogon failed: %v, output: %s", err2, string(output2))
	}

	// Method 3: Get domain name and try domain-prefixed authentication
	// 50/50 This gets deleted later. I've yet to see it dump its debug string to the
	// log so the primary funciton seems to be sufficient.
	domainName := getDomainName()
	Debug("Detected domain name: %s", domainName)

	if domainName != "" {
		// Try with domain prefix
		Debug("Trying smbclient with domain prefix %s\\%s", domainName, username)
		cmd3, err := SafeCommandContext(ctx, "smbclient", "//localhost/ipc$", "-U", domainName+"\\"+username+"%"+password, "-c", "exit")
		if err != nil {
			Debug("Command sanitization failed for smbclient domain: %v", err)
		} else {
			output3, err3 := cmd3.CombinedOutput()
			if err3 == nil {
				Debug("smbclient domain auth successful for %s\\%s", domainName, username)
				return true
			} else {
				Debug("smbclient domain auth failed: %v, output: %s", err3, string(output3))
			}
		}

		// Try with realm format
		realm := strings.ToLower(domainName) + ".local"
		Debug("Trying smbclient with realm format %s@%s", username, realm)
		cmd4, err := SafeCommandContext(ctx, "smbclient", "//localhost/ipc$", "-U", username+"@"+realm+"%"+password, "-c", "exit")
		if err != nil {
			Debug("Command sanitization failed for smbclient realm: %v", err)
		} else {
			if err4 := cmd4.Run(); err4 == nil {
				Debug("smbclient realm auth successful for %s@%s", username, realm)
				return true
			} else {
				Debug("smbclient realm auth failed: %v", err4)
			}
		}
	}

	Debug("All SAMBA authentication attempts failed for %s", username)
	return false
}

// getDomainName retrieves the current domain name
func getDomainName() string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd, err := SafeCommandContext(ctx, "samba-tool", "domain", "info", "127.0.0.1")
	if err != nil {
		Debug("Command sanitization failed for samba-tool domain info: %v", err)
		return ""
	}
	output, err2 := cmd.Output()
	if err2 != nil {
		Debug("Failed to get domain info: %v", err2)
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "NetBIOS Domain:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				domain := strings.TrimSpace(parts[1])
				Debug("Found domain name: %s", domain)
				return domain
			}
		}
	}

	Debug("Could not parse domain name from output: %s", string(output))
	return ""
}

// CheckDomainAdminStatus checks if a user is in admin groups (Domain Admins or Administrators)
func CheckDomainAdminStatus(username string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Only use samba-tool... after MUCH testing, it is the most reliable method
	// Check both "Domain Admins" and "Administrators" groups
	adminGroups := []string{"Domain Admins", "Administrators"}

	for _, group := range adminGroups {
		cmd, err := SafeCommandContext(ctx, "samba-tool", "group", "listmembers", group)
		if err != nil {
			Debug("Command sanitization failed for samba-tool group: %v", err)
			return false
		}
		output, err2 := cmd.Output()

		if err2 == nil && len(output) > 0 {
			members := strings.Split(string(output), "\n")
			for _, member := range members {
				member = strings.TrimSpace(member)
				if member == username && member != "" {
					return true
				}
			}
		}
	}

	return false
}

// VerifyCurrentPassword verifies a user's current password directly against the Samba domain
func VerifyCurrentPassword(username, password string) bool {
	Debug("VerifyCurrentPassword called for user: %s", username)
	Debug("Verifying current password for user: %s", username)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get domain name
	domainName := getDomainName()
	if domainName == "" {
		Debug("No domain name detected for password verification\n")
		return false
	}

	// Try with domain prefix first
	Debug("Trying password verification with domain prefix %s\\%s\n", domainName, username)
	cmd, err := SafeCommandContext(ctx, "smbclient", "//localhost/ipc$", "-U", domainName+"\\"+username+"%"+password, "-c", "exit")
	if err != nil {
		Debug("Command sanitization failed for smbclient domain: %v\n", err)
		return false
	}
	output, err2 := cmd.CombinedOutput()
	if err2 == nil {
		Debug("Password verification successful for %s\\%s\n", domainName, username)
		return true
	} else {
		Debug("Password verification failed for %s\\%s: %v, output: %s\n", domainName, username, err2, string(output))
	}

	// Try without domain prefix
	Debug("Trying password verification without domain prefix for %s\n", username)
	cmd2, err := SafeCommandContext(ctx, "smbclient", "//localhost/ipc$", "-U", username+"%"+password, "-c", "exit")
	if err != nil {
		Debug("Command sanitization failed for smbclient: %v", err)
		return false
	}
	output2, err3 := cmd2.CombinedOutput()
	if err3 == nil {
		Debug("Password verification successful for %s", username)
		return true
	} else {
		Debug("Password verification failed for %s: %v, output: %s", username, err3, string(output2))
	}

	Debug("Password verification failed for user %s\n", username)
	return false
}

// CheckLocalAdminStatus determines if a local (PAM) user is an administrator
// by checking common sudo-capable groups and root. This is used to allow
// bootstrap by local admins on unprovisioned systems or fallback access should
// the domain authentication fail or domain servives are damaged.
func CheckLocalAdminStatus(username string) bool {
	if username == "root" {
		// ta-da! this was the hardest part of the whole project :P
		return true
	}

	// Get user's groups: id -nG <user>
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd, err := SafeCommandContext(ctx, "id", "-nG", username)
	if err != nil {
		Debug("Command sanitization failed for id: %v", err)
		return false
	}
	output, err2 := cmd.CombinedOutput()
	if err2 == nil {
		groupsLine := strings.TrimSpace(string(output))
		if groupsLine != "" {
			groups := strings.Fields(groupsLine)
			for _, g := range groups {
				switch g {
				case "sudo", "wheel", "admin":
					return true
				}
			}
		}
	}

	// Fallback: check getent for standard admin groups membership
	// this may get deleted later. its not entirely necessary and I
	// am not sure it actually ever runs. I've yet to see it dump its
	// debug string to the log so the primary funciton seems to be sufficient.
	adminGroups := []string{"sudo", "wheel", "admin"}
	for _, grp := range adminGroups {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
		cmd2, err := SafeCommandContext(ctx2, "getent", "group", grp)
		if err != nil {
			Debug("Command sanitization failed for getent: %v", err)
			cancel2()
			continue
		}
		out2, err2 := cmd2.CombinedOutput()
		cancel2()
		if err2 == nil {
			// format: group_name:*:GID:user1,user2
			parts := strings.Split(strings.TrimSpace(string(out2)), ":")
			if len(parts) >= 4 {
				members := strings.Split(parts[3], ",")
				for _, m := range members {
					if strings.TrimSpace(m) == username {
						return true
					}
				}
			}
		}
	}

	return false
}
