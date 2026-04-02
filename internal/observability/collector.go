package observability

import (
	"sort"
	"sync"
	"time"
)

// DefaultMaxSessions is the default session map capacity before LRU eviction.
const DefaultMaxSessions = 10000

// MetricsCollector performs thread-safe in-memory metrics aggregation.
type MetricsCollector struct {
	mu        sync.RWMutex
	startedAt time.Time

	totalTokens TokenUsageSummary
	sessions    map[string]*SessionMetric
	agents      map[string]*AgentMetric

	toolExecs int64
	tools     map[string]*ToolMetric

	// Policy decision counters.
	policyBlocks   int64
	policyObserves int64
	policyByReason map[string]int64

	// MaxSessions caps the session map size; 0 means DefaultMaxSessions.
	MaxSessions int
}

// NewCollector creates a new MetricsCollector.
func NewCollector() *MetricsCollector {
	return &MetricsCollector{
		startedAt:      time.Now(),
		sessions:       make(map[string]*SessionMetric),
		agents:         make(map[string]*AgentMetric),
		tools:          make(map[string]*ToolMetric),
		policyByReason: make(map[string]int64),
		MaxSessions:    DefaultMaxSessions,
	}
}

// RecordTokenUsage records a token usage event.
func (c *MetricsCollector) RecordTokenUsage(usage TokenUsage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalTokens.InputTokens += usage.InputTokens
	c.totalTokens.OutputTokens += usage.OutputTokens
	c.totalTokens.TotalTokens += usage.TotalTokens
	c.totalTokens.CacheTokens += usage.CacheTokens

	if usage.SessionKey != "" {
		sm, ok := c.sessions[usage.SessionKey]
		if !ok {
			c.evictOldestSession()
			sm = &SessionMetric{SessionKey: usage.SessionKey}
			c.sessions[usage.SessionKey] = sm
		}
		sm.InputTokens += usage.InputTokens
		sm.OutputTokens += usage.OutputTokens
		sm.TotalTokens += usage.TotalTokens
		sm.RequestCount++
		sm.LastUpdated = time.Now()
	}

	if usage.AgentName != "" {
		am, ok := c.agents[usage.AgentName]
		if !ok {
			am = &AgentMetric{Name: usage.AgentName}
			c.agents[usage.AgentName] = am
		}
		am.InputTokens += usage.InputTokens
		am.OutputTokens += usage.OutputTokens
	}
}

// RecordToolExecution records a tool execution event.
func (c *MetricsCollector) RecordToolExecution(name, agentName string, duration time.Duration, success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.toolExecs++

	tm, ok := c.tools[name]
	if !ok {
		tm = &ToolMetric{Name: name}
		c.tools[name] = tm
	}
	tm.Count++
	tm.TotalDuration += duration
	tm.AvgDuration = tm.TotalDuration / time.Duration(tm.Count)
	if !success {
		tm.Errors++
	}

	if agentName != "" {
		am, ok := c.agents[agentName]
		if !ok {
			am = &AgentMetric{Name: agentName}
			c.agents[agentName] = am
		}
		am.ToolCalls++
	}
}

// RecordPolicyDecision records a policy block or observe event.
// verdict is a string such as "block" or "observe".
func (c *MetricsCollector) RecordPolicyDecision(verdict, reason string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch verdict {
	case "block":
		c.policyBlocks++
	case "observe":
		c.policyObserves++
	}
	if reason != "" {
		c.policyByReason[reason]++
	}
}

// evictOldestSession removes the least-recently-updated session when
// the session map reaches MaxSessions capacity. Must be called with mu held.
func (c *MetricsCollector) evictOldestSession() {
	max := c.MaxSessions
	if max <= 0 {
		max = DefaultMaxSessions
	}
	if len(c.sessions) < max {
		return
	}
	var oldestKey string
	var oldestTime time.Time
	for k, sm := range c.sessions {
		if oldestKey == "" || sm.LastUpdated.Before(oldestTime) {
			oldestKey = k
			oldestTime = sm.LastUpdated
		}
	}
	if oldestKey != "" {
		delete(c.sessions, oldestKey)
	}
}

// Snapshot returns a point-in-time copy of all metrics.
func (c *MetricsCollector) Snapshot() SystemSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	byReason := make(map[string]int64, len(c.policyByReason))
	for k, v := range c.policyByReason {
		byReason[k] = v
	}

	snap := SystemSnapshot{
		StartedAt:        c.startedAt,
		Uptime:           time.Since(c.startedAt),
		TokenUsageTotal:  c.totalTokens,
		ToolExecutions:   c.toolExecs,
		ToolBreakdown:    make(map[string]ToolMetric, len(c.tools)),
		AgentBreakdown:   make(map[string]AgentMetric, len(c.agents)),
		SessionBreakdown: make(map[string]SessionMetric, len(c.sessions)),
		Policy: PolicyMetrics{
			Blocks:   c.policyBlocks,
			Observes: c.policyObserves,
			ByReason: byReason,
		},
	}

	for k, v := range c.tools {
		snap.ToolBreakdown[k] = *v
	}
	for k, v := range c.agents {
		snap.AgentBreakdown[k] = *v
	}
	for k, v := range c.sessions {
		snap.SessionBreakdown[k] = *v
	}

	return snap
}

// SessionMetrics returns metrics for a specific session, or nil if not found.
func (c *MetricsCollector) SessionMetrics(sessionKey string) *SessionMetric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sm, ok := c.sessions[sessionKey]
	if !ok {
		return nil
	}
	cp := *sm
	return &cp
}

// TopSessions returns the top N sessions by total tokens.
func (c *MetricsCollector) TopSessions(limit int) []SessionMetric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sessions := make([]SessionMetric, 0, len(c.sessions))
	for _, sm := range c.sessions {
		sessions = append(sessions, *sm)
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].TotalTokens > sessions[j].TotalTokens
	})
	if limit > 0 && limit < len(sessions) {
		sessions = sessions[:limit]
	}
	return sessions
}

// Reset clears all collected metrics.
func (c *MetricsCollector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.startedAt = time.Now()
	c.totalTokens = TokenUsageSummary{}
	c.sessions = make(map[string]*SessionMetric)
	c.agents = make(map[string]*AgentMetric)
	c.tools = make(map[string]*ToolMetric)
	c.toolExecs = 0
	c.policyBlocks = 0
	c.policyObserves = 0
	c.policyByReason = make(map[string]int64)
}
