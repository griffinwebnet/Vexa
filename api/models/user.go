package models

// User represents a domain user
type User struct {
	Username    string   `json:"username"`
	FullName    string   `json:"full_name"`
	Email       string   `json:"email"`
	Enabled     bool     `json:"enabled"`
	Groups      []string `json:"groups"`
	Description string   `json:"description"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Username           string `json:"username" binding:"required"`
	Password           string `json:"password" binding:"required"`
	FullName           string `json:"full_name"`
	Email              string `json:"email"`
	Description        string `json:"description"`
	Group              string `json:"group"`
	OUPath             string `json:"ou_path"`
	MustChangePassword bool   `json:"must_change_password"`
}

// UpdateUserRequest represents the request to update an existing user
type UpdateUserRequest struct {
	FullName    *string `json:"full_name,omitempty"`
	Email       *string `json:"email,omitempty"`
	Description *string `json:"description,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
	Group       *string `json:"group,omitempty"`
	OUPath      *string `json:"ou_path,omitempty"`
}
