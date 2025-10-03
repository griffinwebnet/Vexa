package services

import (
	"fmt"
	"os"

	"github.com/vexa/api/exec"
)

// DNSService handles DNS-related business logic
type DNSService struct {
	sambaTool *exec.SambaTool
}

// NewDNSService creates a new DNSService instance
func NewDNSService() *DNSService {
	return &DNSService{
		sambaTool: exec.NewSambaTool(),
	}
}

// GetDNSStatus returns the current DNS server status and configuration
func (s *DNSService) GetDNSStatus() (map[string]interface{}, error) {

	// TODO: Implement real Samba DNS status check
	return nil, fmt.Errorf("DNS status check not implemented for production mode")
}

// UpdateDNSForwarders updates the DNS forwarder configuration
func (s *DNSService) UpdateDNSForwarders(primary, secondary string) error {
	if os.Getenv("ENV") == "development" {
		// Simulate successful update in development mode
		return nil
	}

	// Use Samba DNS forwarder command
	forwarders := []string{primary, secondary}
	output, err := s.sambaTool.DNSForwarders(forwarders)
	if err != nil {
		return fmt.Errorf("failed to update DNS forwarders: %s", output)
	}

	return nil
}
