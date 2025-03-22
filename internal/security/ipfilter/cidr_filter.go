package ipfilter

import (
	"net"
	"net/http"
	"strings"
)

// CIDRFilter implements IP filtering based on IP addresses and CIDR ranges
type CIDRFilter struct {
	config        *IPFilterConfig
	ipNets        []*net.IPNet
	individualIPs map[string]bool
}

// NewCIDRFilter creates a new CIDR-based IP filter
func NewCIDRFilter(config *IPFilterConfig) (*CIDRFilter, error) {
	filter := &CIDRFilter{
		config:        config,
		ipNets:        make([]*net.IPNet, 0),
		individualIPs: make(map[string]bool),
	}

	// Parse IP addresses and CIDR ranges
	for _, ipStr := range config.IPs {
		// Check if it's a CIDR range
		if strings.Contains(ipStr, "/") {
			_, ipNet, err := net.ParseCIDR(ipStr)
			if err != nil {
				return nil, err
			}
			filter.ipNets = append(filter.ipNets, ipNet)
		} else {
			// Individual IP address
			ip := net.ParseIP(ipStr)
			if ip == nil {
				return nil, &net.ParseError{Type: "IP address", Text: ipStr}
			}
			filter.individualIPs[ip.String()] = true
		}
	}

	return filter, nil
}

// IsAllowed checks if an IP address should be allowed
func (f *CIDRFilter) IsAllowed(ip net.IP) bool {
	// Check individual IPs
	if f.individualIPs[ip.String()] {
		return f.config.Mode == IPFilterModeAllow
	}

	// Check CIDR ranges
	for _, ipNet := range f.ipNets {
		if ipNet.Contains(ip) {
			return f.config.Mode == IPFilterModeAllow
		}
	}

	// If no match, do the opposite of the mode
	return f.config.Mode == IPFilterModeDeny
}

// IsRequestAllowed checks if a request should be allowed
func (f *CIDRFilter) IsRequestAllowed(req *http.Request) bool {
	// Extract client IP
	ipStr := req.RemoteAddr

	// Check for X-Forwarded-For header
	if forwardedFor := req.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// Use the leftmost IP (client IP)
		ips := strings.Split(forwardedFor, ",")
		ipStr = strings.TrimSpace(ips[0])
	}

	// Extract IP from host:port format
	host, _, err := net.SplitHostPort(ipStr)
	if err == nil {
		ipStr = host
	}

	// Parse IP
	ip := net.ParseIP(ipStr)
	if ip == nil {
		// Can't parse IP, default to deny for safety
		return false
	}

	return f.IsAllowed(ip)
}

// GetConfig returns the IP filter configuration
func (f *CIDRFilter) GetConfig() *IPFilterConfig {
	return f.config
}

// GetCountryForIP returns an empty string (not implemented for CIDR filter)
func (f *CIDRFilter) GetCountryForIP(ip net.IP) string {
	return ""
}
