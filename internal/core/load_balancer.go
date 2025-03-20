package core

import (
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// LoadBalancerType defines the load balancing algorithm
type LoadBalancerType string

const (
	// LBTypeRoundRobin distributes requests in a rotating sequence
	LBTypeRoundRobin LoadBalancerType = "round_robin"

	// LBTypeWeighted distributes requests based on target weights
	LBTypeWeighted LoadBalancerType = "weighted"

	// LBTypeLeastConn distributes requests to the target with the least active connections
	LBTypeLeastConn LoadBalancerType = "least_conn"
)

// LoadBalancer distributes requests across multiple upstream targets
type LoadBalancer struct {
	targets []*Target
	lbType  LoadBalancerType
	current uint64 // For round-robin algorithm
	mu      sync.RWMutex
}

// Target represents an upstream service target
type Target struct {
	URL          *url.URL
	Weight       int // For weighted load balancing
	Available    bool
	ActiveConns  int32         // For least connections load balancing
	Draining     bool          // For connection draining during removal
	DrainTimeout time.Duration // Timeout for connection draining
}

// NewLoadBalancer creates a new load balancer with the given targets
func NewLoadBalancer(lbType LoadBalancerType) *LoadBalancer {
	return &LoadBalancer{
		targets: make([]*Target, 0),
		lbType:  lbType,
		current: 0,
	}
}

// AddTarget adds a new target to the load balancer
func (lb *LoadBalancer) AddTarget(targetURL string, weight int) error {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return err
	}

	target := &Target{
		URL:          parsed,
		Weight:       weight,
		Available:    true,
		ActiveConns:  0,
		Draining:     false,
		DrainTimeout: 30 * time.Second, // Default drain timeout
	}

	lb.mu.Lock()
	lb.targets = append(lb.targets, target)
	lb.mu.Unlock()

	return nil
}

// IncrementConn increments the active connection count for a target
func (lb *LoadBalancer) IncrementConn(target *Target) {
	atomic.AddInt32(&target.ActiveConns, 1)
}

// DecrementConn decrements the active connection count for a target
func (lb *LoadBalancer) DecrementConn(target *Target) {
	atomic.AddInt32(&target.ActiveConns, -1)
}

// RemoveTarget removes a target from the load balancer
func (lb *LoadBalancer) RemoveTarget(targetURL string) bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for i, target := range lb.targets {
		if target.URL.String() == targetURL {
			// Remove target by slicing
			lb.targets = append(lb.targets[:i], lb.targets[i+1:]...)
			return true
		}
	}

	return false
}

// DrainTarget starts graceful connection draining for a target
func (lb *LoadBalancer) DrainTarget(targetURL string, timeout time.Duration) bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for _, target := range lb.targets {
		if target.URL.String() == targetURL {
			// Mark as draining
			target.Draining = true

			if timeout > 0 {
				target.DrainTimeout = timeout
			}

			// Start a goroutine to monitor draining
			go lb.monitorDraining(target)
			return true
		}
	}

	return false
}

// monitorDraining monitors a draining target and removes it when safe
func (lb *LoadBalancer) monitorDraining(target *Target) {
	// Start a timer for the draining timeout
	timer := time.NewTimer(target.DrainTimeout)
	defer timer.Stop()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if all connections are done
			if atomic.LoadInt32(&target.ActiveConns) <= 0 {
				lb.RemoveTarget(target.URL.String())
				return
			}
		case <-timer.C:
			// Timeout expired, force removal
			lb.RemoveTarget(target.URL.String())
			return
		}
	}
}

// SetTargetAvailability sets the availability of a target
func (lb *LoadBalancer) SetTargetAvailability(targetURL string, available bool) bool {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	for _, target := range lb.targets {
		if target.URL.String() == targetURL {
			target.Available = available
			return true
		}
	}

	return false
}

// GetNextTarget returns the next target based on the load balancing algorithm
func (lb *LoadBalancer) GetNextTarget() (*Target, bool) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	// Check if we have any targets
	if len(lb.targets) == 0 {
		return nil, false
	}

	// Get available targets
	var availableTargets []*Target
	for _, target := range lb.targets {
		if target.Available && !target.Draining {
			availableTargets = append(availableTargets, target)
		}
	}

	// Check if we have any available targets
	if len(availableTargets) == 0 {
		// If no non-draining targets, check if we have any draining targets
		// This is a fallback to ensure requests are still served during draining
		for _, target := range lb.targets {
			if target.Available && target.Draining {
				availableTargets = append(availableTargets, target)
			}
		}

		// If still no targets, return false
		if len(availableTargets) == 0 {
			return nil, false
		}
	}

	// Select target based on load balancing algorithm
	switch lb.lbType {
	case LBTypeRoundRobin:
		return lb.roundRobin(availableTargets), true
	case LBTypeWeighted:
		return lb.weighted(availableTargets), true
	case LBTypeLeastConn:
		return lb.leastConnections(availableTargets), true
	default:
		// Default to round-robin
		return lb.roundRobin(availableTargets), true
	}
}

// leastConnections implements the least connections load balancing algorithm
func (lb *LoadBalancer) leastConnections(targets []*Target) *Target {
	if len(targets) == 0 {
		return nil
	}

	// Find the target with the least active connections
	leastConnTarget := targets[0]
	minConn := atomic.LoadInt32(&leastConnTarget.ActiveConns)

	for _, target := range targets[1:] {
		currConn := atomic.LoadInt32(&target.ActiveConns)
		if currConn < minConn {
			minConn = currConn
			leastConnTarget = target
		}
	}

	return leastConnTarget
}

// roundRobin implements the round-robin load balancing algorithm
func (lb *LoadBalancer) roundRobin(targets []*Target) *Target {
	// Get the next index with atomic operation to avoid race conditions
	idx := atomic.AddUint64(&lb.current, 1) % uint64(len(targets))
	return targets[idx]
}

// weighted implements the weighted load balancing algorithm
func (lb *LoadBalancer) weighted(targets []*Target) *Target {
	// This is a simplified implementation of weighted load balancing
	// A more sophisticated implementation would consider current load, etc.

	// Calculate total weight
	totalWeight := 0
	for _, target := range targets {
		totalWeight += target.Weight
	}

	// If all weights are zero, default to round-robin
	if totalWeight == 0 {
		return lb.roundRobin(targets)
	}

	// Select target based on weight
	// This is a simple implementation that doesn't distribute perfectly
	// A production implementation would use a more sophisticated algorithm
	idx := atomic.AddUint64(&lb.current, 1) % uint64(totalWeight)

	runningTotal := 0
	for _, target := range targets {
		runningTotal += target.Weight
		if uint64(runningTotal) > idx {
			return target
		}
	}

	// Default to the last target (should never happen, but just in case)
	return targets[len(targets)-1]
}
