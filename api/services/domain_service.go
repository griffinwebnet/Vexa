package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os/exec"
	"strings"
	"time"

	vexaexec "github.com/griffinwebnet/vexa/api/exec"
	"github.com/griffinwebnet/vexa/api/models"
	"github.com/griffinwebnet/vexa/api/utils"
)

// DomainService handles domain-related business logic
type DomainService struct {
	sambaTool *vexaexec.SambaTool
	system    *vexaexec.System
}

// NewDomainService creates a new DomainService instance
func NewDomainService() *DomainService {
	return &DomainService{
		sambaTool: vexaexec.NewSambaTool(),
		system:    vexaexec.NewSystem(),
	}
}

// GetDomainStatus returns the current status of the domain controller
func (s *DomainService) GetDomainStatus() (*models.DomainStatusResponse, error) {
	utils.Info("Checking domain controller status")
	// Check if domain is provisioned using multiple methods
	provisioned := false

	// Method 1: Try samba-tool domain info
	_, err := s.sambaTool.DomainInfo("127.0.0.1")
	if err == nil {
		provisioned = true
	}

	// Method 2: Check if Samba configuration exists
	if !provisioned {
		// Check if /etc/samba/smb.conf exists and has domain configuration
		if s.system.FileExists("/etc/samba/smb.conf") {
			// Try to parse the config for domain settings
			if s.hasDomainConfiguration() {
				provisioned = true
			}
		}
	}

	// Method 3: Check if Samba databases exist
	if !provisioned {
		if s.system.FileExists("/var/lib/samba/private/sam.ldb") {
			provisioned = true
		}
	}

	// Check if DC is running
	dcReady, _ := s.system.ServiceStatus("samba-ad-dc")

	response := &models.DomainStatusResponse{
		Provisioned: provisioned,
		DCReady:     dcReady,
		DNSReady:    dcReady, // DNS is internal to Samba
	}

	if provisioned {
		// Try multiple methods to get domain info
		domain, realm := s.getDomainInfo()

		if domain != "" {
			response.Domain = domain
		} else {
			response.Domain = "PROVISIONED"
		}

		if realm != "" {
			response.Realm = realm
		} else {
			response.Realm = "PROVISIONED"
		}

		fmt.Printf("DEBUG: Final domain info - Domain: %s, Realm: %s\n", response.Domain, response.Realm)
	}

	return response, nil
}

// hasDomainConfiguration checks if the Samba config has domain controller settings
func (s *DomainService) hasDomainConfiguration() bool {
	cmd := exec.Command("testparm", "-s", "--parameter-name", "server role")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	role := strings.TrimSpace(string(output))
	return strings.Contains(strings.ToLower(role), "active directory domain controller")
}

// getDomainInfo tries multiple methods to get domain and realm information
func (s *DomainService) getDomainInfo() (string, string) {
	fmt.Printf("DEBUG: Attempting to get domain info\n")

	// Method 1: Try samba-tool domain info
	domain, realm := s.parseSambaToolDomainInfo()
	if domain != "" && realm != "" {
		fmt.Printf("DEBUG: Got domain info from samba-tool: %s, %s\n", domain, realm)
		return domain, realm
	}

	// Method 2: Try parsing smb.conf
	domain2, realm2 := s.parseSmbConf()
	if domain2 != "" || realm2 != "" {
		fmt.Printf("DEBUG: Got domain info from smb.conf: %s, %s\n", domain2, realm2)
		if domain == "" {
			domain = domain2
		}
		if realm == "" {
			realm = realm2
		}
	}

	// Method 3: Try testparm output
	domain3, realm3 := s.parseTestparm()
	if domain3 != "" || realm3 != "" {
		fmt.Printf("DEBUG: Got domain info from testparm: %s, %s\n", domain3, realm3)
		if domain == "" {
			domain = domain3
		}
		if realm == "" {
			realm = realm3
		}
	}

	fmt.Printf("DEBUG: Final parsed domain info: %s, %s\n", domain, realm)
	return domain, realm
}

// parseSambaToolDomainInfo parses output from samba-tool domain info
func (s *DomainService) parseSambaToolDomainInfo() (string, string) {
	output, err := s.sambaTool.DomainInfo("127.0.0.1")
	if err != nil {
		fmt.Printf("DEBUG: samba-tool domain info failed: %v\n", err)
		return "", ""
	}

	fmt.Printf("DEBUG: samba-tool domain info output:\n%s\n", output)

	var domain, realm string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for NetBIOS Domain (short domain name)
		if strings.Contains(line, "NetBIOS Domain:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				domain = strings.TrimSpace(parts[1])
				fmt.Printf("DEBUG: Found NetBIOS Domain: %s\n", domain)
			}
		}
		// Look for DNS Domain (realm)
		if strings.Contains(line, "DNS Domain:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				realm = strings.TrimSpace(parts[1])
				fmt.Printf("DEBUG: Found DNS Domain: %s\n", realm)
			}
		}
	}

	return domain, realm
}

// parseSmbConf parses the smb.conf file for domain info
func (s *DomainService) parseSmbConf() (string, string) {
	// This would read /etc/samba/smb.conf and parse workgroup and realm
	// For now, return empty - implement if needed
	return "", ""
}

// parseTestparm parses testparm output for domain info
func (s *DomainService) parseTestparm() (string, string) {
	cmd := exec.Command("testparm", "-s", "--parameter-name", "workgroup")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("DEBUG: testparm workgroup failed: %v\n", err)
		return "", ""
	}

	domain := strings.TrimSpace(string(output))
	fmt.Printf("DEBUG: testparm workgroup output: %s\n", domain)

	// Get realm
	cmd2 := exec.Command("testparm", "-s", "--parameter-name", "realm")
	output2, err2 := cmd2.Output()
	if err2 != nil {
		fmt.Printf("DEBUG: testparm realm failed: %v\n", err2)
		return domain, ""
	}

	realm := strings.TrimSpace(string(output2))
	fmt.Printf("DEBUG: testparm realm output: %s\n", realm)

	return domain, realm
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
		options := vexaexec.GroupCreateOptions{
			Description: "", // Empty description for default groups
		}
		s.sambaTool.GroupCreate(group, options)
	}

	return nil
}

// ProvisionDomainWithOutput provisions a new domain with streaming CLI output
func (s *DomainService) ProvisionDomainWithOutput(req models.ProvisionDomainRequest, outputChan chan<- string) error {
	utils.Info("Starting domain provisioning for domain: %s, realm: %s", req.Domain, req.Realm)

	// Set defaults
	if req.DNSBackend == "" {
		req.DNSBackend = "SAMBA_INTERNAL"
	}

	outputChan <- "Starting domain provisioning..."
	outputChan <- fmt.Sprintf("Domain: %s, Realm: %s", req.Domain, req.Realm)

	// Check if samba-tool is available
	outputChan <- "Checking if samba-tool is available..."
	utils.Info("Checking samba-tool availability")
	_, err := s.sambaTool.Run("--version")
	if err != nil {
		utils.Error("samba-tool not available: %v", err)
		outputChan <- "ERROR: samba-tool not found or not accessible"
		outputChan <- "ERROR: Please ensure Samba is installed: apt install samba samba-tools"
		return fmt.Errorf("samba-tool not available: %v", err)
	}
	utils.Info("samba-tool is available")
	outputChan <- "samba-tool is available"

	// Clean up existing Samba configuration to avoid conflicts
	outputChan <- "Cleaning up existing Samba configuration..."
	s.system.RemoveFile("/etc/samba/smb.conf")
	outputChan <- "Removed /etc/samba/smb.conf"

	// Stop conflicting services
	outputChan <- "Stopping conflicting services..."
	services := []string{"smbd", "nmbd", "winbind", "samba-ad-dc"}
	for _, service := range services {
		outputChan <- fmt.Sprintf("Stopping %s...", service)
		s.system.StopService(service)
		time.Sleep(1 * time.Second) // Give services time to stop
	}

	// Clean up Samba databases completely to avoid corruption/bugs
	outputChan <- "Cleaning up Samba databases..."
	files := []string{
		"/var/lib/samba/private/sam.ldb",
		"/var/lib/samba/private/secrets.ldb",
		"/var/cache/samba/gencache.tdb",
	}
	for _, file := range files {
		outputChan <- fmt.Sprintf("Removing %s...", file)
		s.system.RemoveFile(file)
	}

	// Generate a secure admin password
	outputChan <- "Generating admin password..."
	adminPassword := generateAdminPassword()
	outputChan <- "Admin password generated"

	options := vexaexec.DomainProvisionOptions{
		Domain:        req.Domain,
		Realm:         req.Realm,
		AdminPassword: adminPassword,
		DNSBackend:    req.DNSBackend,
		DNSForwarder:  req.DNSForwarder,
	}

	outputChan <- "Starting domain provision command..."
	outputChan <- fmt.Sprintf("Command: samba-tool domain provision --realm=%s --domain=%s --server-role=dc --dns-backend=%s", req.Realm, req.Domain, req.DNSBackend)
	if req.DNSForwarder != "" {
		outputChan <- fmt.Sprintf("DNS Forwarder: %s", req.DNSForwarder)
	}

	output, err := s.sambaTool.DomainProvisionWithOutput(options, outputChan)
	if err != nil {
		outputChan <- fmt.Sprintf("ERROR: Domain provisioning failed: %s", output)
		outputChan <- "ERROR: Command exit code indicates failure"
		outputChan <- "ERROR: Check if samba-tool is installed and accessible"
		outputChan <- "ERROR: Verify system permissions for domain provisioning"
		return fmt.Errorf("domain provisioning failed: %s", output)
	}

	outputChan <- "Domain provisioning completed successfully"
	outputChan <- "Creating default groups..."

	// Create default groups
	if err := s.createDefaultGroupsWithOutput(outputChan); err != nil {
		outputChan <- fmt.Sprintf("WARNING: Failed to create default groups: %v", err)
		// Don't fail the entire operation for this
	}

	outputChan <- "Starting Samba AD DC service..."
	// Start Samba service
	if err := s.system.EnableAndStartService("samba-ad-dc"); err != nil {
		outputChan <- fmt.Sprintf("ERROR: Failed to start Samba AD DC service: %v", err)
		return fmt.Errorf("failed to start Samba AD DC service: %w", err)
	}

	outputChan <- "Samba AD DC service started successfully"
	outputChan <- "Domain provisioning completed!"

	return nil
}

// createDefaultGroupsWithOutput creates default groups with output streaming
func (s *DomainService) createDefaultGroupsWithOutput(outputChan chan<- string) error {
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
		outputChan <- fmt.Sprintf("Creating group: %s", group)
		// Try to create group, ignore if it already exists
		options := vexaexec.GroupCreateOptions{
			Description: "", // Empty description for default groups
		}
		output, err := s.sambaTool.GroupCreate(group, options)
		if err != nil {
			outputChan <- fmt.Sprintf("Group %s creation result: %s", group, output)
			// Continue with other groups even if one fails
		} else {
			outputChan <- fmt.Sprintf("Group %s created successfully", group)
		}
		time.Sleep(500 * time.Millisecond) // Small delay between group creation
	}

	return nil
}

// generateAdminPassword generates a secure admin password
func generateAdminPassword() string {
	// Generate a cryptographically secure random password
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	const length = 16

	// Use crypto/rand for secure random generation
	bytes := make([]byte, length)
	for i := range bytes {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback to a default secure password if random generation fails
			return "SecurePass123!@#$"
		}
		bytes[i] = charset[n.Int64()]
	}

	// Ensure at least one uppercase, lowercase, digit, and special character
	password := string(bytes)

	// Replace first 4 characters to ensure complexity requirements
	password = "A" + password[1:]                // Uppercase
	password = password[:1] + "a" + password[2:] // Lowercase
	password = password[:2] + "1" + password[3:] // Digit
	password = password[:3] + "!" + password[4:] // Special char

	return password
}
