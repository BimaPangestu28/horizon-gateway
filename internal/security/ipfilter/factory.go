package ipfilter

// NewIPFilter creates an IP filter based on configuration
func NewIPFilter(config *IPFilterConfig) (IPFilter, error) {
	if !config.Enabled {
		return nil, nil
	}

	// Check if we need geolocation filtering
	if (len(config.Countries) > 0 || len(config.Regions) > 0) && config.GeoDBPath != "" {
		return NewGeoFilter(config)
	}

	// Otherwise, use simple CIDR filter
	return NewCIDRFilter(config)
}
