package ipfilter

import (
	"bufio"
	"net"
	"net/http"
	"os"
	"strings"
)

// ThreatFilter implements IP filtering based on threat intelligence
type ThreatFilter struct {
	baseFilter  IPFilter
	threatIPs   map[string]bool
	threatCIDRs []*net.IPNet
}

// NewThreatFilter creates a new threat intelligence-based IP filter
func NewThreatFilter(baseFilter IPFilter, threatDBPath string) (*ThreatFilter, error) {
	filter := &ThreatFilter{
		baseFilter:  baseFilter,
		threatIPs:   make(map[string]bool),
		threatCIDRs: make([]*net.IPNet, 0),
	}

	// Load threat database if provided
	if threatDBPath != "" {
		if err := filter.loadThreatDB(threatDBPath); err != nil {
			return nil, err
		}
	}

	return filter, nil
}

// loadThreatDB loads IPs and CIDR ranges from a threat intelligence database
func (f *ThreatFilter) loadThreatDB(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check if it's a CIDR range
		if strings.Contains(line, "/") {
			_, ipNet, err := net.ParseCIDR(line)
			if err == nil {
				f.threatCIDRs = append(f.threatCIDRs, ipNet)
			}
		} else {
			// Individual IP address
			ip := net.ParseIP(line)
			if ip != nil {
				f.threatIPs[ip.String()] = true
			}
		}
	}

	return scanner.Err()
}

// IsAllowed checks if an IP address should be allowed
func (f *ThreatFilter) IsAllowed(ip net.IP) bool {
	// First check base filter
	if !f.baseFilter.IsAllowed(ip) {
		return false
	}

	// If not configured to block threat IPs, allow
	if !f.baseFilter.GetConfig().BlockThreatIPs {
		return true
	}

	// Check if IP is in threat list
	if f.threatIPs[ip.String()] {
		return false
	}

	// Check CIDR ranges
	for _, ipNet := range f.threatCIDRs {
		if ipNet.Contains(ip) {
			return false
		}
	}

	// IP is not in threat list
	return true
}

// IsRequestAllowed checks if a request should be allowed
func (f *ThreatFilter) IsRequestAllowed(req *http.Request) bool {
	return f.baseFilter.IsRequestAllowed(req)
}

// GetConfig returns the IP filter configuration
func (f *ThreatFilter) GetConfig() *IPFilterConfig {
	return f.baseFilter.GetConfig()
}

// GetCountryForIP returns the country code for an IP address
func (f *ThreatFilter) GetCountryForIP(ip net.IP) string {
	return f.baseFilter.GetCountryForIP(ip)
}
