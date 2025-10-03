package services

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

// OverlayService handles Headscale/Tailscale overlay networking
type OverlayService struct {
	ddnsService *DDNSService
}

// NewOverlayService creates a new OverlayService
func NewOverlayService() *OverlayService {
	return &OverlayService{
		ddnsService: NewDDNSService(),
	}
}

// SetupOverlay configures Headscale and Tailscale
func (s *OverlayService) SetupOverlay(fqdn string, ddns *DDNSProvider) error {
	// Configure DDNS if enabled
	if ddns != nil && ddns.Enabled {
		if err := s.ddnsService.SetupDDNS(ddns); err != nil {
			return fmt.Errorf("failed to setup DDNS: %v", err)
		}
	}

	// Install Headscale
	if err := s.installHeadscale(); err != nil {
		return fmt.Errorf("failed to install headscale: %v", err)
	}

	// Configure Headscale
	if err := s.configureHeadscale(fqdn); err != nil {
		return fmt.Errorf("failed to configure headscale: %v", err)
	}

	// Start Headscale
	if err := s.startHeadscale(); err != nil {
		return fmt.Errorf("failed to start headscale: %v", err)
	}

	// Install Tailscale
	if err := s.installTailscale(); err != nil {
		return fmt.Errorf("failed to install tailscale: %v", err)
	}

	// Join this server to the mesh
	if err := s.joinMesh(); err != nil {
		return fmt.Errorf("failed to join mesh: %v", err)
	}

	return nil
}

// configureHeadscale sets up Headscale configuration
func (s *OverlayService) configureHeadscale(fqdn string) error {
	// Create config directory
	if err := os.MkdirAll("/etc/headscale", 0755); err != nil {
		return err
	}

	// Create config from template
	tmpl := template.Must(template.New("config").Parse(`---
# Vexa Mesh Configuration
server_url: https://{{ .FQDN }}
listen_addr: 0.0.0.0:8080
metrics_listen_addr: 127.0.0.1:9090
grpc_listen_addr: 0.0.0.0:50443
grpc_allow_insecure: true

noise:
  private_key_path: /var/lib/headscale/noise_private.key

prefixes:
  v4: 100.64.0.0/10
  v6: fd7a:115c:a1e0::/48
  allocation: sequential

derp:
  server:
    enabled: false
  urls:
    - https://controlplane.tailscale.com/derpmap/default
  auto_update_enabled: true
  update_frequency: 24h

disable_check_updates: false
ephemeral_node_inactivity_timeout: 30m

database:
  type: sqlite
  debug: false
  gorm:
    prepare_stmt: true
    parameterized_queries: true
    skip_err_record_not_found: true
    slow_threshold: 1000
  sqlite:
    path: /var/lib/headscale/db.sqlite
    write_ahead_log: true
    wal_autocheckpoint: 1000

log:
  format: text
  level: info

policy:
  mode: file
  path: ""

# Windows AD Integration
enable_windows_networking: true
allow_netbios_broadcast: true

dns:
  magic_dns: true
  base_domain: {{ .Domain }}.mesh
  override_local_dns: true
  nameservers:
    global:
      - 100.64.0.1
    split:
      {{ .Domain }}.internal:
        - 100.64.0.1
  search_domains: 
    - {{ .Domain }}.internal
    - {{ .Domain }}.mesh
    - {{ .Domain }}
  extra_records:
    - name: "{{ .Domain }}.internal"
      type: "A"
      value: "100.64.0.1"
    - name: "{{ .Domain }}"
      type: "A"
      value: "100.64.0.1"
    - name: "dc-01.{{ .Domain }}.internal"
      type: "A"
      value: "100.64.0.1"

unix_socket: /var/run/headscale/headscale.sock
unix_socket_permission: "0770"
logtail:
  enabled: false
randomize_client_port: false`))

	f, err := os.Create("/etc/headscale/config.yaml")
	if err != nil {
		return err
	}
	defer f.Close()

	// Extract domain from FQDN
	domain := strings.Split(fqdn, ".")[0]

	data := struct {
		FQDN   string
		Domain string
	}{
		FQDN:   fqdn,
		Domain: domain,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	// Create empty ACL file
	aclConfig := `groups:
  admin:
    - admin
  users:
    - "*"

acls:
  - action: accept
    users: ["*"]
    ports: ["*:*"]`

	if err := os.WriteFile("/etc/headscale/acl.yaml", []byte(aclConfig), 0644); err != nil {
		return err
	}

	return nil
}

// Rest of the service implementation...`
