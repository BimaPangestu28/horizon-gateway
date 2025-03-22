package ipfilter

import (
	"net"
	"net/http"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

// GeoFilter implements IP filtering based on geolocation
type GeoFilter struct {
	config       *IPFilterConfig
	geoDB        *geoip2.Reader
	countriesMap map[string]bool
	regionsMap   map[string]bool
	cidrFilter   *CIDRFilter // For IP/CIDR filtering
}

// NewGeoFilter creates a new geolocation-based IP filter
func NewGeoFilter(config *IPFilterConfig) (*GeoFilter, error) {
	filter := &GeoFilter{
		config:       config,
		countriesMap: make(map[string]bool),
		regionsMap:   make(map[string]bool),
	}

	// Initialize CIDR filter for IP/CIDR filtering
	cidrFilter, err := NewCIDRFilter(config)
	if err != nil {
		return nil, err
	}
	filter.cidrFilter = cidrFilter

	// Open GeoIP database
	if config.GeoDBPath != "" {
		db, err := geoip2.Open(config.GeoDBPath)
		if err != nil {
			return nil, err
		}
		filter.geoDB = db
	}

	// Build country map
	for _, country := range config.Countries {
		filter.countriesMap[strings.ToUpper(country)] = true
	}

	// Build region map
	for _, region := range config.Regions {
		filter.regionsMap[strings.ToUpper(region)] = true
	}

	return filter, nil
}

// IsAllowed checks if an IP address should be allowed
func (f *GeoFilter) IsAllowed(ip net.IP) bool {
	// First check IP/CIDR rules
	if !f.cidrFilter.IsAllowed(ip) {
		return false
	}

	// If no geo database or no country/region lists, allow by default
	if f.geoDB == nil || (len(f.config.Countries) == 0 && len(f.config.Regions) == 0) {
		return true
	}

	// Look up country
	country := f.GetCountryForIP(ip)
	if country == "" {
		// Unknown country, default to allow
		return true
	}

	// Check country list
	countryMatch := f.countriesMap[country]

	// Get region if available
	var regionMatch bool
	if f.geoDB != nil && len(f.config.Regions) > 0 {
		record, err := f.geoDB.City(ip)
		if err == nil && record.Subdivisions != nil && len(record.Subdivisions) > 0 {
			region := strings.ToUpper(record.Subdivisions[0].IsoCode)
			regionMatch = f.regionsMap[region]
		}
	}

	// In allow mode, we need a match to allow
	if f.config.Mode == IPFilterModeAllow {
		return countryMatch || regionMatch
	}

	// In deny mode, a match means deny
	return !(countryMatch || regionMatch)
}

// IsRequestAllowed checks if a request should be allowed
func (f *GeoFilter) IsRequestAllowed(req *http.Request) bool {
	// Extract client IP (same as in CIDRFilter)
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
func (f *GeoFilter) GetConfig() *IPFilterConfig {
	return f.config
}

// GetCountryForIP returns the country code for an IP address
func (f *GeoFilter) GetCountryForIP(ip net.IP) string {
	if f.geoDB == nil {
		return ""
	}

	record, err := f.geoDB.Country(ip)
	if err != nil {
		return ""
	}

	return strings.ToUpper(record.Country.IsoCode)
}

// Close closes the GeoIP database
func (f *GeoFilter) Close() error {
	if f.geoDB != nil {
		return f.geoDB.Close()
	}
	return nil
}
