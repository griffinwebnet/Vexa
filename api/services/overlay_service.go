package services

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
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

// JoinScripts contains scripts for different platforms
type JoinScripts struct {
	Windows string `json:"windows"`
	Linux   string `json:"linux"`
}

// SetupOverlay configures Headscale and Tailscale
func (s *OverlayService) SetupOverlay(fqdn string) error {
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

// installHeadscale installs Headscale using the recommended .deb package
func (s *OverlayService) installHeadscale() error {
	// Check if already installed and working
	if _, err := exec.LookPath("headscale"); err == nil {
		// Test if the binary actually works
		if testCmd := exec.Command("headscale", "--help"); testCmd.Run() == nil {
			return nil
		}
		// If it exists but doesn't work, remove it and reinstall
		fmt.Printf("DEBUG: Removing corrupted headscale installation\n")
		exec.Command("dpkg", "--remove", "headscale").Run()
		exec.Command("apt", "autoremove", "-y").Run()
	}

	// Detect architecture
	arch := "amd64"
	if unameM, err := exec.Command("uname", "-m").Output(); err == nil {
		archStr := strings.TrimSpace(string(unameM))
		switch archStr {
		case "x86_64":
			arch = "amd64"
		case "aarch64":
			arch = "arm64"
		default:
			arch = "amd64" // fallback
		}
	}

	fmt.Printf("DEBUG: Detected architecture: %s\n", arch)

	// Get latest version from GitHub API
	fmt.Printf("DEBUG: Getting latest headscale version\n")
	versionCmd := exec.Command("curl", "-s", "https://api.github.com/repos/juanfont/headscale/releases/latest")
	versionOutput, err := versionCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get latest version: %v", err)
	}

	// Parse version from JSON (simple approach)
	versionOutputStr := string(versionOutput)
	versionStart := strings.Index(versionOutputStr, `"tag_name":"v`)
	if versionStart == -1 {
		return fmt.Errorf("could not parse version from GitHub API response")
	}
	versionStart += len(`"tag_name":"v`)
	versionEnd := strings.Index(versionOutputStr[versionStart:], `"`)
	if versionEnd == -1 {
		return fmt.Errorf("could not parse version from GitHub API response")
	}
	version := versionOutputStr[versionStart : versionStart+versionEnd]

	fmt.Printf("DEBUG: Latest headscale version: %s\n", version)

	// Download .deb package
	debURL := fmt.Sprintf("https://github.com/juanfont/headscale/releases/download/v%s/headscale_%s_linux_%s.deb", version, version, arch)
	fmt.Printf("DEBUG: Downloading headscale .deb from: %s\n", debURL)

	downloadCmd := exec.Command("wget", "--output-document=headscale.deb", debURL)
	downloadOutput, err := downloadCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("DEBUG: Download failed: %v, output: %s\n", err, string(downloadOutput))
		return fmt.Errorf("failed to download headscale .deb: %v", err)
	}

	// Install the .deb package
	fmt.Printf("DEBUG: Installing headscale .deb package\n")
	installCmd := exec.Command("dpkg", "-i", "headscale.deb")
	installOutput, err := installCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("DEBUG: Installation failed: %v, output: %s\n", err, string(installOutput))
		// Try to fix dependencies
		exec.Command("apt", "install", "-f", "-y").Run()
		// Try installation again
		installCmd = exec.Command("dpkg", "-i", "headscale.deb")
		_, err = installCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to install headscale .deb: %v", err)
		}
	}

	// Clean up downloaded file
	os.Remove("headscale.deb")

	// Test the installation
	if testCmd := exec.Command("headscale", "--help"); testCmd.Run() != nil {
		return fmt.Errorf("headscale installation failed - binary not working")
	}

	fmt.Printf("DEBUG: Successfully installed headscale version %s\n", version)
	return nil
}

// configureHeadscale sets up Headscale configuration
func (s *OverlayService) configureHeadscale(fqdn string) error {
	// Create config and data directories
	if err := os.MkdirAll("/etc/headscale", 0755); err != nil {
		return err
	}
	if err := os.MkdirAll("/var/lib/headscale", 0755); err != nil {
		return err
	}
	if err := os.MkdirAll("/var/run/headscale", 0755); err != nil {
		return err
	}

	// Extract domain from FQDN
	domain := strings.Split(fqdn, ".")[0]

	// Create config from template
	// Use port 50443 for all Headscale communication (no web UI exposure)
	tmpl := template.Must(template.New("config").Parse(`---
# Vexa Mesh Configuration
server_url: http://{{ .FQDN }}:50443
listen_addr: 0.0.0.0:50443
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
      {{ .Domain }}.local:
        - 100.64.0.1
  search_domains:
    - {{ .Domain }}.local
    - {{ .Domain }}.mesh
    - {{ .Domain }}
  extra_records:
    - name: "{{ .Domain }}.local"
      type: "A"
      value: "100.64.0.1"
    - name: "{{ .Domain }}"
      type: "A"
      value: "100.64.0.1"
    - name: "dc-01.{{ .Domain }}.local"
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

	return nil
}

// startHeadscale starts and enables the Headscale service
func (s *OverlayService) startHeadscale() error {
	// Create systemd service
	serviceFile := `[Unit]
Description=Headscale Controller
After=network.target

[Service]
Type=simple
User=root
Group=root
ExecStart=/usr/local/bin/headscale serve -c /etc/headscale/config.yaml
Restart=always
RestartSec=5
Environment=HOME=/var/lib/headscale

[Install]
WantedBy=multi-user.target`

	if err := os.WriteFile("/etc/systemd/system/headscale.service", []byte(serviceFile), 0644); err != nil {
		return err
	}

	// Reload systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return err
	}

	// Start and enable service
	if err := exec.Command("systemctl", "enable", "--now", "headscale").Run(); err != nil {
		return err
	}

	return nil
}

// installTailscale installs Tailscale client
func (s *OverlayService) installTailscale() error {
	// Check if already installed
	if _, err := exec.LookPath("tailscale"); err == nil {
		return nil
	}

	// Add Tailscale repo and install
	cmd := exec.Command("curl", "-fsSL", "https://pkgs.tailscale.com/stable/ubuntu/jammy.noarmor.gpg", "-o", "/usr/share/keyrings/tailscale-archive-keyring.gpg")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("curl", "-fsSL", "https://pkgs.tailscale.com/stable/ubuntu/jammy.tailscale-keyring.list", "-o", "/etc/apt/sources.list.d/tailscale.list")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("apt", "update")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("apt", "install", "-y", "tailscale")
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// joinMesh joins this server to the Headscale mesh
func (s *OverlayService) joinMesh() error {
	// Generate a pre-auth key for this server
	cmd := exec.Command("headscale", "preauthkey", "create", "--reusable", "--expiration", "8760h")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to create pre-auth key: %v", err)
	}
	authKey := strings.TrimSpace(string(output))

	// Determine login server from headscale config/env if available
	loginServer := os.Getenv("HEADSCALE_SERVER_URL")
	if loginServer == "" {
		if content, err := os.ReadFile("/etc/headscale/config.yaml"); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, raw := range lines {
				line := strings.TrimSpace(raw)
				if strings.HasPrefix(line, "server_url:") {
					val := strings.TrimSpace(strings.TrimPrefix(line, "server_url:"))
					val = strings.Trim(val, "\"'")
					loginServer = val
					break
				}
			}
		}
	}
	if loginServer == "" {
		loginServer = "http://localhost:8080/mesh"
	}

	// Join mesh
	cmd = exec.Command("tailscale", "up",
		"--login-server="+loginServer,
		"--authkey="+authKey,
		"--accept-routes",
		"--accept-dns=false",
		"--hostname=vexa-server")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to join mesh: %v", err)
	}

	return nil
}

// AddMachine generates scripts for joining a new machine
func (s *OverlayService) AddMachine(name string) (*JoinScripts, error) {
	// Generate pre-auth key
	cmd := exec.Command("headscale", "preauthkey", "create", "--reusable", "--expiration", "8760h")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	authKey := strings.TrimSpace(string(output))

	// Get login server URL from config/env
	serverURL := os.Getenv("HEADSCALE_SERVER_URL")
	if serverURL == "" {
		if content, err := os.ReadFile("/etc/headscale/config.yaml"); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, raw := range lines {
				line := strings.TrimSpace(raw)
				if strings.HasPrefix(line, "server_url:") {
					val := strings.TrimSpace(strings.TrimPrefix(line, "server_url:"))
					serverURL = strings.Trim(val, "\"'")
					break
				}
			}
		}
	}
	if serverURL == "" {
		serverURL = "http://localhost:8080/mesh"
	}

	// Generate Windows PowerShell script
	windowsScript := fmt.Sprintf(`# Vexa Domain Join Script (Windows)
# Generated by Vexa for machine: %s

# Pre-auth key for unattended join
$AuthKey = "%s"

# Install Tailscale
$TailscaleURL = "https://pkgs.tailscale.com/stable/tailscale-setup-latest-amd64.msi"
$TailscaleMSI = "$env:TEMP\tailscale.msi"
Invoke-WebRequest -Uri $TailscaleURL -OutFile $TailscaleMSI
Start-Process msiexec.exe -ArgumentList "/i ""$TailscaleMSI"" /quiet" -Wait

# Configure unattended and join network
Start-Process tailscale.exe -ArgumentList "up --authkey=$AuthKey --login-server=%s --unattended --accept-routes" -Wait

# Verify it's set
tailscale debug prefs | Select-String "unattended"`, name, authKey, serverURL)

	// Generate Linux shell script
	linuxScript := fmt.Sprintf(`#!/bin/bash
# Vexa Domain Join Script (Linux)
# Generated by Vexa for machine: %s

# Pre-auth key for unattended join
AUTH_KEY="%s"

# Install Tailscale
if command -v apt-get &> /dev/null; then
    # Debian/Ubuntu
    curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/jammy.noarmor.gpg | sudo tee /usr/share/keyrings/tailscale-archive-keyring.gpg >/dev/null
    curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/jammy.tailscale-keyring.list | sudo tee /etc/apt/sources.list.d/tailscale.list
    sudo apt-get update
    sudo apt-get install -y tailscale
elif command -v dnf &> /dev/null; then
    # RHEL/Rocky
    dnf config-manager --add-repo https://pkgs.tailscale.com/stable/rhel/9/tailscale.repo
    dnf install -y tailscale
elif command -v zypper &> /dev/null; then
    # SUSE
    sudo zypper ar -g -r https://pkgs.tailscale.com/stable/opensuse/leap/15.4 tailscale
    sudo zypper ref
    sudo zypper in tailscale
fi

# Stop any existing Tailscale
sudo systemctl stop tailscaled || true

# Join network
sudo tailscale up \
    --authkey="$AUTH_KEY" \
    --login-server="%s" \
    --accept-routes \
    --unattended

# Enable service
sudo systemctl enable --now tailscaled`, name, authKey, serverURL)

	return &JoinScripts{
		Windows: windowsScript,
		Linux:   linuxScript,
	}, nil
}

// GetOverlayStatus returns the current overlay network status
func (s *OverlayService) GetOverlayStatus() (map[string]interface{}, error) {
	// Check if Headscale is running
	headscaleActive := false
	cmd := exec.Command("systemctl", "is-active", "headscale")
	if err := cmd.Run(); err == nil {
		headscaleActive = true
	}

	// Check if Tailscale is running
	tailscaleActive := false
	cmd = exec.Command("systemctl", "is-active", "tailscaled")
	if err := cmd.Run(); err == nil {
		tailscaleActive = true
	}

	// Get Tailscale status
	var ip, hostname string
	if tailscaleActive {
		cmd = exec.Command("tailscale", "status", "--json")
		output, err := cmd.Output()
		if err == nil {
			status := string(output)
			// Parse IP and hostname (simplified)
			if strings.Contains(status, "100.") {
				parts := strings.Split(strings.Split(status, "100.")[1], "\"")
				if len(parts) > 0 {
					ip = "100." + parts[0]
				}
			}
			if strings.Contains(status, "\"Hostname\":") {
				parts := strings.Split(strings.Split(status, "\"Hostname\":")[1], "\"")
				if len(parts) > 1 {
					hostname = parts[1]
				}
			}
		}
	}

	// Derive configured FQDN from Headscale server_url
	var fqdn string
	serverURL := os.Getenv("HEADSCALE_SERVER_URL")
	if serverURL == "" {
		if content, err := os.ReadFile("/etc/headscale/config.yaml"); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, raw := range lines {
				line := strings.TrimSpace(raw)
				if strings.HasPrefix(line, "server_url:") {
					val := strings.TrimSpace(strings.TrimPrefix(line, "server_url:"))
					serverURL = strings.Trim(val, "\"'")
					break
				}
			}
		}
	}
	if serverURL != "" {
		if u, err := url.Parse(serverURL); err == nil && u.Host != "" {
			fqdn = u.Hostname()
		} else {
			// fallback: strip scheme manually
			s := serverURL
			if strings.Contains(s, "://") {
				parts := strings.SplitN(s, "://", 2)
				s = parts[1]
			}
			fqdn = strings.Split(s, "/")[0]
		}
	}

	return map[string]interface{}{
		"enabled":     headscaleActive && tailscaleActive,
		"headscale":   headscaleActive,
		"tailscale":   tailscaleActive,
		"ip":          ip,
		"hostname":    hostname,
		"fqdn":        fqdn,
		"mesh_domain": "mesh",
	}, nil
}

// TestFQDNAccessibility tests if an FQDN is publicly accessible
func (s *OverlayService) TestFQDNAccessibility(fqdn string) (map[string]interface{}, error) {
	fmt.Printf("DEBUG: Testing FQDN accessibility for: %s\n", fqdn)

	// Test basic DNS resolution first
	dnsCmd := exec.Command("nslookup", fqdn)
	dnsOutput, dnsErr := dnsCmd.CombinedOutput()

	if dnsErr != nil {
		fmt.Printf("DEBUG: DNS resolution failed: %v\n", dnsErr)
		return map[string]interface{}{
			"accessible":  false,
			"reason":      "DNS resolution failed",
			"details":     string(dnsOutput),
			"can_proceed": true,
			"message":     "DNS resolution failed, but you can proceed with setup and configure DNS later",
		}, nil
	}

	fmt.Printf("DEBUG: DNS resolution successful: %s\n", string(dnsOutput))

	// Test HTTP connectivity on port 80
	httpCmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "--connect-timeout", "10", fmt.Sprintf("http://%s", fqdn))
	httpOutput, httpErr := httpCmd.CombinedOutput()
	httpCode := strings.TrimSpace(string(httpOutput))

	fmt.Printf("DEBUG: HTTP test on port 80: code=%s, err=%v\n", httpCode, httpErr)

	// Test HTTPS connectivity on port 443
	httpsCmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "--connect-timeout", "10", "-k", fmt.Sprintf("https://%s", fqdn))
	httpsOutput, httpsErr := httpsCmd.CombinedOutput()
	httpsCode := strings.TrimSpace(string(httpsOutput))

	fmt.Printf("DEBUG: HTTPS test on port 443: code=%s, err=%v\n", httpsCode, httpsErr)

	// Determine accessibility
	accessible := false
	reason := ""
	canProceed := true
	message := ""

	if httpCode == "200" || httpsCode == "200" {
		accessible = true
		reason = "HTTP/HTTPS accessible"
		message = "Great! Your FQDN is publicly accessible and ready for remote connections"
	} else if httpCode == "000" && httpsCode == "000" {
		reason = "No HTTP/HTTPS service responding"
		message = "FQDN resolves but no web service is responding. This is normal for a domain controller - you can proceed with setup"
	} else if strings.Contains(httpCode, "timeout") || strings.Contains(httpsCode, "timeout") {
		reason = "Connection timeout"
		message = "Connection timed out. Check firewall settings and port forwarding. You can still proceed with setup"
	} else {
		reason = fmt.Sprintf("HTTP returned %s, HTTPS returned %s", httpCode, httpsCode)
		message = "FQDN is reachable but returning unexpected responses. You can proceed with setup"
	}

	return map[string]interface{}{
		"accessible":  accessible,
		"reason":      reason,
		"http_code":   httpCode,
		"https_code":  httpsCode,
		"can_proceed": canProceed,
		"message":     message,
		"dns_output":  string(dnsOutput),
	}, nil
}
