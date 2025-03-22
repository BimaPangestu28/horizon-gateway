package discovery

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type TLSConfig struct {
	Enabled        bool   `yaml:"enabled"`
	CertFile       string `yaml:"cert_file"`
	KeyFile        string `yaml:"key_file"`
	CAFile         string `yaml:"ca_file"`
	SkipVerify     bool   `yaml:"skip_verify"`
	ServerName     string `yaml:"server_name"`
	MinVersion     string `yaml:"min_version"`
	MaxVersion     string `yaml:"max_version"`
	ClientAuthType string `yaml:"client_auth_type"`
}

func createTLSConfigFromFiles(certFile, keyFile, caFile string) (*tls.Config, error) {
	tlsConfig := &tls.Config{}

	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificate and key: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if caFile != "" {
		caPEM, err := ioutil.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("failed to append CA certificate to pool")
		}
		tlsConfig.RootCAs = certPool
	}

	return tlsConfig, nil
}

func createTLSConfigFromConfig(config *TLSConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.SkipVerify,
	}

	if config.ServerName != "" {
		tlsConfig.ServerName = config.ServerName
	}

	if config.CertFile != "" && config.KeyFile != "" {
		if !fileExists(config.CertFile) {
			return nil, fmt.Errorf("certificate file not found: %s", config.CertFile)
		}
		if !fileExists(config.KeyFile) {
			return nil, fmt.Errorf("key file not found: %s", config.KeyFile)
		}

		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificate and key: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if config.CAFile != "" {
		if !fileExists(config.CAFile) {
			return nil, fmt.Errorf("CA file not found: %s", config.CAFile)
		}

		caPEM, err := ioutil.ReadFile(config.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("failed to append CA certificate to pool")
		}
		tlsConfig.RootCAs = certPool
	}

	if config.MinVersion != "" {
		minVersion, err := parseTLSVersion(config.MinVersion)
		if err != nil {
			return nil, err
		}
		tlsConfig.MinVersion = minVersion
	}

	if config.MaxVersion != "" {
		maxVersion, err := parseTLSVersion(config.MaxVersion)
		if err != nil {
			return nil, err
		}
		tlsConfig.MaxVersion = maxVersion
	}

	switch config.ClientAuthType {
	case "NoClientCert":
		tlsConfig.ClientAuth = tls.NoClientCert
	case "RequestClientCert":
		tlsConfig.ClientAuth = tls.RequestClientCert
	case "RequireAnyClientCert":
		tlsConfig.ClientAuth = tls.RequireAnyClientCert
	case "VerifyClientCertIfGiven":
		tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven
	case "RequireAndVerifyClientCert":
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	case "":
		// Default is no client auth
	default:
		return nil, fmt.Errorf("unknown client auth type: %s", config.ClientAuthType)
	}

	return tlsConfig, nil
}

func parseTLSVersion(version string) (uint16, error) {
	switch version {
	case "1.0", "TLS10":
		return tls.VersionTLS10, nil
	case "1.1", "TLS11":
		return tls.VersionTLS11, nil
	case "1.2", "TLS12":
		return tls.VersionTLS12, nil
	case "1.3", "TLS13":
		return tls.VersionTLS13, nil
	default:
		return 0, fmt.Errorf("unknown TLS version: %s", version)
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

type ServiceInstanceWithTTL struct {
	ServiceInstance
	TTL time.Duration `json:"ttl,omitempty"`
}
