package ipfilter

import (
	"net"
	"net/http"
)

// IPFilterMode defines the mode of IP filtering
type IPFilterMode string

const (
	// IPFilterModeAllow allows listed IPs (denies all others)
	IPFilterModeAllow IPFilterMode = "allow"

	// IPFilterModeDeny denies listed IPs (allows all others)
	IPFilterModeDeny IPFilterMode = "deny"
)

// IPFilterConfig defines IP filtering configuration
type IPFilterConfig struct {
	// Enabled indicates whether IP filtering is enabled
	Enabled bool `yaml:"enabled"`

	// Mode is the filtering mode (allow or deny)
	Mode IPFilterMode `yaml:"mode" default:"deny"`

	// IPs is a list of IP addresses or CIDR ranges
	IPs []string `yaml:"ips,omitempty"`

	// Countries is a list of country codes to filter
	Countries []string `yaml:"countries,omitempty"`

	// Regions is a list of region codes to filter
	Regions []string `yaml:"regions,omitempty"`

	// GeoDBPath is the path to the GeoIP database file
	GeoDBPath string `yaml:"geo_db_path,omitempty"`

	// ThreatDBPath is the path to the threat intelligence database
	ThreatDBPath string `yaml:"threat_db_path,omitempty"`

	// BlockThreatIPs indicates whether to block IPs from threat intelligence
	BlockThreatIPs bool `yaml:"block_threat_ips" default:"false"`

	// LogOnly indicates whether to only log violations without blocking
	LogOnly bool `yaml:"log_only" default:"false"`

	// ResponseStatusCode is the HTTP status code to return when blocked
	ResponseStatusCode int `yaml:"response_status_code" default:"403"`

	// ResponseBody is the body to return when blocked
	ResponseBody string `yaml:"response_body" default:"Access denied by IP filter"`
}

// IPFilter interface defines IP filtering functionality
type IPFilter interface {
	// IsAllowed checks if an IP address should be allowed
	IsAllowed(ip net.IP) bool

	// IsRequestAllowed checks if a request should be allowed
	IsRequestAllowed(req *http.Request) bool

	// GetConfig returns the IP filter configuration
	GetConfig() *IPFilterConfig

	// GetCountryForIP returns the country code for an IP address
	GetCountryForIP(ip net.IP) string
}
