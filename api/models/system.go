package models

// SystemStatusResponse represents system status information
type SystemStatusResponse struct {
	OS             string `json:"os"`
	Architecture   string `json:"architecture"`
	SambaInstalled bool   `json:"samba_installed"`
	BindInstalled  bool   `json:"bind_installed"`
	APIVersion     string `json:"api_version"`
}

// ServiceStatus represents the status of a system service
type ServiceStatus struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
	Status string `json:"status"`
}
