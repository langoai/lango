// Package agentpool manages a pool of discovered P2P agents with health checking,
// weighted selection, and capability-based filtering.
package agentpool

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Sentinel errors for pool operations.
var (
	ErrNoAgents    = errors.New("no agents available")
	ErrAgentExists = errors.New("agent already registered")
	ErrNotFound    = errors.New("agent not found")
)

// AgentStatus represents the health status of a pooled agent.
type AgentStatus string

const (
	StatusHealthy   AgentStatus = "healthy"
	StatusDegraded  AgentStatus = "degraded"
	StatusUnhealthy AgentStatus = "unhealthy"
	StatusUnknown   AgentStatus = "unknown"
)

// AgentPerformance tracks runtime performance metrics for a pooled agent.
type AgentPerformance struct {
	AvgLatencyMs float64 `json:"avgLatencyMs"`
	SuccessRate  float64 `json:"successRate"`
	TotalCalls   int     `json:"totalCalls"`
}

// Agent represents a discovered P2P agent in the pool.
type Agent struct {
	DID          string            `json:"did"`
	Name         string            `json:"name"`
	PeerID       string            `json:"peerId"`
	Capabilities []string          `json:"capabilities"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Status       AgentStatus       `json:"status"`
	TrustScore   float64           `json:"trustScore"`
	PricePerCall float64           `json:"pricePerCall"`
	Available    bool              `json:"available"`
	Performance  AgentPerformance  `json:"performance"`
	Latency      time.Duration     `json:"latency"`
	LastSeen     time.Time         `json:"lastSeen"`
	LastHealthy  time.Time         `json:"lastHealthy"`
	FailCount    int               `json:"failCount"`
}

// HasCapability reports whether the agent advertises the given capability.
func (a *Agent) HasCapability(cap string) bool {
	for _, c := range a.Capabilities {
		if c == cap {
			return true
		}
	}
	return false
}

// Pool manages a set of P2P agents with thread-safe access.
type Pool struct {
	mu     sync.RWMutex
	agents map[string]*Agent // keyed by DID
	logger *zap.SugaredLogger
}

// New creates an empty agent pool.
func New(logger *zap.SugaredLogger) *Pool {
	return &Pool{
		agents: make(map[string]*Agent),
		logger: logger,
	}
}

// Add registers an agent in the pool. Returns ErrAgentExists if the DID is already registered.
func (p *Pool) Add(agent *Agent) error {
	if agent.DID == "" {
		return fmt.Errorf("agent DID is empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.agents[agent.DID]; ok {
		return ErrAgentExists
	}

	if agent.Status == "" {
		agent.Status = StatusUnknown
	}
	if agent.LastSeen.IsZero() {
		agent.LastSeen = time.Now()
	}

	p.agents[agent.DID] = agent
	p.logger.Debugw("agent added to pool", "did", agent.DID, "name", agent.Name)
	return nil
}

// Update replaces an agent in the pool. The agent is matched by DID.
func (p *Pool) Update(agent *Agent) {
	p.mu.Lock()
	defer p.mu.Unlock()

	agent.LastSeen = time.Now()
	p.agents[agent.DID] = agent
}

// Remove removes an agent from the pool by DID.
func (p *Pool) Remove(did string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.agents, did)
	p.logger.Debugw("agent removed from pool", "did", did)
}

// Get returns an agent by DID or nil if not found.
func (p *Pool) Get(did string) *Agent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.agents[did]
}

// List returns all agents in the pool.
func (p *Pool) List() []*Agent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*Agent, 0, len(p.agents))
	for _, a := range p.agents {
		result = append(result, a)
	}
	return result
}

// Size returns the number of agents in the pool.
func (p *Pool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.agents)
}

// FindByCapability returns all healthy agents that advertise the given capability.
func (p *Pool) FindByCapability(cap string) []*Agent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []*Agent
	for _, a := range p.agents {
		if a.HasCapability(cap) && a.Status != StatusUnhealthy {
			result = append(result, a)
		}
	}
	return result
}

// UpdatePerformance records a call outcome and recalculates running averages.
func (p *Pool) UpdatePerformance(did string, latencyMs float64, success bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	a, ok := p.agents[did]
	if !ok {
		return
	}

	perf := &a.Performance
	perf.TotalCalls++
	// Running average for latency.
	perf.AvgLatencyMs = perf.AvgLatencyMs + (latencyMs-perf.AvgLatencyMs)/float64(perf.TotalCalls)
	// Running average for success rate.
	var s float64
	if success {
		s = 1.0
	}
	perf.SuccessRate = perf.SuccessRate + (s-perf.SuccessRate)/float64(perf.TotalCalls)
}

// MarkHealthy updates an agent's status and records the health check time.
func (p *Pool) MarkHealthy(did string, latency time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	a, ok := p.agents[did]
	if !ok {
		return
	}
	a.Status = StatusHealthy
	a.Latency = latency
	a.LastHealthy = time.Now()
	a.FailCount = 0
}

// MarkUnhealthy updates an agent's status after a failed health check.
func (p *Pool) MarkUnhealthy(did string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	a, ok := p.agents[did]
	if !ok {
		return
	}
	a.FailCount++
	if a.FailCount >= 3 {
		a.Status = StatusUnhealthy
	} else {
		a.Status = StatusDegraded
	}
}

// EvictStale removes agents not seen within the given threshold.
func (p *Pool) EvictStale(threshold time.Duration) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	cutoff := time.Now().Add(-threshold)
	evicted := 0
	for did, a := range p.agents {
		if a.LastSeen.Before(cutoff) {
			delete(p.agents, did)
			evicted++
			p.logger.Debugw("evicted stale agent", "did", did, "lastSeen", a.LastSeen)
		}
	}
	return evicted
}

// HealthCheckFunc pings an agent and returns its latency if reachable.
type HealthCheckFunc func(ctx context.Context, agent *Agent) (time.Duration, error)

// HealthChecker periodically checks agent health.
type HealthChecker struct {
	pool     *Pool
	checkFn  HealthCheckFunc
	interval time.Duration
	cancel   context.CancelFunc
	logger   *zap.SugaredLogger
}

// NewHealthChecker creates a health checker for the given pool.
func NewHealthChecker(pool *Pool, checkFn HealthCheckFunc, interval time.Duration, logger *zap.SugaredLogger) *HealthChecker {
	return &HealthChecker{
		pool:     pool,
		checkFn:  checkFn,
		interval: interval,
		logger:   logger,
	}
}

// Start begins periodic health checking.
func (hc *HealthChecker) Start(wg *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(context.Background())
	hc.cancel = cancel

	wg.Add(1)
	go func() {
		defer wg.Done()
		hc.loop(ctx)
	}()

	hc.logger.Infow("health checker started", "interval", hc.interval)
}

// Stop halts the health checker.
func (hc *HealthChecker) Stop() {
	if hc.cancel != nil {
		hc.cancel()
	}
}

func (hc *HealthChecker) loop(ctx context.Context) {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.checkAll(ctx)
		}
	}
}

func (hc *HealthChecker) checkAll(ctx context.Context) {
	agents := hc.pool.List()
	var wg sync.WaitGroup
	wg.Add(len(agents))
	for _, a := range agents {
		go func(a *Agent) {
			defer wg.Done()
			latency, err := hc.checkFn(ctx, a)
			if err != nil {
				hc.pool.MarkUnhealthy(a.DID)
				hc.logger.Debugw("health check failed", "did", a.DID, "error", err)
			} else {
				hc.pool.MarkHealthy(a.DID, latency)
			}
		}(a)
	}
	wg.Wait()
}

// SelectorWeights configures the relative importance of selection criteria.
type SelectorWeights struct {
	Trust        float64 // weight for trust score [0,1]
	Capability   float64 // weight for capability match breadth
	Performance  float64 // weight for success rate / latency
	Price        float64 // weight for price (lower is better)
	Availability float64 // weight for availability / health status
	// Legacy aliases (used if the new fields are zero).
	Latency float64 // weight for latency (lower is better)
	Health  float64 // weight for health status
}

// DefaultWeights returns production-default selector weights.
func DefaultWeights() SelectorWeights {
	return SelectorWeights{
		Trust:        0.35,
		Capability:   0.25,
		Performance:  0.20,
		Price:        0.15,
		Availability: 0.05,
	}
}

// Selector picks agents from the pool using weighted scoring.
type Selector struct {
	pool    *Pool
	weights SelectorWeights
}

// NewSelector creates a weighted selector for the given pool.
func NewSelector(pool *Pool, weights SelectorWeights) *Selector {
	return &Selector{pool: pool, weights: weights}
}

// Select picks the best agent for the given capability.
// Returns ErrNoAgents if no suitable agent is found.
func (s *Selector) Select(capability string) (*Agent, error) {
	candidates := s.pool.FindByCapability(capability)
	if len(candidates) == 0 {
		return nil, ErrNoAgents
	}

	var best *Agent
	bestScore := -1.0

	for _, a := range candidates {
		score := s.score(a)
		if score > bestScore {
			bestScore = score
			best = a
		}
	}

	return best, nil
}

// SelectN picks the top N agents for the given capability.
func (s *Selector) SelectN(capability string, n int) ([]*Agent, error) {
	candidates := s.pool.FindByCapability(capability)
	if len(candidates) == 0 {
		return nil, ErrNoAgents
	}

	type scored struct {
		agent *Agent
		score float64
	}

	items := make([]scored, len(candidates))
	for i, a := range candidates {
		items[i] = scored{agent: a, score: s.score(a)}
	}

	// Simple selection sort for top N (N is expected to be small).
	for i := range min(n, len(items)) {
		for j := i + 1; j < len(items); j++ {
			if items[j].score > items[i].score {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	count := min(n, len(items))
	result := make([]*Agent, count)
	for i := range count {
		result[i] = items[i].agent
	}
	return result, nil
}

// SelectRandom picks a random agent from healthy candidates for the given capability.
func (s *Selector) SelectRandom(capability string) (*Agent, error) {
	candidates := s.pool.FindByCapability(capability)
	if len(candidates) == 0 {
		return nil, ErrNoAgents
	}
	return candidates[rand.IntN(len(candidates))], nil
}

// ScoreWithCaps computes a weighted score considering required capabilities.
func (s *Selector) ScoreWithCaps(a *Agent, requiredCaps []string) float64 {
	return s.scoreAgent(a, requiredCaps)
}

// SelectBest picks the top N agents for the given required capabilities.
func (s *Selector) SelectBest(agents []*Agent, requiredCaps []string, n int) []*Agent {
	if len(agents) == 0 {
		return nil
	}

	type scored struct {
		agent *Agent
		score float64
	}

	items := make([]scored, len(agents))
	for i, a := range agents {
		items[i] = scored{agent: a, score: s.scoreAgent(a, requiredCaps)}
	}

	for i := range min(n, len(items)) {
		for j := i + 1; j < len(items); j++ {
			if items[j].score > items[i].score {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	count := min(n, len(items))
	result := make([]*Agent, count)
	for i := range count {
		result[i] = items[i].agent
	}
	return result
}

// score computes a weighted score for an agent. Higher is better.
func (s *Selector) score(a *Agent) float64 {
	return s.scoreAgent(a, nil)
}

// scoreAgent computes the full weighted score with optional capability matching.
func (s *Selector) scoreAgent(a *Agent, requiredCaps []string) float64 {
	w := s.weights
	trustComponent := a.TrustScore * w.Trust

	// Capability match: fraction of required caps that the agent supports.
	var capComponent float64
	if len(requiredCaps) > 0 && w.Capability > 0 {
		matched := 0
		for _, rc := range requiredCaps {
			if a.HasCapability(rc) {
				matched++
			}
		}
		capComponent = (float64(matched) / float64(len(requiredCaps))) * w.Capability
	} else if w.Capability > 0 {
		capComponent = w.Capability // no required caps → full score
	}

	// Performance component: success rate (0..1).
	var perfComponent float64
	if w.Performance > 0 {
		perfComponent = a.Performance.SuccessRate * w.Performance
	}

	// Price component: lower is better; normalize against a reference of 10.0.
	var priceComponent float64
	if w.Price > 0 {
		priceNorm := 1.0 - min(a.PricePerCall/10.0, 1.0)
		priceComponent = priceNorm * w.Price
	}

	// Availability / health component.
	availWeight := w.Availability
	if availWeight == 0 {
		availWeight = w.Health // legacy fallback
	}
	var availComponent float64
	switch a.Status {
	case StatusHealthy:
		availComponent = availWeight
	case StatusDegraded:
		availComponent = availWeight * 0.5
	}

	// Legacy latency component (used when Performance weight is zero).
	var latencyComponent float64
	if w.Latency > 0 && w.Performance == 0 {
		latencyMs := float64(a.Latency.Milliseconds())
		latencyNorm := 1.0 - min(latencyMs/10000.0, 1.0)
		latencyComponent = latencyNorm * w.Latency
	}

	return trustComponent + capComponent + perfComponent + priceComponent + availComponent + latencyComponent
}
