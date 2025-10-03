package exec

import (
	"os/exec"
	"strings"
)

// SambaTool provides an interface for executing samba-tool commands
type SambaTool struct{}

// NewSambaTool creates a new SambaTool instance
func NewSambaTool() *SambaTool {
	return &SambaTool{}
}

// Run executes a samba-tool command with the given arguments
func (s *SambaTool) Run(args ...string) (string, error) {
	cmd := exec.Command("samba-tool", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// UserCreate creates a new user
func (s *SambaTool) UserCreate(username, password string, options UserCreateOptions) (string, error) {
	args := []string{"user", "create", username, password}

	if options.FullName != "" {
		args = append(args, "--given-name="+options.FullName)
	}
	if options.Email != "" {
		args = append(args, "--mail-address="+options.Email)
	}
	if options.OUPath != "" {
		args = append(args, "--userou="+options.OUPath)
	}
	if options.Description != "" {
		args = append(args, "--description="+options.Description)
	}

	return s.Run(args...)
}

// UserList lists all users
func (s *SambaTool) UserList() (string, error) {
	return s.Run("user", "list")
}

// UserShow shows details for a specific user
func (s *SambaTool) UserShow(username string) (string, error) {
	return s.Run("user", "show", username)
}

// UserDelete deletes a user
func (s *SambaTool) UserDelete(username string) (string, error) {
	return s.Run("user", "delete", username)
}

// UserDisable disables a user account
func (s *SambaTool) UserDisable(username string) (string, error) {
	return s.Run("user", "disable", username)
}

// UserEnable enables a user account
func (s *SambaTool) UserEnable(username string) (string, error) {
	return s.Run("user", "enable", username)
}

// DNSForwarders sets DNS forwarders for the domain
func (s *SambaTool) DNSForwarders(forwarders []string) (string, error) {
	args := []string{"dns", "server", "set", "forwarder"}
	args = append(args, forwarders...)
	return s.Run(args...)
}

// GroupCreate creates a new group
func (s *SambaTool) GroupCreate(name string, options GroupCreateOptions) (string, error) {
	args := []string{"group", "add", name}

	if options.Description != "" {
		args = append(args, "--description="+options.Description)
	}

	return s.Run(args...)
}

// GroupList lists all groups
func (s *SambaTool) GroupList() (string, error) {
	return s.Run("group", "list")
}

// GroupListMembers lists members of a group
func (s *SambaTool) GroupListMembers(groupName string) (string, error) {
	return s.Run("group", "listmembers", groupName)
}

// GroupDelete deletes a group
func (s *SambaTool) GroupDelete(groupName string) (string, error) {
	return s.Run("group", "delete", groupName)
}

// GroupAddMembers adds members to a group
func (s *SambaTool) GroupAddMembers(groupName string, members []string) (string, error) {
	args := []string{"group", "addmembers", groupName}
	args = append(args, members...)
	return s.Run(args...)
}

// GroupRemoveMembers removes members from a group
func (s *SambaTool) GroupRemoveMembers(groupName string, members []string) (string, error) {
	args := []string{"group", "removemembers", groupName}
	args = append(args, members...)
	return s.Run(args...)
}

// GroupModify modifies an existing group
func (s *SambaTool) GroupModify(groupName string, description string) (string, error) {
	args := []string{"group", "modify", groupName}
	if description != "" {
		args = append(args, "--description="+description)
	}
	return s.Run(args...)
}

// DomainProvision provisions a new domain
func (s *SambaTool) DomainProvision(options DomainProvisionOptions) (string, error) {
	args := []string{
		"domain", "provision",
		"--realm=" + options.Realm,
		"--domain=" + options.Domain,
		"--adminpass=" + options.AdminPassword,
		"--server-role=dc",
		"--dns-backend=" + options.DNSBackend,
		"--use-rfc2307",
	}

	if options.DNSForwarder != "" {
		args = append(args, "--option=dns forwarder = "+options.DNSForwarder)
	}

	// Workaround for Samba 4.19.5 LXC bug: use minimal VFS chain with dfs_samba4
	// The security context bug occurs in ACL manipulation regardless of backend
	// See: https://bugzilla.samba.org/show_bug.cgi?id=15203
	args = append(args, "--option=vfs objects = dfs_samba4")

	// Run provision command
	cmd := exec.Command("samba-tool", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// DomainInfo gets domain information
func (s *SambaTool) DomainInfo(server string) (string, error) {
	return s.Run("domain", "info", server)
}

// ParseUserList parses the output of user list command
func (s *SambaTool) ParseUserList(output string) []string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var users []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			users = append(users, strings.TrimSpace(line))
		}
	}
	return users
}

// ParseGroupList parses the output of group list command
func (s *SambaTool) ParseGroupList(output string) []string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var groups []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			groups = append(groups, strings.TrimSpace(line))
		}
	}
	return groups
}

// ParseGroupMembers parses the output of group listmembers command
func (s *SambaTool) ParseGroupMembers(output string) []string {
	return s.ParseUserList(output) // Same parsing logic
}

// UserCreateOptions represents options for user creation
type UserCreateOptions struct {
	FullName    string
	Email       string
	OUPath      string
	Description string
}

// GroupCreateOptions represents options for group creation
type GroupCreateOptions struct {
	Description string
}

// DomainProvisionOptions represents options for domain provisioning
type DomainProvisionOptions struct {
	Domain        string
	Realm         string
	AdminPassword string
	DNSBackend    string
	DNSForwarder  string
}
