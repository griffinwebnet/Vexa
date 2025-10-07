package services

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/griffinwebnet/vexa/api/utils"
)

// DDNSProvider represents a DDNS provider configuration
type DDNSProvider struct {
	Name     string `json:"name"`     // Provider name (cloudflare, dynu, duckdns)
	Enabled  bool   `json:"enabled"`  // Whether DDNS is enabled
	Domain   string `json:"domain"`   // Domain to update
	Username string `json:"username"` // API username/email/token
	Password string `json:"password"` // API password/key
	Interval int    `json:"interval"` // Update interval in minutes
}

// DDNSService handles dynamic DNS updates using ddclient
type DDNSService struct{}

// NewDDNSService creates a new DDNSService
func NewDDNSService() *DDNSService {
	return &DDNSService{}
}

// SetupDDNS configures ddclient for DDNS updates
func (s *DDNSService) SetupDDNS(provider *DDNSProvider) error {
	// Save config
	if err := s.saveConfig(provider); err != nil {
		return err
	}

	// Create ddclient config
	if err := s.createDDClientConfig(provider); err != nil {
		return err
	}

	// Start/stop service based on enabled state
	if provider.Enabled {
		cmd, cmdErr := utils.SafeCommand("systemctl", "enable", "--now", "ddclient")
		if cmdErr != nil {
			return fmt.Errorf("command sanitization failed: %v", cmdErr)
		}
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start ddclient: %v", err)
		}
	} else {
		cmd, cmdErr := utils.SafeCommand("systemctl", "disable", "--now", "ddclient")
		if cmdErr != nil {
			return fmt.Errorf("command sanitization failed: %v", cmdErr)
		}
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stop ddclient: %v", err)
		}
	}

	return nil
}

// createDDClientConfig creates ddclient configuration
func (s *DDNSService) createDDClientConfig(provider *DDNSProvider) error {
	var config string

	switch provider.Name {
	case "cloudflare":
		config = fmt.Sprintf(`# ddclient configuration for Cloudflare
daemon=%d
syslog=yes
ssl=yes
use=web, web=https://api.ipify.org

protocol=cloudflare
zone=%s
ttl=300
%s=%s`,
			provider.Interval*60, // Convert minutes to seconds
			provider.Domain,
			provider.Username, // email
			provider.Password) // API key

	case "dynu":
		config = fmt.Sprintf(`# ddclient configuration for Dynu
daemon=%d
syslog=yes
ssl=yes
use=web, web=https://api.ipify.org

protocol=dyndns2
server=api.dynu.com
login=%s
password=%s
%s`,
			provider.Interval*60,
			provider.Username,
			provider.Password,
			provider.Domain)

	case "duckdns":
		config = fmt.Sprintf(`# ddclient configuration for DuckDNS
daemon=%d
syslog=yes
ssl=yes
use=web, web=https://api.ipify.org

protocol=duckdns
server=www.duckdns.org
login=nouser
password=%s
%s`,
			provider.Interval*60,
			provider.Password, // token
			provider.Domain)

	default:
		return fmt.Errorf("unsupported DDNS provider: %s", provider.Name)
	}

	// Write config
	if err := os.WriteFile("/etc/ddclient.conf", []byte(config), 0600); err != nil {
		return err
	}

	return nil
}

// saveConfig saves DDNS configuration
func (s *DDNSService) saveConfig(provider *DDNSProvider) error {
	data, err := json.MarshalIndent(provider, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("/etc/vexa/ddns.json", data, 0600)
}

// GetDDNSConfig returns current DDNS configuration
func (s *DDNSService) GetDDNSConfig() (*DDNSProvider, error) {
	data, err := os.ReadFile("/etc/vexa/ddns.json")
	if err != nil {
		if os.IsNotExist(err) {
			return &DDNSProvider{
				Name:     "cloudflare",
				Enabled:  false,
				Interval: 5,
			}, nil
		}
		return nil, err
	}

	var provider DDNSProvider
	if err := json.Unmarshal(data, &provider); err != nil {
		return nil, err
	}

	return &provider, nil
}
