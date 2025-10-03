package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"text/template"
)

// DDNSProvider represents a DDNS provider configuration
type DDNSProvider struct {
	Name     string `json:"name"`     // Provider name (cloudflare, dynu)
	Enabled  bool   `json:"enabled"`  // Whether DDNS is enabled
	Domain   string `json:"domain"`   // Domain to update
	Username string `json:"username"` // API username/token
	Password string `json:"password"` // API password/key
	Interval int    `json:"interval"` // Update interval in minutes
}

// DDNSService handles dynamic DNS updates
type DDNSService struct{}

// NewDDNSService creates a new DDNSService
func NewDDNSService() *DDNSService {
	return &DDNSService{}
}

// SetupDDNS configures and starts DDNS updates
func (s *DDNSService) SetupDDNS(provider *DDNSProvider) error {
	// Save config
	if err := s.saveConfig(provider); err != nil {
		return err
	}

	// Configure ddclient
	if err := s.configureDDClient(provider); err != nil {
		return err
	}

	// Start/stop service based on enabled flag
	if provider.Enabled {
		if err := exec.Command("systemctl", "enable", "--now", "ddclient").Run(); err != nil {
			return fmt.Errorf("failed to start DDNS service: %v", err)
		}
	} else {
		if err := exec.Command("systemctl", "disable", "--now", "ddclient").Run(); err != nil {
			return fmt.Errorf("failed to stop DDNS service: %v", err)
		}
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

// configureDDClient sets up ddclient configuration
func (s *DDNSService) configureDDClient(provider *DDNSProvider) error {
	// Create config from template
	tmpl := template.Must(template.New("config").Parse(`# ddclient configuration for Vexa
daemon={{ .Interval }}
use=web, web=checkip.dyndns.org/, web-skip='IP Address'
ssl=yes

{{ if eq .Name "cloudflare" }}
protocol=cloudflare
zone={{ .Domain }}
login={{ .Username }}
password={{ .Password }}
{{ .Domain }}

{{ else if eq .Name "dynu" }}
protocol=dyndns2
server=api.dynu.com
login={{ .Username }}
password={{ .Password }}
{{ .Domain }}

{{ else if eq .Name "duckdns" }}
protocol=duckdns
server=www.duckdns.org
login={{ .Username }}
password={{ .Password }}
{{ .Domain }}

{{ else if eq .Name "noip" }}
protocol=noip
server=dynupdate.no-ip.com
login={{ .Username }}
password={{ .Password }}
{{ .Domain }}

{{ end }}`))

	f, err := os.Create("/etc/ddclient.conf")
	if err != nil {
		return err
	}
	defer f.Close()

	if err := tmpl.Execute(f, provider); err != nil {
		return err
	}

	// Set permissions
	if err := os.Chmod("/etc/ddclient.conf", 0600); err != nil {
		return err
	}

	return nil
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
