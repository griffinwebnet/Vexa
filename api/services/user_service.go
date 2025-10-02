package services

import (
	"fmt"
	"os"

	"github.com/vexa/api/exec"
	"github.com/vexa/api/models"
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
	// Development mode: Return dummy data for UI testing
	if os.Getenv("ENV") == "development" {
		return s.getDevUsers(), nil
	}

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

	return nil
}

// GetUser returns details for a specific user
func (s *UserService) GetUser(username string) (*models.User, error) {
	output, err := s.sambaTool.UserShow(username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %s", output)
	}

	// TODO: Parse user details from output
	user := &models.User{
		Username: username,
		Enabled:  true,
	}

	return user, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(username string, req models.UpdateUserRequest) error {
	// TODO: Implement user update logic
	return fmt.Errorf("user update not implemented yet")
}

// DeleteUser removes a user from the domain
func (s *UserService) DeleteUser(username string) error {
	// Development mode: Simulate deletion
	if os.Getenv("ENV") == "development" {
		return nil
	}

	output, err := s.sambaTool.UserDelete(username)
	if err != nil {
		return fmt.Errorf("failed to delete user: %s", output)
	}
	return nil
}

// DisableUser disables a user account
func (s *UserService) DisableUser(username string) error {
	// Development mode: Simulate disable
	if os.Getenv("ENV") == "development" {
		return nil
	}

	output, err := s.sambaTool.UserDisable(username)
	if err != nil {
		return fmt.Errorf("failed to disable user: %s", output)
	}
	return nil
}

// EnableUser enables a user account
func (s *UserService) EnableUser(username string) error {
	// Development mode: Simulate enable
	if os.Getenv("ENV") == "development" {
		return nil
	}

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

// getDevUsers returns dummy users for UI development
func (s *UserService) getDevUsers() []models.User {
	return []models.User{
		{Username: "jsmith", FullName: "John Smith", Email: "jsmith@example.com", Enabled: true, Groups: []string{"Domain Users", "IT Staff"}},
		{Username: "mjohnson", FullName: "Mary Johnson", Email: "mjohnson@example.com", Enabled: true, Groups: []string{"Domain Users", "Finance"}},
		{Username: "bwilliams", FullName: "Bob Williams", Email: "bwilliams@example.com", Enabled: true, Groups: []string{"Domain Users", "Sales"}},
		{Username: "administrator", FullName: "Administrator", Email: "admin@example.com", Enabled: true, Groups: []string{"Domain Admins", "Domain Users"}},
		{Username: "sarah.lee", FullName: "Sarah Lee", Email: "sarah.lee@example.com", Enabled: true, Groups: []string{"Domain Users", "HR"}},
		{Username: "mike.chen", FullName: "Mike Chen", Email: "mike.chen@example.com", Enabled: false, Groups: []string{"Domain Users"}},
	}
}
