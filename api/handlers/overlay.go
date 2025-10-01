package handlers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type OverlayStatus struct {
	Enabled    bool   `json:"enabled"`
	FQDN       string `json:"fqdn,omitempty"`
	MeshDomain string `json:"mesh_domain,omitempty"`
	PreAuthKey string `json:"pre_auth_key,omitempty"`
}

type OverlaySetupRequest struct {
	FQDN string `json:"fqdn" binding:"required"`
}

// GetOverlayStatus returns the current overlay networking status
func GetOverlayStatus(c *gin.Context) {
	// Check if headscale is installed and running
	cmd := exec.Command("systemctl", "is-active", "headscale")
	err := cmd.Run()

	if err != nil {
		// Not enabled
		c.JSON(http.StatusOK, OverlayStatus{
			Enabled: false,
		})
		return
	}

	// Read config to get FQDN and mesh domain
	// TODO: Parse /etc/headscale/config.yaml
	c.JSON(http.StatusOK, OverlayStatus{
		Enabled:    true,
		FQDN:       "example.com", // TODO: Read from config
		MeshDomain: "example.mesh",
	})
}

// SetupOverlay installs and configures Headscale
func SetupOverlay(c *gin.Context) {
	var req OverlaySetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Extract domain parts
	domainParts := strings.Split(req.FQDN, ".")
	meshDomain := domainParts[0] + ".mesh"

	// Install Headscale
	if err := installHeadscale(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to install Headscale",
			"details": err.Error(),
		})
		return
	}

	// Generate config
	configPath := "/etc/headscale/config.yaml"
	config := generateHeadscaleConfig(req.FQDN, meshDomain)

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create config directory",
		})
		return
	}

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to write config file",
		})
		return
	}

	// Start Headscale
	cmd := exec.Command("systemctl", "enable", "--now", "headscale")
	if err := cmd.Run(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start Headscale",
		})
		return
	}

	// Create infra user
	cmd = exec.Command("headscale", "users", "create", "infra")
	if err := cmd.Run(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create infra user",
		})
		return
	}

	// Generate pre-auth key (15 years = 131400 hours)
	cmd = exec.Command("headscale", "preauthkeys", "create", "--user", "infra", "--expiration", "131400h", "--reusable")
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create pre-auth key",
		})
		return
	}

	preAuthKey := strings.TrimSpace(string(output))

	// Install Tailscale client
	if err := installTailscale(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to install Tailscale client",
		})
		return
	}

	// Connect to Headscale
	cmd = exec.Command("tailscale", "up",
		"--login-server", fmt.Sprintf("http://localhost:8080"),
		"--authkey", preAuthKey,
		"--hostname", "vexa-dc",
	)
	if err := cmd.Run(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect Tailscale client",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Overlay networking configured successfully",
		"fqdn":         req.FQDN,
		"mesh_domain":  meshDomain,
		"pre_auth_key": preAuthKey,
		"internal_ip":  "100.64.0.1",
	})
}

func installHeadscale() error {
	// Detect OS and install accordingly
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		// Debian/Ubuntu
		cmd := exec.Command("bash", "-c", `
			wget -O /tmp/headscale.deb https://github.com/juanfont/headscale/releases/latest/download/headscale_*_linux_amd64.deb
			dpkg -i /tmp/headscale.deb
		`)
		return cmd.Run()
	} else if _, err := os.Stat("/etc/redhat-release"); err == nil {
		// RHEL/CentOS
		cmd := exec.Command("bash", "-c", `
			wget -O /tmp/headscale.rpm https://github.com/juanfont/headscale/releases/latest/download/headscale_*_linux_amd64.rpm
			rpm -i /tmp/headscale.rpm
		`)
		return cmd.Run()
	}

	return fmt.Errorf("unsupported OS")
}

func installTailscale() error {
	// Install Tailscale using official script
	cmd := exec.Command("bash", "-c", "curl -fsSL https://tailscale.com/install.sh | sh")
	return cmd.Run()
}

func generateHeadscaleConfig(fqdn, meshDomain string) string {
	// Get AD domain from samba
	// TODO: Parse from samba config
	adDomain := "example.local"

	return fmt.Sprintf(`---
server_url: http://%s:8080
listen_addr: 0.0.0.0:8080
metrics_listen_addr: 127.0.0.1:9090
grpc_listen_addr: 127.0.0.1:50443
grpc_allow_insecure: false

private_key_path: /var/lib/headscale/private.key
noise:
  private_key_path: /var/lib/headscale/noise_private.key

ip_prefixes:
  - fd7a:115c:a1e0::/48
  - 100.64.0.0/10

derp:
  server:
    enabled: false
  urls:
    - https://controlplane.tailscale.com/derpmap/default
  paths: []
  auto_update_enabled: true
  update_frequency: 24h

disable_check_updates: false
ephemeral_node_inactivity_timeout: 30m

database:
  type: sqlite3
  sqlite:
    path: /var/lib/headscale/db.sqlite

acme_url: https://acme-v02.api.letsencrypt.org/directory
acme_email: ""
tls_letsencrypt_hostname: ""
tls_letsencrypt_cache_dir: /var/lib/headscale/cache
tls_letsencrypt_challenge_type: HTTP-01
tls_letsencrypt_listen: ":http"
tls_cert_path: ""
tls_key_path: ""

log:
  format: text
  level: info

dns_config:
  override_local_dns: true
  nameservers:
    - 100.64.0.1
  domains:
    - %s
  magic_dns: true
  base_domain: %s
  split_dns:
    %s:
      - 100.64.0.1

unix_socket: /var/run/headscale/headscale.sock
unix_socket_permission: "0770"
`, fqdn, meshDomain, adDomain)
}
