package models

// Computer represents a domain computer/device
type Computer struct {
	Name            string `json:"name"`
	DNSName         string `json:"dns_name"`
	OperatingSystem string `json:"operating_system"`
	LastLogon       string `json:"last_logon"`
	ConnectionType  string `json:"connection_type"` // "local", "overlay", "offline"
	Online          bool   `json:"online"`
	IPAddress       string `json:"ip_address,omitempty"`
	OverlayIP       string `json:"overlay_ip,omitempty"`
	OverlayURL      string `json:"overlay_url,omitempty"`
}
