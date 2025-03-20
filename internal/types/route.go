package types

import (
	"net/url"
	"time"
)

// Route represents a configured API route
type Route struct {
	Name        string
	ListenPath  string
	UpstreamURL string
	Methods     []string
	StripPath   bool
	Headers     map[string]string
	QueryParams map[string]string
	Host        string
	Priority    int
}

// Target represents an upstream service target
type Target struct {
	URL          *url.URL
	Weight       int
	Available    bool
	ActiveConns  int32
	Draining     bool
	DrainTimeout time.Duration
}
