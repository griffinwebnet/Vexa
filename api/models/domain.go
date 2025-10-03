package models

// ProvisionDomainRequest represents the request to provision a new domain
type ProvisionDomainRequest struct {
	Domain       string `json:"domain" binding:"required"`
	Realm        string `json:"realm" binding:"required"`
	DNSBackend   string `json:"dns_backend"`
	DNSForwarder string `json:"dns_forwarder"`
}

// DomainStatusResponse represents the current status of the domain
type DomainStatusResponse struct {
	Provisioned bool   `json:"provisioned"`
	Domain      string `json:"domain,omitempty"`
	Realm       string `json:"realm,omitempty"`
	DCReady     bool   `json:"dc_ready"`
	DNSReady    bool   `json:"dns_ready"`
}

// DomainInfo represents detailed domain information
type DomainInfo struct {
	Domain     string `json:"domain"`
	Realm      string `json:"realm"`
	ForestMode string `json:"forest_mode"`
	DomainMode string `json:"domain_mode"`
	DCName     string `json:"dc_name"`
	DNSServer  string `json:"dns_server"`
	LDAPServer string `json:"ldap_server"`
}
