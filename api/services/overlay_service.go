package services

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
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

	// Join this server to the mesh via localhost (not dependent on external connectivity)
	if err := s.joinMeshLocal(fqdn); err != nil {
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

// joinMeshLocal joins the mesh using localhost connection (not dependent on external connectivity)
func (s *OverlayService) joinMeshLocal(fqdn string) error {
	fmt.Printf("DEBUG: Joining mesh via localhost connection\n")

	// Generate a pre-auth key for this server
	cmd := exec.Command("headscale", "preauthkey", "create", "--reusable", "--expiration", "8760h")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to create pre-auth key: %v", err)
	}
	authKey := strings.TrimSpace(string(output))

	// Use localhost for the login server to avoid external connectivity dependency
	loginServer := "http://127.0.0.1:50443"

	fmt.Printf("DEBUG: Using localhost login server: %s\n", loginServer)

	// Join mesh using localhost
	joinCmd := exec.Command("tailscale", "up", "--authkey", authKey, "--login-server", loginServer, "--accept-routes")
	joinOutput, joinErr := joinCmd.CombinedOutput()
	if joinErr != nil {
		return fmt.Errorf("failed to join mesh via localhost: %v, output: %s", joinErr, string(joinOutput))
	}

	fmt.Printf("DEBUG: Successfully joined mesh via localhost\n")
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

// TestFQDNWithListener tests FQDN accessibility by setting up a temporary listener
func (s *OverlayService) TestFQDNWithListener(fqdn string) (map[string]interface{}, error) {
	fmt.Printf("DEBUG: Testing FQDN with temporary listener: %s\n", fqdn)

	// First check DNS resolution
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

	fmt.Printf("DEBUG: DNS resolution successful\n")

	// Check if port 50443 is already in use
	fmt.Printf("DEBUG: Checking if port 50443 is available\n")
	checkCmd := exec.Command("netstat", "-tlnp")
	checkOutput, checkErr := checkCmd.Output()
	if checkErr == nil {
		outputStr := string(checkOutput)
		if strings.Contains(outputStr, ":50443") {
			fmt.Printf("DEBUG: Port 50443 is already in use\n")
			return map[string]interface{}{
				"accessible":  true,
				"reason":      "Port 50443 already in use",
				"message":     "Port 50443 is already in use. This may indicate Headscale is already running or another service is using this port.",
				"can_proceed": true,
			}, nil
		}
	}

	// Start a temporary HTTP server on port 50443
	fmt.Printf("DEBUG: Starting temporary listener on port 50443\n")

	// Create a simple HTTP server that responds with a test message
	// Explicitly bind to all interfaces (0.0.0.0) to ensure it's accessible from other devices
	listener, err := net.Listen("tcp", "0.0.0.0:50443")
	if err != nil {
		fmt.Printf("DEBUG: Failed to bind to port 50443: %v\n", err)

		// Check if it's a permission issue
		if strings.Contains(err.Error(), "permission denied") {
			return map[string]interface{}{
				"accessible":  false,
				"reason":      "Permission denied",
				"message":     "Cannot bind to port 50443 - permission denied. Make sure the API is running with sufficient privileges.",
				"can_proceed": true,
			}, nil
		}

		// Check if port is already in use
		if strings.Contains(err.Error(), "address already in use") {
			return map[string]interface{}{
				"accessible":  true,
				"reason":      "Port already in use",
				"message":     "Port 50443 is already in use. This may indicate Headscale is already running.",
				"can_proceed": true,
			}, nil
		}

		return map[string]interface{}{
			"accessible":  false,
			"reason":      "Port binding failed",
			"message":     fmt.Sprintf("Could not bind to port 50443: %v", err),
			"can_proceed": true,
		}, nil
	}

	fmt.Printf("DEBUG: Successfully bound to port 50443\n")
	const testPort = 50443

	// Start the server in a goroutine
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("DEBUG: Received request: %s %s\n", r.Method, r.URL.Path)
			// Handle different paths that Headscale might use
			if r.URL.Path == "/" || r.URL.Path == "/api/v1" || r.URL.Path == "/mesh" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"test": "vexa-fqdn-test", "status": "success"}`))
				fmt.Printf("DEBUG: Sent 200 response\n")
			} else {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Not found"))
				fmt.Printf("DEBUG: Sent 404 response\n")
			}
		}),
	}

	go func() {
		fmt.Printf("DEBUG: Starting HTTP server on port 50443\n")
		if err := server.Serve(listener); err != nil {
			fmt.Printf("DEBUG: Server error: %v\n", err)
		}
	}()

	// Give the server a moment to start
	fmt.Printf("DEBUG: Waiting for server to start...\n")
	time.Sleep(2 * time.Second)

	// Test if the server is actually listening
	fmt.Printf("DEBUG: Testing if server is listening on port %d...\n", testPort)
	testListener, testErr := net.Listen("tcp", fmt.Sprintf(":%d", testPort))
	if testErr != nil {
		fmt.Printf("DEBUG: Port %d is in use (good!): %v\n", testPort, testErr)
	} else {
		fmt.Printf("DEBUG: Port %d is still available (bad!) - server didn't start properly\n", testPort)
		testListener.Close()
	}

	// First test local connectivity to make sure our listener is working
	fmt.Printf("DEBUG: Testing local listener on port %d\n", testPort)
	localTestCmd := exec.Command("curl", "-v", "--connect-timeout", "5", fmt.Sprintf("http://127.0.0.1:%d", testPort))
	localOutput, localErr := localTestCmd.CombinedOutput()
	localCode := "000" // Default to failure

	// Parse the output to get HTTP status code
	outputStr := string(localOutput)
	fmt.Printf("DEBUG: Local curl output: %s\n", outputStr)
	fmt.Printf("DEBUG: Local curl error: %v\n", localErr)

	// Look for HTTP status in the verbose output
	if strings.Contains(outputStr, "HTTP/1.1 200") || strings.Contains(outputStr, "HTTP/2 200") {
		localCode = "200"
	} else if strings.Contains(outputStr, "Connection refused") {
		localCode = "000"
	} else if strings.Contains(outputStr, "timeout") {
		localCode = "timeout"
	}

	fmt.Printf("DEBUG: Local test result: code=%s\n", localCode)

	// Also test using the server's LAN IP address
	// Get the server's LAN IP address
	lanIPCmd := exec.Command("hostname", "-I")
	lanIPOutput, _ := lanIPCmd.Output()
	lanIP := strings.TrimSpace(strings.Split(string(lanIPOutput), " ")[0])

	if lanIP != "" {
		fmt.Printf("DEBUG: Testing with LAN IP %s on port %d\n", lanIP, testPort)
		lanTestCmd := exec.Command("curl", "-v", "--connect-timeout", "5", fmt.Sprintf("http://%s:%d", lanIP, testPort))
		lanOutput, lanErr := lanTestCmd.CombinedOutput()

		fmt.Printf("DEBUG: LAN IP curl output: %s\n", string(lanOutput))
		fmt.Printf("DEBUG: LAN IP curl error: %v\n", lanErr)

		if strings.Contains(string(lanOutput), "HTTP/1.1 200") || strings.Contains(string(lanOutput), "HTTP/2 200") {
			fmt.Printf("DEBUG: LAN IP test successful!\n")
		} else {
			fmt.Printf("DEBUG: LAN IP test failed. This may indicate a firewall or binding issue.\n")
		}
	}

	// Now test external FQDN accessibility
	fmt.Printf("DEBUG: Testing external FQDN accessibility via port 50443\n")
	testCmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "--connect-timeout", "10", fmt.Sprintf("http://%s:50443", fqdn))
	testOutput, testErr := testCmd.CombinedOutput()
	testCode := strings.TrimSpace(string(testOutput))

	// Stop the temporary server
	listener.Close()
	server.Close()

	fmt.Printf("DEBUG: External FQDN test result: code=%s, err=%v\n", testCode, testErr)

	// Analyze results
	accessible := false
	reason := ""
	message := ""
	canProceed := true

	// Check if local listener is working
	if localCode != "200" {
		reason = "Local listener failed"
		message = "Failed to start test listener. Port 50443 may be in use or system error occurred."
		canProceed = false
	} else if testCode == "200" {
		// External test succeeded
		accessible = true
		reason = "FQDN is publicly accessible"
		message = "Perfect! Your FQDN is accessible from the internet. Remote users will be able to connect."
	} else if testCode == "000" || strings.Contains(testCode, "timeout") {
		// External test failed - likely DNS or firewall issue
		reason = "FQDN not accessible from internet"
		message = "Cannot reach your FQDN from the internet. Check DNS settings and port forwarding."
		canProceed = false
	} else {
		// External test got some response but not 200
		accessible = true
		reason = fmt.Sprintf("FQDN accessible (HTTP %s)", testCode)
		message = "Your FQDN is reachable, though returning an unexpected response. You can proceed with setup."
	}

	return map[string]interface{}{
		"accessible":    accessible,
		"reason":        reason,
		"local_code":    localCode,
		"external_code": testCode,
		"can_proceed":   canProceed,
		"message":       message,
		"dns_output":    string(dnsOutput),
	}, nil
}

// TestFQDNAccessibility tests if an FQDN is publicly accessible (legacy method)
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

	// Test Tailscale port 50443 connectivity
	tailscaleCmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "--connect-timeout", "10", fmt.Sprintf("http://%s:50443", fqdn))
	tailscaleOutput, tailscaleErr := tailscaleCmd.CombinedOutput()
	tailscaleCode := strings.TrimSpace(string(tailscaleOutput))

	fmt.Printf("DEBUG: Tailscale port 50443 test: code=%s, err=%v\n", tailscaleCode, tailscaleErr)

	// Also test basic HTTP connectivity on port 80 as secondary check
	httpCmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "--connect-timeout", "10", fmt.Sprintf("http://%s", fqdn))
	httpOutput, httpErr := httpCmd.CombinedOutput()
	httpCode := strings.TrimSpace(string(httpOutput))

	fmt.Printf("DEBUG: HTTP test on port 80: code=%s, err=%v\n", httpCode, httpErr)

	// Determine accessibility
	accessible := false
	reason := ""
	canProceed := true
	message := ""

	// Primary test: Check Tailscale port 50443
	if tailscaleCode == "000" || strings.Contains(tailscaleCode, "timeout") {
		// Port 50443 not accessible - this is expected before setup
		// Fall back to HTTP port 80 test for basic connectivity
		isHttpGood := httpCode == "200" || httpCode == "301" || httpCode == "302" || httpCode == "404"

		if isHttpGood {
			accessible = true
			reason = "FQDN accessible (port 50443 not yet configured - normal before setup)"
			message = "Great! Your FQDN is reachable. Port 50443 will be configured during setup. You can proceed with overlay networking setup."
		} else if httpCode == "000" || strings.Contains(httpCode, "timeout") {
			reason = "FQDN may not be publicly accessible"
			message = "Cannot reach your FQDN. Check DNS settings and ensure the domain points to this server's public IP."
			canProceed = false
		} else {
			accessible = true
			reason = fmt.Sprintf("FQDN accessible (HTTP %s)", httpCode)
			message = "Your FQDN is reachable. You can proceed with overlay networking setup."
		}
	} else {
		// Port 50443 is responding - either already configured or something else is running
		accessible = true
		reason = fmt.Sprintf("Port 50443 is responding (HTTP %s)", tailscaleCode)
		message = "Port 50443 is already responding. This may indicate Tailscale is already configured or another service is using this port."
	}

	return map[string]interface{}{
		"accessible":     accessible,
		"reason":         reason,
		"tailscale_code": tailscaleCode,
		"http_code":      httpCode,
		"can_proceed":    canProceed,
		"message":        message,
		"dns_output":     string(dnsOutput),
	}, nil
}
