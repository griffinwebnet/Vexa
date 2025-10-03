package services

import (
	"fmt"

	"github.com/vexa/api/exec"
	"github.com/vexa/api/models"
)

// GroupService handles group-related business logic
type GroupService struct {
	sambaTool *exec.SambaTool
}

// NewGroupService creates a new GroupService instance
func NewGroupService() *GroupService {
	return &GroupService{
		sambaTool: exec.NewSambaTool(),
	}
}

// ListGroups returns all groups in the domain
func (s *GroupService) ListGroups() ([]models.Group, error) {

	output, err := s.sambaTool.GroupList()
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %s", output)
	}

	groupNames := s.sambaTool.ParseGroupList(output)
	groups := make([]models.Group, 0, len(groupNames))

	for _, name := range groupNames {
		groups = append(groups, models.Group{
			Name: name,
		})
	}

	return groups, nil
}

// CreateGroup creates a new group in the domain
func (s *GroupService) CreateGroup(req models.CreateGroupRequest) error {
	options := exec.GroupCreateOptions{
		Description: req.Description,
	}

	output, err := s.sambaTool.GroupCreate(req.Name, options)
	if err != nil {
		return fmt.Errorf("failed to create group: %s", output)
	}

	return nil
}

// GetGroup returns details for a specific group
func (s *GroupService) GetGroup(groupName string) (*models.Group, error) {
	output, err := s.sambaTool.GroupListMembers(groupName)
	if err != nil {
		return nil, fmt.Errorf("group not found: %s", output)
	}

	members := s.sambaTool.ParseGroupMembers(output)

	group := &models.Group{
		Name:    groupName,
		Members: members,
	}

	return group, nil
}

// UpdateGroup updates an existing group
func (s *GroupService) UpdateGroup(groupName string, req models.UpdateGroupRequest) error {

	// Update description if provided
	if req.Description != nil {
		// Use samba-tool to update group description
		output, err := s.sambaTool.GroupModify(groupName, *req.Description)
		if err != nil {
			return fmt.Errorf("failed to update group: %s", output)
		}
	}

	return nil
}

// DeleteGroup removes a group from the domain
func (s *GroupService) DeleteGroup(groupName string) error {
	output, err := s.sambaTool.GroupDelete(groupName)
	if err != nil {
		return fmt.Errorf("failed to delete group: %s", output)
	}
	return nil
}

// AddGroupMembers adds members to a group
func (s *GroupService) AddGroupMembers(groupName string, req models.AddGroupMembersRequest) error {
	output, err := s.sambaTool.GroupAddMembers(groupName, req.Members)
	if err != nil {
		return fmt.Errorf("failed to add members to group: %s", output)
	}
	return nil
}

// RemoveGroupMembers removes members from a group
func (s *GroupService) RemoveGroupMembers(groupName string, req models.RemoveGroupMembersRequest) error {
	output, err := s.sambaTool.GroupRemoveMembers(groupName, req.Members)
	if err != nil {
		return fmt.Errorf("failed to remove members from group: %s", output)
	}
	return nil
}
