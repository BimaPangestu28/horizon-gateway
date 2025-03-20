package config

import (
	"regexp"
)

// Regular expression to match environment variables in the form of ${VAR} or $VAR
var envVarPattern = regexp.MustCompile(`\${([a-zA-Z0-9_]+)}|\$([a-zA-Z0-9_]+)`)
