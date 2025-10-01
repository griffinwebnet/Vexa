package models

// Group represents a domain group
type Group struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Members     []string `json:"members"`
}

// CreateGroupRequest represents the request to create a new group
type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateGroupRequest represents the request to update an existing group
type UpdateGroupRequest struct {
	Description *string `json:"description,omitempty"`
}

// AddGroupMembersRequest represents the request to add members to a group
type AddGroupMembersRequest struct {
	Members []string `json:"members" binding:"required"`
}

// RemoveGroupMembersRequest represents the request to remove members from a group
type RemoveGroupMembersRequest struct {
	Members []string `json:"members" binding:"required"`
}
