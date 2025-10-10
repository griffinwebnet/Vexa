package services

import (
	"fmt"
	"strings"

	"github.com/griffinwebnet/vexa/api/exec"
	"github.com/griffinwebnet/vexa/api/models"
	"github.com/griffinwebnet/vexa/api/utils"
)

// UserService handles user-related business logic
type UserService struct {
	sambaTool *exec.SambaTool
}

// NewUserService creates a new UserService instance
func NewUserService() *UserService {
	return &UserService{
		sambaTool: exec.NewSambaTool(),
	}
}

// ListUsers returns all users in the domain (excluding system accounts)
func (s *UserService) ListUsers() ([]models.User, error) {
	// System accounts to filter out
	systemAccounts := map[string]bool{
		"krbtgt":        true, // Kerberos service account
		"Administrator": true, // Built-in administrator
		"Guest":         true, // Built-in guest account
	}

	output, err := s.sambaTool.UserList()
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %s", output)
	}

	usernames := s.sambaTool.ParseUserList(output)
	users := make([]models.User, 0, len(usernames))

	for _, username := range usernames {
		// Skip system accounts
		if systemAccounts[username] {
			utils.Debug("Filtering out system account: %s", username)
			continue
		}

		users = append(users, models.User{
			Username: username,
			Enabled:  true, // TODO: Get actual status
		})
	}

	utils.Info("Listed %d users (filtered %d system accounts)", len(users), len(usernames)-len(users))
	return users, nil
}

// CreateUser creates a new user in the domain
func (s *UserService) CreateUser(req models.CreateUserRequest) error {
	options := exec.UserCreateOptions{
		FullName:    req.FullName,
		Email:       req.Email,
		OUPath:      req.OUPath,
		Description: req.Description,
	}

	output, err := s.sambaTool.UserCreate(req.Username, req.Password, options)
	if err != nil {
		return fmt.Errorf("failed to create user: %s", output)
	}

	// Add user to group if specified
	if req.Group != "" && req.Group != "Domain Users" {
		if err := s.addUserToGroup(req.Username, req.Group); err != nil {
			// User created but group add failed - return warning
			return fmt.Errorf("user created but failed to add to group: %w", err)
		}
	}

	// Set "must change password at next login" flag if requested
	if req.MustChangePassword {
		if err := s.SetMustChangePassword(req.Username); err != nil {
			// User created but flag setting failed - log warning but don't fail
			fmt.Printf("WARNING: Failed to set must-change-password flag for user %s: %v\n", req.Username, err)
		}
	}

	return nil
}

// GetUser returns details for a specific user
func (s *UserService) GetUser(username string) (*models.User, error) {
	output, err := s.sambaTool.UserShow(username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %s", output)
	}

	// Parse user details from output
	user := s.parseUserShowOutput(username, output)

	// Get user groups
	groups, err := s.getUserGroups(username)
	if err == nil {
		user.Groups = groups
	}

	return user, nil
}

// parseUserShowOutput parses the output from samba-tool user show
func (s *UserService) parseUserShowOutput(username, output string) *models.User {
	user := &models.User{
		Username: username,
		Enabled:  true, // Default to enabled
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse common fields
		if strings.HasPrefix(line, "dn:") {
			// Extract OU path from DN
			// DN format: CN=username,OU=ou_name,DC=domain,DC=local
			parts := strings.Split(line, ",")
			if len(parts) > 1 {
				// Find OU part
				for _, part := range parts[1:] {
					if strings.HasPrefix(part, "OU=") {
						ouName := strings.TrimPrefix(part, "OU=")
						user.Description = ouName // Temporary - we'll add OU path field later
						break
					}
				}
			}
		}

		if strings.Contains(line, "displayName:") {
			user.FullName = strings.TrimSpace(strings.TrimPrefix(line, "displayName:"))
		}

		if strings.Contains(line, "mail:") {
			user.Email = strings.TrimSpace(strings.TrimPrefix(line, "mail:"))
		}

		if strings.Contains(line, "description:") {
			user.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		}

		if strings.Contains(line, "userAccountControl:") {
			// Check if account is disabled (bit 2 = disabled)
			uac := strings.TrimSpace(strings.TrimPrefix(line, "userAccountControl:"))
			// This is a simplified check - in reality UAC is a bit field
			user.Enabled = !strings.Contains(uac, "2") // Simplified check
		}
	}

	return user
}

// getUserGroups gets the groups a user belongs to
func (s *UserService) getUserGroups(username string) ([]string, error) {
	output, err := s.sambaTool.Run("user", "show", username, "--attributes=memberOf")
	if err != nil {
		return nil, err
	}

	var groups []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "memberOf:") {
			// Extract group name from DN
			// Format: CN=GroupName,OU=Groups,DC=domain,DC=local
			dn := strings.TrimPrefix(line, "memberOf:")
			parts := strings.Split(dn, ",")
			if len(parts) > 0 && strings.HasPrefix(parts[0], "CN=") {
				groupName := strings.TrimPrefix(parts[0], "CN=")
				groups = append(groups, groupName)
			}
		}
	}

	return groups, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(username string, req models.UpdateUserRequest) error {
	// Get user's DN first
	userDN, err := s.getUserDN(username)
	if err != nil {
		utils.Error("Failed to get user DN for %s: %v", username, err)
		return fmt.Errorf("failed to get user DN: %v", err)
	}

	// Update full name if provided
	if req.FullName != nil && *req.FullName != "" {
		utils.Info("Updating full name for user %s to: %s", username, *req.FullName)
		if err := s.modifyLDAPAttribute(userDN, "givenName", *req.FullName); err != nil {
			utils.Error("Failed to update full name: %v", err)
			return fmt.Errorf("failed to update full name: %v", err)
		}
		// Also update displayName
		if err := s.modifyLDAPAttribute(userDN, "displayName", *req.FullName); err != nil {
			utils.Warn("Failed to update display name: %v", err)
		}
	}

	// Update email if provided
	if req.Email != nil && *req.Email != "" {
		utils.Info("Updating email for user %s to: %s", username, *req.Email)
		if err := s.modifyLDAPAttribute(userDN, "mail", *req.Email); err != nil {
			utils.Error("Failed to update email: %v", err)
			return fmt.Errorf("failed to update email: %v", err)
		}
	}

	// Update description if provided
	if req.Description != nil && *req.Description != "" {
		utils.Info("Updating description for user %s", username)
		if err := s.modifyLDAPAttribute(userDN, "description", *req.Description); err != nil {
			utils.Error("Failed to update description: %v", err)
			return fmt.Errorf("failed to update description: %v", err)
		}
	}

	// Update enabled status if provided
	if req.Enabled != nil {
		if *req.Enabled {
			output, err := s.sambaTool.UserEnable(username)
			if err != nil {
				return fmt.Errorf("failed to enable user: %s", output)
			}
		} else {
			output, err := s.sambaTool.UserDisable(username)
			if err != nil {
				return fmt.Errorf("failed to disable user: %s", output)
			}
		}
	}

	// Update group membership if provided
	if req.Group != nil {
		utils.Info("Updating group membership for user %s to: %s", username, *req.Group)

		// First, get current groups the user belongs to
		currentGroups, err := s.getUserGroups(username)
		if err != nil {
			utils.Warn("Could not get current groups for user %s: %v", username, err)
			currentGroups = []string{}
		}

		// Remove user from all non-system groups
		for _, group := range currentGroups {
			// Don't remove from system groups
			if group == "Domain Users" || group == "Users" {
				continue
			}

			utils.Info("Removing user %s from group: %s", username, group)
			output, err := s.sambaTool.Run("group", "removemembers", group, username)
			if err != nil {
				utils.Warn("Failed to remove user %s from group %s: %s", username, group, output)
			}
		}

		// Add user to new group if specified
		if *req.Group != "" && *req.Group != "Domain Users" {
			utils.Info("Adding user %s to group: %s", username, *req.Group)
			output, err := s.sambaTool.Run("group", "addmembers", *req.Group, username)
			if err != nil {
				return fmt.Errorf("failed to add user to group %s: %s", *req.Group, output)
			}
			utils.Info("Successfully added user %s to group %s", username, *req.Group)
		}
	}

	// TODO: Update OU path if provided (this requires moving the user object in LDAP)
	// This is more complex and would require LDAP modify operations

	return nil
}

// DeleteUser removes a user from the domain
func (s *UserService) DeleteUser(username string) error {

	output, err := s.sambaTool.UserDelete(username)
	if err != nil {
		return fmt.Errorf("failed to delete user: %s", output)
	}
	return nil
}

// DisableUser disables a user account
func (s *UserService) DisableUser(username string) error {

	output, err := s.sambaTool.UserDisable(username)
	if err != nil {
		return fmt.Errorf("failed to disable user: %s", output)
	}
	return nil
}

// EnableUser enables a user account
func (s *UserService) EnableUser(username string) error {

	output, err := s.sambaTool.UserEnable(username)
	if err != nil {
		return fmt.Errorf("failed to enable user: %s", output)
	}
	return nil
}

// addUserToGroup adds a user to a group
func (s *UserService) addUserToGroup(username, groupName string) error {
	output, err := s.sambaTool.GroupAddMembers(groupName, []string{username})
	if err != nil {
		return fmt.Errorf("failed to add user to group: %s", output)
	}
	return nil
}

// ChangeUserPassword changes a user's password
func (s *UserService) ChangeUserPassword(username, newPassword string) error {
	utils.Info("Changing password for user: %s (password length: %d)", username, len(newPassword))

	// Log password complexity (without revealing the password)
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false
	for _, ch := range newPassword {
		if ch >= 'A' && ch <= 'Z' {
			hasUpper = true
		} else if ch >= 'a' && ch <= 'z' {
			hasLower = true
		} else if ch >= '0' && ch <= '9' {
			hasDigit = true
		} else {
			hasSpecial = true
		}
	}
	utils.Info("Password complexity: length=%d, upper=%v, lower=%v, digit=%v, special=%v",
		len(newPassword), hasUpper, hasLower, hasDigit, hasSpecial)

	output, err := s.sambaTool.UserSetPassword(username, newPassword)
	if err != nil {
		utils.Error("Failed to change password for user %s: %v", username, err)
		utils.Error("Samba output: %s", output)
		return fmt.Errorf("password change failed: %s", output)
	}
	utils.Info("Password changed successfully for user: %s", username)
	return nil
}

// UpdateUserProfile updates a user's profile information
func (s *UserService) UpdateUserProfile(username, fullName, email string) error {
	utils.Info("Updating profile for user: %s (FullName: %s, Email: %s)", username, fullName, email)

	// Get user's DN first
	userDN, err := s.getUserDN(username)
	if err != nil {
		utils.Error("Failed to get user DN for %s: %v", username, err)
		return fmt.Errorf("failed to get user DN: %v", err)
	}

	// Update full name if provided
	if fullName != "" {
		utils.Info("Updating full name for user %s to: %s", username, fullName)
		if err := s.modifyLDAPAttribute(userDN, "givenName", fullName); err != nil {
			utils.Error("Failed to update full name: %v", err)
			return fmt.Errorf("failed to update full name: %v", err)
		}
		// Also update displayName
		if err := s.modifyLDAPAttribute(userDN, "displayName", fullName); err != nil {
			utils.Warn("Failed to update display name: %v", err)
		}
		utils.Info("Full name updated successfully for user: %s", username)
	}

	// Update email if provided
	if email != "" {
		utils.Info("Updating email for user %s to: %s", username, email)
		if err := s.modifyLDAPAttribute(userDN, "mail", email); err != nil {
			utils.Error("Failed to update email: %v", err)
			return fmt.Errorf("failed to update email: %v", err)
		}
		utils.Info("Email updated successfully for user: %s", username)
	}

	utils.Info("Profile update completed successfully for user: %s", username)
	return nil
}

// SetMustChangePassword sets the "must change password at next login" flag
func (s *UserService) SetMustChangePassword(username string) error {
	// Use LDAP modify to set pwdLastSet to 0 (forces password change on next login)
	_, err := s.sambaTool.Run("user", "setpassword", username, "--newpassword=PLACEHOLDER", "--must-change")
	if err != nil {
		// If that doesn't work, try setting pwdLastSet to 0 via LDAP
		output2, err2 := s.sambaTool.Run("ldapmodify", "-H", "ldapi://", "-Y", "GSSAPI", "-x", "-D", "cn=admin,dc=vfw5788,dc=local", "-W")
		if err2 != nil {
			return fmt.Errorf("failed to set must-change-password flag: %s", output2)
		}
	}
	return nil
}

// ClearMustChangePassword clears the "must change password at next login" flag
func (s *UserService) ClearMustChangePassword(username string) error {
	// This is more complex - we need to set pwdLastSet to current time
	// For now, we'll use a simple approach: reset password with current password
	return nil // TODO: Implement this properly
}

// ToggleMustChangePassword toggles the must-change-password flag
func (s *UserService) ToggleMustChangePassword(username string) error {
	// Check current state first
	output, err := s.sambaTool.UserShow(username)
	if err != nil {
		return fmt.Errorf("failed to get user info: %s", output)
	}

	// Simple approach: just set the flag (we'll improve this later)
	return s.SetMustChangePassword(username)
}

// getUserDN gets the DN (Distinguished Name) for a user
func (s *UserService) getUserDN(username string) (string, error) {
	output, err := s.sambaTool.UserShow(username)
	if err != nil {
		return "", err
	}

	// Parse the DN from the output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "dn:") {
			dn := strings.TrimSpace(strings.TrimPrefix(line, "dn:"))
			return dn, nil
		}
	}

	return "", fmt.Errorf("could not find DN for user %s", username)
}

// modifyLDAPAttribute modifies a single LDAP attribute for a user
func (s *UserService) modifyLDAPAttribute(userDN, attribute, value string) error {
	// Create LDIF content for the modification
	ldif := fmt.Sprintf(`dn: %s
changetype: modify
replace: %s
%s: %s
`, userDN, attribute, attribute, value)

	// Use ldbmodify to apply the change (without relax control to avoid affecting passwords)
	cmd, cmdErr := utils.SafeCommand("ldbmodify", "-H", "/var/lib/samba/private/sam.ldb")
	if cmdErr != nil {
		return fmt.Errorf("command sanitization failed: %v", cmdErr)
	}

	// Pass LDIF via stdin
	cmd.Stdin = strings.NewReader(ldif)
	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.Error("ldbmodify failed for attribute %s: %s", attribute, string(output))
		return fmt.Errorf("ldbmodify failed: %s", string(output))
	}

	utils.Debug("Successfully modified attribute %s for DN: %s", attribute, userDN)
	return nil
}
