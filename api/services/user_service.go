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

// ListUsers returns all users in the domain
func (s *UserService) ListUsers() ([]models.User, error) {

	output, err := s.sambaTool.UserList()
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %s", output)
	}

	usernames := s.sambaTool.ParseUserList(output)
	users := make([]models.User, 0, len(usernames))

	for _, username := range usernames {
		users = append(users, models.User{
			Username: username,
			Enabled:  true, // TODO: Get actual status
		})
	}

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
	// Update full name if provided
	if req.FullName != nil {
		cmd, cmdErr := utils.SafeCommand("samba-tool", "user", "set", username, "--full-name="+*req.FullName)
		if cmdErr != nil {
			return fmt.Errorf("command sanitization failed: %v", cmdErr)
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to update full name: %s", string(output))
		}
	}

	// Update email if provided
	if req.Email != nil {
		cmd, cmdErr := utils.SafeCommand("samba-tool", "user", "set", username, "--mail-address="+*req.Email)
		if cmdErr != nil {
			return fmt.Errorf("command sanitization failed: %v", cmdErr)
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to update email: %s", string(output))
		}
	}

	// Update description if provided
	if req.Description != nil {
		cmd, cmdErr := utils.SafeCommand("samba-tool", "user", "set", username, "--description="+*req.Description)
		if cmdErr != nil {
			return fmt.Errorf("command sanitization failed: %v", cmdErr)
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to update description: %s", string(output))
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
	if req.Group != nil && *req.Group != "" && *req.Group != "Domain Users" {
		// Remove from current groups and add to new group
		// For now, we'll just add to the new group
		output, err := s.sambaTool.Run("group", "addmembers", *req.Group, username)
		if err != nil {
			return fmt.Errorf("failed to update group membership: %s", output)
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
	fmt.Printf("DEBUG: Changing password for user: %s\n", username)
	output, err := s.sambaTool.UserSetPassword(username, newPassword)
	if err != nil {
		fmt.Printf("DEBUG: Failed to change password for %s: %v, output: %s\n", username, err, output)
		return fmt.Errorf("failed to change password: %s", output)
	}
	fmt.Printf("DEBUG: Password changed successfully for user: %s\n", username)
	return nil
}

// UpdateUserProfile updates a user's profile information
func (s *UserService) UpdateUserProfile(username, fullName, email string) error {
	fmt.Printf("DEBUG: Updating profile for %s - FullName: %s, Email: %s\n", username, fullName, email)

	// Update full name if provided
	if fullName != "" {
		cmd, cmdErr := utils.SafeCommand("samba-tool", "user", "set", username, "--full-name="+fullName)
		if cmdErr != nil {
			return fmt.Errorf("command sanitization failed: %v", cmdErr)
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to update full name: %s", string(output))
		}
		fmt.Printf("DEBUG: Full name updated successfully for %s\n", username)
	}

	// Update email if provided
	if email != "" {
		cmd, cmdErr := utils.SafeCommand("samba-tool", "user", "set", username, "--mail-address="+email)
		if cmdErr != nil {
			return fmt.Errorf("command sanitization failed: %v", cmdErr)
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to update email: %s", string(output))
		}
		fmt.Printf("DEBUG: Email updated successfully for %s\n", username)
	}

	return nil
}

// SetMustChangePassword sets the "must change password at next login" flag
func (s *UserService) SetMustChangePassword(username string) error {
	// Use LDAP modify to set pwdLastSet to 0 (forces password change on next login)
	output, err := s.sambaTool.Run("user", "setpassword", username, "--newpassword=PLACEHOLDER", "--must-change")
	if err != nil {
		// If that doesn't work, try setting pwdLastSet to 0 via LDAP
		output, err = s.sambaTool.Run("ldapmodify", "-H", "ldapi://", "-Y", "GSSAPI", "-x", "-D", "cn=admin,dc=vfw5788,dc=local", "-W")
		if err != nil {
			return fmt.Errorf("failed to set must-change-password flag: %s", output)
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
