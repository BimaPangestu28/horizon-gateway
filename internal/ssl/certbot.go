package ssl

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bimapangestu28/horizon/internal/config"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

// CertbotManager handles SSL certificate management with Certbot
type CertbotManager struct {
	logger        logging.Logger
	domains       []string
	email         string
	certPath      string
	renewInterval time.Duration
	certbotBin    string
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	reloadHook    func() error
}

// NewCertbotManager creates a new Certbot certificate manager
func NewCertbotManager(cfg *config.CertbotConfig, logger logging.Logger, reloadHook func() error) (*CertbotManager, error) {
	if len(cfg.Domains) == 0 {
		return nil, fmt.Errorf("no domains specified for SSL certificates")
	}

	if cfg.Email == "" {
		return nil, fmt.Errorf("email is required for Let's Encrypt certificate registration")
	}

	// Use default certbot if not specified
	certbotBin := cfg.CertbotBin
	if certbotBin == "" {
		// Try to find certbot in PATH
		var err error
		certbotBin, err = exec.LookPath("certbot")
		if err != nil {
			return nil, fmt.Errorf("certbot not found in PATH: %w", err)
		}
	}

	// Use default cert path if not specified
	certPath := cfg.CertPath
	if certPath == "" {
		certPath = "/etc/letsencrypt"
	}

	// Use default renewal interval if not specified
	renewInterval := cfg.RenewInterval
	if renewInterval == 0 {
		renewInterval = 12 * time.Hour // Check twice a day
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &CertbotManager{
		logger:        logger,
		domains:       cfg.Domains,
		email:         cfg.Email,
		certPath:      certPath,
		renewInterval: renewInterval,
		certbotBin:    certbotBin,
		ctx:           ctx,
		cancel:        cancel,
		reloadHook:    reloadHook,
	}, nil
}

// Start begins the certificate management process
func (m *CertbotManager) Start() error {
	m.logger.Info("Starting SSL certificate manager",
		"domains", strings.Join(m.domains, ","),
		"cert_path", m.certPath,
		"renew_interval", m.renewInterval.String())

	// Check and obtain initial certificates
	err := m.obtainCertificates()
	if err != nil {
		return fmt.Errorf("failed to obtain initial certificates: %w", err)
	}

	// Start renewal goroutine
	go m.renewalLoop()

	return nil
}

// Stop halts the certificate renewal process
func (m *CertbotManager) Stop() {
	m.cancel()
	m.logger.Info("SSL certificate manager stopped")
}

// GetCertificatePaths returns the paths to the certificate and key files for a domain
func (m *CertbotManager) GetCertificatePaths(domain string) (certFile, keyFile string, err error) {
	// Check if domain is managed
	found := false
	for _, d := range m.domains {
		if d == domain || (d[0] == '*' && strings.HasSuffix(domain, d[1:])) {
			found = true
			break
		}
	}

	if !found {
		return "", "", fmt.Errorf("domain %s not managed by this certificate manager", domain)
	}

	// Get actual domain directory (Certbot might fold wildcards or use the first domain for multiple domains)
	var certDomain string
	if strings.HasPrefix(domain, "*.") {
		// Certbot stores wildcard domains as _wildcard_.example.com
		parts := strings.SplitN(domain, ".", 2)
		if len(parts) < 2 {
			return "", "", fmt.Errorf("invalid wildcard domain: %s", domain)
		}
		certDomain = "_wildcard_." + parts[1]
	} else {
		certDomain = domain
	}

	// Try direct domain match first
	certPath := filepath.Join(m.certPath, "live", certDomain, "fullchain.pem")
	keyPath := filepath.Join(m.certPath, "live", certDomain, "privkey.pem")

	// Check if files exist
	if _, err := os.Stat(certPath); err == nil {
		if _, err := os.Stat(keyPath); err == nil {
			return certPath, keyPath, nil
		}
	}

	// If direct match fails, try the first domain (which Certbot might use)
	if len(m.domains) > 0 && m.domains[0] != domain {
		firstDomain := m.domains[0]
		if strings.HasPrefix(firstDomain, "*.") {
			parts := strings.SplitN(firstDomain, ".", 2)
			if len(parts) >= 2 {
				firstDomain = "_wildcard_." + parts[1]
			}
		}

		certPath = filepath.Join(m.certPath, "live", firstDomain, "fullchain.pem")
		keyPath = filepath.Join(m.certPath, "live", firstDomain, "privkey.pem")

		// Check if files exist
		if _, err := os.Stat(certPath); err == nil {
			if _, err := os.Stat(keyPath); err == nil {
				return certPath, keyPath, nil
			}
		}
	}

	return "", "", fmt.Errorf("certificate files for domain %s not found", domain)
}

// UpdateServerConfig updates the server's TLS configuration with the certificate paths
func (m *CertbotManager) UpdateServerConfig(serverCfg *config.ServerConfig) error {
	if len(m.domains) == 0 {
		return fmt.Errorf("no domains configured for SSL certificates")
	}

	// Use the first domain as the primary one
	primaryDomain := m.domains[0]
	certFile, keyFile, err := m.GetCertificatePaths(primaryDomain)
	if err != nil {
		return fmt.Errorf("failed to get certificate paths: %w", err)
	}

	// Create or update TLS config
	if serverCfg.TLS == nil {
		serverCfg.TLS = &config.TLSConfig{
			Enabled:  true,
			CertFile: certFile,
			KeyFile:  keyFile,
		}
	} else {
		serverCfg.TLS.Enabled = true
		serverCfg.TLS.CertFile = certFile
		serverCfg.TLS.KeyFile = keyFile
	}

	return nil
}

// obtainCertificates gets or renews certificates for all configured domains
func (m *CertbotManager) obtainCertificates() error {
	m.logger.Info("Checking/obtaining SSL certificates", "domains", strings.Join(m.domains, ","))

	// Build basic certbot command
	args := []string{
		"certonly",
		"--non-interactive",
		"--agree-tos",
		"--email", m.email,
		"--webroot",
		"--webroot-path", "/var/www/certbot",
	}

	// Add domains
	for _, domain := range m.domains {
		args = append(args, "-d", domain)
	}

	// Run certbot command
	cmd := exec.Command(m.certbotBin, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Check if it's because the cert already exists
		if strings.Contains(string(output), "Certificate not yet due for renewal") {
			m.logger.Info("Certificates already exist and are not due for renewal")
			return nil
		}

		m.logger.Error("Failed to obtain certificates",
			"error", err,
			"output", string(output))
		return fmt.Errorf("certbot error: %s: %w", string(output), err)
	}

	m.logger.Info("Successfully obtained/renewed certificates",
		"domains", strings.Join(m.domains, ","))

	// If we got this far, certificates were obtained or renewed
	if m.reloadHook != nil {
		if err := m.reloadHook(); err != nil {
			m.logger.Error("Failed to reload configuration after certificate renewal",
				"error", err)
		}
	}

	return nil
}

// renewalLoop periodically checks and renews certificates
func (m *CertbotManager) renewalLoop() {
	ticker := time.NewTicker(m.renewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := m.renewCertificates()
			if err != nil {
				m.logger.Error("Certificate renewal failed", "error", err)
			}
		case <-m.ctx.Done():
			return
		}
	}
}

// renewCertificates runs certbot renew to check and renew certificates
func (m *CertbotManager) renewCertificates() error {
	m.logger.Info("Checking for certificate renewals")

	// Run certbot renew
	cmd := exec.Command(m.certbotBin, "renew", "--non-interactive")
	output, err := cmd.CombinedOutput()

	if err != nil {
		m.logger.Error("Certbot renewal failed",
			"error", err,
			"output", string(output))
		return fmt.Errorf("certbot renewal error: %s: %w", string(output), err)
	}

	// Check if any renewals were performed
	if strings.Contains(string(output), "Cert not due for renewal") {
		m.logger.Info("No certificates due for renewal")
		return nil
	}

	if strings.Contains(string(output), "Congratulations") ||
		strings.Contains(string(output), "Successfully renewed") {
		m.logger.Info("Successfully renewed certificates")

		// Call the reload hook to update the server with new certificates
		if m.reloadHook != nil {
			if err := m.reloadHook(); err != nil {
				m.logger.Error("Failed to reload configuration after certificate renewal",
					"error", err)
			}
		}
	}

	return nil
}
