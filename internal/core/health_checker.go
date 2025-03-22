package core

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

// HealthCheckConfig defines the configuration for health checks
type HealthCheckConfig struct {
	// Path is the endpoint path to check on the target
	Path string

	// Interval is how often to check the target
	Interval time.Duration

	// Timeout is how long to wait for a response
	Timeout time.Duration

	// HealthyThreshold is the number of consecutive successes required to mark a target as healthy
	HealthyThreshold int

	// UnhealthyThreshold is the number of consecutive failures required to mark a target as unhealthy
	UnhealthyThreshold int
}

// DefaultHealthCheckConfig provides default health check settings
var DefaultHealthCheckConfig = HealthCheckConfig{
	Path:               "/health",
	Interval:           10 * time.Second,
	Timeout:            5 * time.Second,
	HealthyThreshold:   2,
	UnhealthyThreshold: 3,
}

// HealthChecker performs health checks on upstream targets
type HealthChecker struct {
	loadBalancer *LoadBalancer
	config       HealthCheckConfig
	client       *http.Client
	logger       logging.Logger
	stopCh       chan struct{}
	wg           sync.WaitGroup
	checks       map[string]*healthCheck
	mu           sync.Mutex
}

// healthCheck tracks the health status of a target
type healthCheck struct {
	target            *Target
	successiveSuccess int
	successiveFailure int
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(lb *LoadBalancer, config HealthCheckConfig, logger logging.Logger) *HealthChecker {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &HealthChecker{
		loadBalancer: lb,
		config:       config,
		client:       client,
		logger:       logger,
		stopCh:       make(chan struct{}),
		checks:       make(map[string]*healthCheck),
	}
}

// Start begins the health checking process
func (hc *HealthChecker) Start() {
	hc.wg.Add(1)
	go hc.runChecks()
}

// Stop stops the health checking process
func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
	hc.wg.Wait()
}

// runChecks continuously performs health checks on all targets
func (hc *HealthChecker) runChecks() {
	defer hc.wg.Done()

	ticker := time.NewTicker(hc.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.checkAllTargets()
		case <-hc.stopCh:
			return
		}
	}
}

// checkAllTargets performs a health check on all targets
func (hc *HealthChecker) checkAllTargets() {
	// Get a copy of all targets
	hc.mu.Lock()

	// Initialize health checks for new targets
	for _, target := range hc.loadBalancer.targets {
		targetURL := target.URL.String()
		if _, exists := hc.checks[targetURL]; !exists {
			hc.checks[targetURL] = &healthCheck{
				target:            target,
				successiveSuccess: 0,
				successiveFailure: 0,
			}
		}
	}

	// Remove health checks for targets that no longer exist
	for targetURL := range hc.checks {
		exists := false
		for _, target := range hc.loadBalancer.targets {
			if target.URL.String() == targetURL {
				exists = true
				break
			}
		}

		if !exists {
			delete(hc.checks, targetURL)
		}
	}

	// Create a copy of checks to avoid holding the lock during HTTP requests
	checks := make(map[string]*healthCheck, len(hc.checks))
	for k, v := range hc.checks {
		checks[k] = v
	}

	hc.mu.Unlock()

	// Check each target in parallel
	var wg sync.WaitGroup
	for targetURL, check := range checks {
		wg.Add(1)

		go func(targetURL string, check *healthCheck) {
			defer wg.Done()

			// Perform the health check
			healthy := hc.checkTarget(check.target)

			// Update health check status
			hc.mu.Lock()
			defer hc.mu.Unlock()

			// The target might have been removed while we were checking
			if currentCheck, exists := hc.checks[targetURL]; exists {
				if healthy {
					currentCheck.successiveSuccess++
					currentCheck.successiveFailure = 0

					// Mark as healthy if threshold reached
					if !currentCheck.target.Available && currentCheck.successiveSuccess >= hc.config.HealthyThreshold {
						currentCheck.target.Available = true
						hc.logger.Info("Target marked as healthy", "target", targetURL)
					}
				} else {
					currentCheck.successiveFailure++
					currentCheck.successiveSuccess = 0

					// Mark as unhealthy if threshold reached
					if currentCheck.target.Available && currentCheck.successiveFailure >= hc.config.UnhealthyThreshold {
						currentCheck.target.Available = false
						hc.logger.Warn("Target marked as unhealthy", "target", targetURL)
					}
				}
			}
		}(targetURL, check)
	}

	// Wait for all checks to complete
	wg.Wait()
}

// checkTarget performs a health check on a single target
func (hc *HealthChecker) checkTarget(target *Target) bool {
	// Create the health check URL
	healthURL := *target.URL
	healthURL.Path = hc.config.Path

	// Create request with context and timeout
	ctx, cancel := context.WithTimeout(context.Background(), hc.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL.String(), nil)
	if err != nil {
		hc.logger.Error("Failed to create health check request", "target", target.URL.String(), "error", err)
		return false
	}

	// Perform the request
	resp, err := hc.client.Do(req)
	if err != nil {
		hc.logger.Warn("Health check failed", "target", target.URL.String(), "error", err)
		return false
	}
	defer resp.Body.Close()

	// Check response status
	healthy := resp.StatusCode >= 200 && resp.StatusCode < 300

	if !healthy {
		hc.logger.Warn("Health check returned unhealthy status",
			"target", target.URL.String(),
			"status", resp.StatusCode,
		)
	}

	return healthy
}
