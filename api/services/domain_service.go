package services

import (
	"fmt"
	"os"

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

	options := exec.DomainProvisionOptions{
		Domain:        req.Domain,
		Realm:         req.Realm,
		AdminPassword: req.AdminPassword,
		DNSBackend:    req.DNSBackend,
		DNSForwarder:  req.DNSForwarder,
	}

	output, err := s.sambaTool.DomainProvision(options)
	if err != nil {
		return fmt.Errorf("domain provisioning failed: %s", output)
	}

	// Start Samba service
	if err := s.system.EnableAndStartService("samba-ad-dc"); err != nil {
		return fmt.Errorf("failed to start Samba AD DC service: %w", err)
	}

	return nil
}

// GetDomainStatus returns the current status of the domain controller
func (s *DomainService) GetDomainStatus() (*models.DomainStatusResponse, error) {
	// Dev mode: Return dummy status
	if os.Getenv("ENV") == "development" {
		return &models.DomainStatusResponse{
			Provisioned: true,
			Domain:      "EXAMPLE",
			Realm:       "example.local",
			DCReady:     true,
			DNSReady:    true,
		}, nil
	}

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
