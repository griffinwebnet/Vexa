package utils

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// ArgValidator is a function type for validating command arguments
type ArgValidator func(string) bool

// CommandPolicy defines security policies for command execution
type CommandPolicy struct {
	Allowed           bool
	StaticArgs        []string
	PositionalArgs    map[int]ArgValidator
	RequiresElevation bool
	MaxArgs           int
}

// CommandSanitizer provides secure command execution
type CommandSanitizer struct {
	policies map[string]CommandPolicy
}

// NewCommandSanitizer creates a new command sanitizer with security policies
func NewCommandSanitizer() *CommandSanitizer {
	return &CommandSanitizer{
		policies: map[string]CommandPolicy{
			// System commands
			"systemctl": {
				Allowed:    true,
				StaticArgs: []string{"is-active", "start", "stop", "enable", "disable", "daemon-reload", "status", "restart"},
				PositionalArgs: map[int]ArgValidator{
					1: isSafeServiceName,
				},
				RequiresElevation: false,
				MaxArgs:           2,
			},
			"hostname": {
				Allowed:    true,
				StaticArgs: []string{"-I"},
				MaxArgs:    1,
			},
			"id": {
				Allowed:    true,
				StaticArgs: []string{"-u", "-g", "-G", "-n", "-nG"},
				PositionalArgs: map[int]ArgValidator{
					1: isSafeUsername,
				},
				MaxArgs: 3,
			},
			"bash": {
				Allowed:    true,
				StaticArgs: []string{"./update.sh"},
				PositionalArgs: map[int]ArgValidator{
					0: isSafePath,
				},
				MaxArgs: 3,
			},
			"pgrep": {
				Allowed:    true,
				StaticArgs: []string{"-f"},
				PositionalArgs: map[int]ArgValidator{
					1: isSafeProcessName,
				},
				MaxArgs: 2,
			},
			"tail": {
				Allowed:    true,
				StaticArgs: []string{"-n"},
				PositionalArgs: map[int]ArgValidator{
					1: isSafePath,
				},
				MaxArgs: 3,
			},
			"ping": {
				Allowed:    true,
				StaticArgs: []string{"-c", "1", "-W", "1"},
				PositionalArgs: map[int]ArgValidator{
					3: isSafeHostname,
				},
				MaxArgs: 4,
			},
			"ip": {
				Allowed:    true,
				StaticArgs: []string{"route", "get"},
				PositionalArgs: map[int]ArgValidator{
					2: isSafeIPAddress,
				},
				MaxArgs: 3,
			},
			"host": {
				Allowed: true,
				PositionalArgs: map[int]ArgValidator{
					0: isSafeHostname,
				},
				MaxArgs: 1,
			},
			"nslookup": {
				Allowed: true,
				PositionalArgs: map[int]ArgValidator{
					0: isSafeHostname,
				},
				MaxArgs: 1,
			},
			"netstat": {
				Allowed:    true,
				StaticArgs: []string{"-tlnp"},
				MaxArgs:    1,
			},
			"curl": {
				Allowed:    true,
				StaticArgs: []string{"-fsSL", "-v", "--connect-timeout", "-s", "-o", "/dev/null", "-w", "%{http_code}"},
				PositionalArgs: map[int]ArgValidator{
					1: isSafeURL,
				},
				MaxArgs: 8,
			},
			"wget": {
				Allowed:    true,
				StaticArgs: []string{"--output-document"},
				PositionalArgs: map[int]ArgValidator{
					0: isSafeURL,
					1: isSafePath,
				},
				MaxArgs: 2,
			},
			"uname": {
				Allowed:    true,
				StaticArgs: []string{"-m"},
				MaxArgs:    1,
			},
			"which": {
				Allowed: true,
				PositionalArgs: map[int]ArgValidator{
					0: isSafeCommandName,
				},
				MaxArgs: 1,
			},
			"test": {
				Allowed:    true,
				StaticArgs: []string{"-f"},
				PositionalArgs: map[int]ArgValidator{
					1: isSafePath,
				},
				MaxArgs: 2,
			},
			"stat": {
				Allowed:    true,
				StaticArgs: []string{"-c", "%y"},
				PositionalArgs: map[int]ArgValidator{
					2: isSafePath,
				},
				MaxArgs: 3,
			},
			"rm": {
				Allowed:    true,
				StaticArgs: []string{"-f"},
				PositionalArgs: map[int]ArgValidator{
					1: isSafePath,
				},
				RequiresElevation: true,
				MaxArgs:           2,
			},
			"pkill": {
				Allowed:    true,
				StaticArgs: []string{"-f"},
				PositionalArgs: map[int]ArgValidator{
					1: isSafeProcessName,
				},
				RequiresElevation: true,
				MaxArgs:           2,
			},
			"journalctl": {
				Allowed:    true,
				StaticArgs: []string{"-u", "--no-pager", "-n"},
				PositionalArgs: map[int]ArgValidator{
					1: isSafeServiceName,
				},
				MaxArgs: 4,
			},

			// Samba commands
			"samba-tool": {
				Allowed: true,
				StaticArgs: []string{
					"domain", "info", "user", "create", "list", "group", "dns", "add", "settings", "set",
					"127.0.0.1", "ou", "delete", "passwordsettings", "show", "-s", "--parameter-name",
					"computer", "provision", "--use-rfc2307", "--version",
				},
				PositionalArgs: map[int]ArgValidator{
					// Apply flexible validator to all argument positions
					0: isSafeFlexibleSambaArg,
					1: isSafeFlexibleSambaArg,
					2: isSafeFlexibleSambaArg,
					3: isSafeFlexibleSambaArg,
					4: isSafeFlexibleSambaArg,
					5: isSafeFlexibleSambaArg,
					6: isSafeFlexibleSambaArg,
					7: isSafeFlexibleSambaArg,
					8: isSafeFlexibleSambaArg,
					9: isSafeFlexibleSambaArg,
				},
				MaxArgs: 15,
			},
			"testparm": {
				Allowed:    true,
				StaticArgs: []string{"-s", "--parameter-name"},
				PositionalArgs: map[int]ArgValidator{
					2: isSafeSambaArg,
				},
				MaxArgs: 3,
			},
			"smbclient": {
				Allowed:    true,
				StaticArgs: []string{"//localhost/ipc$", "//localhost/netlogon", "-L", "localhost", "-U", "-c", "exit"},
				PositionalArgs: map[int]ArgValidator{
					0: func(arg string) bool {
						// Position 0 can be SMB path or -L flag
						return arg == "-L" || isSafeSMBPath(arg)
					},
					3: isSafeCredential,
				},
				MaxArgs: 6,
			},
			"pamtester": {
				Allowed:    true,
				StaticArgs: []string{"login", "authenticate"},
				PositionalArgs: map[int]ArgValidator{
					0: isSafeServiceName,
					1: isSafeUsername,
					2: isSafeCommandName,
				},
				MaxArgs: 4,
			},
			"wbinfo": {
				Allowed:    true,
				StaticArgs: []string{"--authenticate"},
				PositionalArgs: map[int]ArgValidator{
					0: isSafeCredential,
				},
				MaxArgs: 1,
			},
			"ntlm_auth": {
				Allowed:    true,
				StaticArgs: []string{"--username", "--password", "--domain"},
				PositionalArgs: map[int]ArgValidator{
					0: isSafeUsername,
					1: isSafePassword,
					2: isSafeDomainName,
				},
				MaxArgs: 3,
			},
			"kinit": {
				Allowed: true,
				PositionalArgs: map[int]ArgValidator{
					0: isSafeUsername,
				},
				MaxArgs: 1,
			},
			"kdestroy": {
				Allowed: true,
				MaxArgs: 0,
			},

			// Headscale/Tailscale commands
			"headscale": {
				Allowed:    true,
				StaticArgs: []string{"--help", "users", "list", "create", "preauthkeys", "nodes", "migrate", "--output", "json", "version", "-c", "/etc/headscale/config.yaml", "-o", "-u", "--reusable", "--expiration", "131400h"},
				PositionalArgs: map[int]ArgValidator{
					0: isSafeHeadscaleArg,
					1: isSafeHeadscaleArg,
					2: isSafeHeadscaleArg,
					3: isSafeHeadscaleArg,
					4: isSafeHeadscaleArg,
					5: isSafeHeadscaleArg,
					6: isSafeHeadscaleArg,
					7: isSafeHeadscaleArg,
					8: isSafeHeadscaleArg,
					9: isSafeHeadscaleArg,
				},
				MaxArgs: 15,
			},
			"tailscale": {
				Allowed:    true,
				StaticArgs: []string{"up", "status", "debug", "prefs", "login", "--json", "--output"},
				PositionalArgs: map[int]ArgValidator{
					0: isSafeTailscaleArg,
				},
				MaxArgs: 10,
			},

			// Package management
			"apt": {
				Allowed:    true,
				StaticArgs: []string{"update", "install", "-y", "autoremove", "-f", "list", "--upgradable"},
				PositionalArgs: map[int]ArgValidator{
					2: isSafePackageName,
				},
				RequiresElevation: true,
				MaxArgs:           4,
			},
			"dpkg": {
				Allowed:    true,
				StaticArgs: []string{"-i", "--remove", "-s"},
				PositionalArgs: map[int]ArgValidator{
					1: isSafePackageName,
				},
				RequiresElevation: true,
				MaxArgs:           2,
			},
			"dnf": {
				Allowed:    true,
				StaticArgs: []string{"install", "-y"},
				PositionalArgs: map[int]ArgValidator{
					1: isSafePackageName,
				},
				RequiresElevation: true,
				MaxArgs:           2,
			},
			"zypper": {
				Allowed:    true,
				StaticArgs: []string{"install", "-y"},
				PositionalArgs: map[int]ArgValidator{
					1: isSafePackageName,
				},
				RequiresElevation: true,
				MaxArgs:           2,
			},

			// Vexa CLI
			"vexa": {
				Allowed:    true,
				StaticArgs: []string{"update", "start", "status", "--json", "log", "--build-source"},
				MaxArgs:    5,
			},
		},
	}
}

// SanitizeCommand validates and sanitizes command execution
func (cs *CommandSanitizer) SanitizeCommand(name string, args ...string) error {
	// Check if command is allowed
	policy, exists := cs.policies[name]
	if !exists || !policy.Allowed {
		return fmt.Errorf("command '%s' is not allowed", name)
	}

	// Check argument count
	if len(args) > policy.MaxArgs {
		return fmt.Errorf("too many arguments for command '%s' (max %d)", name, policy.MaxArgs)
	}

	// Validate static arguments
	for i, arg := range args {
		// Check if this position has a specific validator
		if validator, hasValidator := policy.PositionalArgs[i]; hasValidator {
			if !validator(arg) {
				return fmt.Errorf("argument at position %d ('%s') failed validation for command '%s'", i, arg, name)
			}
			continue
		}

		// Check static arguments
		if len(policy.StaticArgs) > 0 {
			found := false
			for _, staticArg := range policy.StaticArgs {
				if arg == staticArg {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("argument '%s' is not allowed for command '%s'", arg, name)
			}
		}
	}

	// Log the command execution for audit
	Info("[SafeExec] %s %v", name, args)

	return nil
}

// Validator functions for different argument types

// isSafeHostname validates hostnames
func isSafeHostname(hostname string) bool {
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}

	// Check for dangerous characters
	dangerous := []string{
		";", "&", "|", "<", ">", "`", "$(", ")", "{", "}", "[", "]",
		"*", "?", "~", "!", "#", "@", "^", "\\", "/", "\"", "'",
		" ", "\t", "\n", "\r",
	}

	for _, danger := range dangerous {
		if strings.Contains(hostname, danger) {
			return false
		}
	}

	// Basic hostname validation
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	return hostnameRegex.MatchString(hostname)
}

// isSafeIPAddress validates IP addresses
func isSafeIPAddress(ip string) bool {
	// Allow common test IPs and localhost
	allowedIPs := []string{
		"1.1.1.1",
		"8.8.8.8",
		"127.0.0.1",
		"localhost",
	}

	for _, allowed := range allowedIPs {
		if ip == allowed {
			return true
		}
	}

	// Basic IP validation
	ipRegex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	return ipRegex.MatchString(ip)
}

// isSafeServiceName validates systemd service names
func isSafeServiceName(service string) bool {
	if len(service) == 0 || len(service) > 100 {
		return false
	}

	// Check for dangerous characters
	dangerous := []string{
		";", "&", "|", "<", ">", "`", "$(", ")", "{", "}", "[", "]",
		"*", "?", "~", "!", "#", "@", "^", "\\", "/", "\"", "'",
		" ", "\t", "\n", "\r",
	}

	for _, danger := range dangerous {
		if strings.Contains(service, danger) {
			return false
		}
	}

	// Allow common services
	allowedServices := []string{
		"headscale", "tailscaled", "samba-ad-dc", "systemd-resolved",
		"vexa-api", "nginx", "apache2", "mysql", "postgresql", "login",
	}

	for _, allowed := range allowedServices {
		if service == allowed {
			return true
		}
	}

	return false
}

// isSafeCommandName validates command names
func isSafeCommandName(cmd string) bool {
	if len(cmd) == 0 || len(cmd) > 50 {
		return false
	}

	// Check for dangerous characters
	dangerous := []string{
		";", "&", "|", "<", ">", "`", "$(", ")", "{", "}", "[", "]",
		"*", "?", "~", "!", "#", "@", "^", "\\", "/", "\"", "'",
		" ", "\t", "\n", "\r",
	}

	for _, danger := range dangerous {
		if strings.Contains(cmd, danger) {
			return false
		}
	}

	return true
}

// isSafeProcessName validates process names for pkill
func isSafeProcessName(process string) bool {
	if len(process) == 0 || len(process) > 100 {
		return false
	}

	// Check for dangerous characters
	dangerous := []string{
		";", "&", "|", "<", ">", "`", "$(", ")", "{", "}", "[", "]",
		"*", "?", "~", "!", "#", "@", "^", "\\", "/", "\"", "'",
		" ", "\t", "\n", "\r",
	}

	for _, danger := range dangerous {
		if strings.Contains(process, danger) {
			return false
		}
	}

	return true
}

// isSafeUsername validates usernames
func isSafeUsername(username string) bool {
	if len(username) == 0 || len(username) > 100 {
		return false
	}

	// Check for dangerous characters
	dangerous := []string{
		";", "&", "|", "<", ">", "`", "$(", ")", "{", "}", "[", "]",
		"*", "?", "~", "!", "#", "@", "^", "\\", "/", "\"", "'",
		" ", "\t", "\n", "\r",
	}

	for _, danger := range dangerous {
		if strings.Contains(username, danger) {
			return false
		}
	}

	return true
}

// isSafePassword validates passwords
func isSafePassword(password string) bool {
	if len(password) == 0 || len(password) > 200 {
		return false
	}

	// Check for dangerous characters
	dangerous := []string{
		";", "&", "|", "<", ">", "`", "$(", ")", "{", "}", "[", "]",
		"*", "?", "~", "!", "#", "^", "\\", "/", "\"", "'",
		" ", "\t", "\n", "\r",
	}

	for _, danger := range dangerous {
		if strings.Contains(password, danger) {
			return false
		}
	}

	return true
}

// isSafeDomainName validates domain names
func isSafeDomainName(domain string) bool {
	if len(domain) == 0 || len(domain) > 100 {
		return false
	}

	// Check for dangerous characters
	dangerous := []string{
		";", "&", "|", "<", ">", "`", "$(", ")", "{", "}", "[", "]",
		"*", "?", "~", "!", "#", "@", "^", "\\", "/", "\"", "'",
		" ", "\t", "\n", "\r",
	}

	for _, danger := range dangerous {
		if strings.Contains(domain, danger) {
			return false
		}
	}

	return true
}

// isSafeSambaArg validates Samba-specific arguments
func isSafeSambaArg(arg string) bool {
	allowedArgs := []string{
		"server role", "workgroup", "realm", "netbios name",
		"domain", "info", "user", "create", "list", "group",
		"dns", "add", "settings", "set", "forwarder", "provision",
	}

	for _, allowed := range allowedArgs {
		if arg == allowed {
			return true
		}
	}

	return false
}

// isSafeFlexibleSambaArg validates Samba arguments including provision flags
func isSafeFlexibleSambaArg(arg string) bool {
	// Allow version flag
	if arg == "--version" {
		return true
	}

	// Allow domain provision arguments with = signs
	if strings.HasPrefix(arg, "--realm=") ||
		strings.HasPrefix(arg, "--domain=") ||
		strings.HasPrefix(arg, "--adminpass=") ||
		strings.HasPrefix(arg, "--server-role=") ||
		strings.HasPrefix(arg, "--dns-backend=") ||
		strings.HasPrefix(arg, "--option=") ||
		strings.HasPrefix(arg, "--given-name=") ||
		strings.HasPrefix(arg, "--mail-address=") ||
		strings.HasPrefix(arg, "--userou=") ||
		strings.HasPrefix(arg, "--description=") ||
		strings.HasPrefix(arg, "--editor=") ||
		strings.HasPrefix(arg, "--full-name=") ||
		strings.HasPrefix(arg, "--newpassword=") ||
		strings.HasPrefix(arg, "--complexity=") ||
		strings.HasPrefix(arg, "--min-pwd-length=") ||
		strings.HasPrefix(arg, "--max-pwd-age=") ||
		strings.HasPrefix(arg, "--history-length=") ||
		strings.HasPrefix(arg, "--min-pwd-age=") {
		return true
	}
	// Fall back to basic samba arg validation
	return isSafeSambaArg(arg)
}

// isSafeSMBPath validates SMB paths
func isSafeSMBPath(path string) bool {
	allowedPaths := []string{
		"//localhost/ipc$",
		"//localhost/netlogon",
	}

	for _, allowed := range allowedPaths {
		if path == allowed {
			return true
		}
	}

	return false
}

// isSafeHeadscaleArg validates Headscale arguments
func isSafeHeadscaleArg(arg string) bool {
	allowedArgs := []string{
		"--help", "users", "list", "create", "preauthkeys", "nodes", "migrate",
		"infrastructure", "-c", "/etc/headscale/config.yaml", "-o", "json",
		"--reusable", "--expiration", "131400h", "-u", "--output", "version",
	}

	for _, allowed := range allowedArgs {
		if arg == allowed {
			return true
		}
	}

	// Allow numeric user IDs
	if matched, _ := regexp.MatchString(`^\d+$`, arg); matched {
		return true
	}

	return false
}

// isSafeTailscaleArg validates Tailscale arguments
func isSafeTailscaleArg(arg string) bool {
	allowedArgs := []string{
		"up", "status", "debug", "prefs", "login",
		"--authkey", "--login-server", "--accept-routes", "--accept-dns=false",
		"--hostname", "--unattended", "--json", "--output",
	}

	for _, allowed := range allowedArgs {
		if arg == allowed {
			return true
		}
	}

	// Allow URLs and hostnames
	if isSafeURL(arg) || isSafeHostname(arg) {
		return true
	}

	return false
}

// isSafePackageName validates package names
func isSafePackageName(pkg string) bool {
	if len(pkg) == 0 || len(pkg) > 100 {
		return false
	}

	// Check for dangerous characters
	dangerous := []string{
		";", "&", "|", "<", ">", "`", "$(", ")", "{", "}", "[", "]",
		"*", "?", "~", "!", "#", "@", "^", "\\", "/", "\"", "'",
		" ", "\t", "\n", "\r",
	}

	for _, danger := range dangerous {
		if strings.Contains(pkg, danger) {
			return false
		}
	}

	// Allow common packages
	allowedPackages := []string{
		"headscale", "tailscale", "samba", "samba-ad-dc", "nginx",
		"apache2", "mysql", "postgresql", "curl", "wget",
	}

	for _, allowed := range allowedPackages {
		if pkg == allowed {
			return true
		}
	}

	return false
}

// isSafePath checks if a path is safe (improved version)
func isSafePath(path string) bool {
	// Clean the path to prevent directory traversal
	clean := filepath.Clean(path)

	// Allow common safe paths with exact boundary checks
	safePaths := []string{
		"/etc/headscale/config.yaml",
		"/var/lib/headscale/",
		"/var/log/vexa/",
		"/usr/local/bin/",
		"/usr/sbin/",
		"/etc/systemd/",
		"/tmp/",
		"/var/tmp/",
		"headscale.deb",
		"tailscale-setup-latest-amd64.msi",
	}

	for _, safe := range safePaths {
		if clean == safe || strings.HasPrefix(clean, safe+string(os.PathSeparator)) {
			return true
		}
	}

	// Check for dangerous patterns
	dangerous := []string{
		"..",
		"~",
		"$(",
		"`",
		";",
		"&",
		"|",
		"<",
		">",
		"*",
		"?",
		"[",
		"]",
		"{",
		"}",
	}

	for _, danger := range dangerous {
		if strings.Contains(clean, danger) {
			return false
		}
	}

	return false // Default to false for unknown paths
}

// isSafeURL checks if a URL is safe (improved version)
func isSafeURL(urlStr string) bool {
	// Parse the URL properly
	u, err := url.Parse(urlStr)
	if err != nil || u.Scheme != "https" {
		return false
	}

	// Get the hostname and normalize it
	host := strings.ToLower(u.Hostname())

	// Trusted domains with exact matching
	trustedDomains := []string{
		"github.com",
		"pkgs.tailscale.com",
		"raw.githubusercontent.com",
	}

	for _, domain := range trustedDomains {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}

	return false
}

// isSafeCredential checks if a credential string is safe
func isSafeCredential(cred string) bool {
	// Check for username%password format
	if strings.Contains(cred, "%") {
		parts := strings.Split(cred, "%")
		if len(parts) == 2 {
			username, password := parts[0], parts[1]
			// Basic validation - only block actual shell injection characters
			// Allow most special characters in passwords including ~, !, @, #, $, ^, *, etc.
			dangerous := []string{
				";", "&", "|", "<", ">", "`", "$(", ")", "{", "}", "[", "]", "\"", "'", "\\",
			}

			for _, danger := range dangerous {
				if strings.Contains(username, danger) || strings.Contains(password, danger) {
					return false
				}
			}

			// Allow reasonable length
			if len(username) > 0 && len(username) <= 100 && len(password) > 0 && len(password) <= 200 {
				return true
			}
		}
	}

	// Check for domain\username%password format
	if strings.Contains(cred, "\\") && strings.Contains(cred, "%") {
		// Similar validation for domain credentials - allow most special chars except dangerous shell chars
		dangerous := []string{
			";", "&", "|", "<", ">", "`", "$(", ")", "{", "}", "[", "]", "\"", "'",
		}

		for _, danger := range dangerous {
			if strings.Contains(cred, danger) {
				return false
			}
		}

		if len(cred) > 0 && len(cred) <= 300 {
			return true
		}
	}

	return false
}

// SafeExec executes a command with sanitization
func (cs *CommandSanitizer) SafeExec(name string, args ...string) (*exec.Cmd, error) {
	if err := cs.SanitizeCommand(name, args...); err != nil {
		return nil, fmt.Errorf("command sanitization failed: %v", err)
	}

	return exec.Command(name, args...), nil
}

// SafeExecContext executes a command with context and sanitization
func (cs *CommandSanitizer) SafeExecContext(ctx context.Context, name string, args ...string) (*exec.Cmd, error) {
	if err := cs.SanitizeCommand(name, args...); err != nil {
		return nil, fmt.Errorf("command sanitization failed: %v", err)
	}

	return exec.CommandContext(ctx, name, args...), nil
}

// Global sanitizer instance
var globalSanitizer = NewCommandSanitizer()

// SafeCommand is a convenience function for safe command execution
func SafeCommand(name string, args ...string) (*exec.Cmd, error) {
	return globalSanitizer.SafeExec(name, args...)
}

// SafeCommandContext is a convenience function for safe command execution with context
func SafeCommandContext(ctx context.Context, name string, args ...string) (*exec.Cmd, error) {
	return globalSanitizer.SafeExecContext(ctx, name, args...)
}
