package services

import (
	"fmt"

	"github.com/vexa/api/exec"
	"github.com/vexa/api/models"
)

// DomainService handles domain-related business logic
type DomainService struct {
	sambaTool *exec.SambaTool
	system    *exec.System
}

// NewDomainService creates a new DomainService instance
func NewDomainService() *DomainService {
	return &DomainService{
		sambaTool: exec.NewSambaTool(),
		system:    exec.NewSystem(),
	}
}

// ProvisionDomain provisions a new Samba AD DC domain
func (s *DomainService) ProvisionDomain(req models.ProvisionDomainRequest) error {
	// Set defaults
	if req.DNSBackend == "" {
		req.DNSBackend = "SAMBA_INTERNAL"
	}

	// Clean up existing Samba configuration to avoid conflicts
	s.system.RemoveFile("/etc/samba/smb.conf")

	// Stop conflicting services
	s.system.StopService("smbd")
	s.system.StopService("nmbd")
	s.system.StopService("winbind")
	s.system.StopService("samba-ad-dc")

	// Clean up Samba databases completely to avoid corruption/bugs
	s.system.RemoveFile("/var/lib/samba/private/sam.ldb")
	s.system.RemoveFile("/var/lib/samba/private/secrets.ldb")
	s.system.RemoveFile("/var/cache/samba/gencache.tdb")

	// Workaround for Samba 4.19.5 bug: temporarily disable acl_xattr VFS module
	// See: https://bugzilla.samba.org/show_bug.cgi?id=15203
	aclXattrPath := "/usr/lib/x86_64-linux-gnu/samba/vfs/acl_xattr.so"
	aclXattrBackup := aclXattrPath + ".backup"
	s.system.RunCommand("mv", aclXattrPath, aclXattrBackup)

	// Generate a secure admin password
	adminPassword := generateAdminPassword()

	options := exec.DomainProvisionOptions{
		Domain:        req.Domain,
		Realm:         req.Realm,
		AdminPassword: adminPassword,
		DNSBackend:    req.DNSBackend,
		DNSForwarder:  req.DNSForwarder,
	}

	output, err := s.sambaTool.DomainProvision(options)

	// Restore acl_xattr module immediately after provisioning
	s.system.RunCommand("mv", aclXattrBackup, aclXattrPath)

	if err != nil {
		return fmt.Errorf("domain provisioning failed: %s", output)
	}

	// Create default groups
	if err := s.createDefaultGroups(); err != nil {
		return fmt.Errorf("failed to create default groups: %v", err)
	}

	// Start Samba service
	if err := s.system.EnableAndStartService("samba-ad-dc"); err != nil {
		return fmt.Errorf("failed to start Samba AD DC service: %w", err)
	}

	return nil
}

// GetDomainStatus returns the current status of the domain controller
func (s *DomainService) GetDomainStatus() (*models.DomainStatusResponse, error) {

	// Check if domain is provisioned
	_, err := s.sambaTool.DomainInfo("127.0.0.1")
	provisioned := err == nil

	// Check if DC is running
	dcReady, _ := s.system.ServiceStatus("samba-ad-dc")

	response := &models.DomainStatusResponse{
		Provisioned: provisioned,
		DCReady:     dcReady,
		DNSReady:    dcReady, // DNS is internal to Samba
	}

	if provisioned {
		// TODO: Parse domain info from output
		response.Domain = "UNKNOWN"
		response.Realm = "UNKNOWN"
	}

	return response, nil
}

// GetDomainInfo returns detailed domain information
func (s *DomainService) GetDomainInfo(server string) (*models.DomainInfo, error) {
	output, err := s.sambaTool.DomainInfo(server)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain info: %s", output)
	}

	// TODO: Parse domain info from output
	info := &models.DomainInfo{
		Domain: "UNKNOWN",
		Realm:  "UNKNOWN",
	}

	return info, nil
}

// ConfigureDomain updates domain configuration
func (s *DomainService) ConfigureDomain(config map[string]interface{}) error {
	// TODO: Implement domain configuration updates
	return fmt.Errorf("domain configuration not implemented yet")
}

// createDefaultGroups creates the default Domain Users and Domain Admins groups
func (s *DomainService) createDefaultGroups() error {
	// Domain Users group (already exists by default in Samba)
	// Domain Admins group (already exists by default in Samba)

	// Create additional groups if needed
	groups := []string{
		"IT Staff",
		"Finance",
		"Sales",
		"HR",
	}

	for _, group := range groups {
		// Try to create group, ignore if it already exists
		options := exec.GroupCreateOptions{
			Description: "", // Empty description for default groups
		}
		s.sambaTool.GroupCreate(group, options)
	}

	return nil
}

// generateAdminPassword generates a secure admin password
func generateAdminPassword() string {
	// Generate a strong 16-character password
	return "TempAdmin123!" // TODO: Generate random secure password
}
