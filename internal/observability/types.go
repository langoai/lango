package observability

import "time"

// TokenUsage records a single token usage event.
type TokenUsage struct {
	Provider     string
	Model        string
	SessionKey   string
	AgentName    string
	InputTokens  int64
	OutputTokens int64
	TotalTokens  int64
	CacheTokens  int64
	Timestamp    time.Time
}

// ToolMetric aggregates metrics for a single tool.
type ToolMetric struct {
	Name          string
	Count         int64
	Errors        int64
	TotalDuration time.Duration
	AvgDuration   time.Duration
}

// AgentMetric aggregates metrics for a single agent.
type AgentMetric struct {
	Name         string
	InputTokens  int64
	OutputTokens int64
	ToolCalls    int64
}

// SessionMetric aggregates metrics for a single session.
type SessionMetric struct {
	SessionKey   string
	InputTokens  int64
	OutputTokens int64
	TotalTokens  int64
	RequestCount int64
	LastUpdated  time.Time
}

// PolicyMetrics aggregates policy decision counts.
type PolicyMetrics struct {
	Blocks   int64            `json:"blocks"`
	Observes int64            `json:"observes"`
	ByReason map[string]int64 `json:"byReason"`
}

// SystemSnapshot is a point-in-time summary of system metrics.
type SystemSnapshot struct {
	StartedAt        time.Time
	Uptime           time.Duration
	TokenUsageTotal  TokenUsageSummary
	ToolExecutions   int64
	ToolBreakdown    map[string]ToolMetric
	AgentBreakdown   map[string]AgentMetric
	SessionBreakdown map[string]SessionMetric
	Policy           PolicyMetrics
}

// TokenUsageSummary aggregates token counts across all providers/models.
type TokenUsageSummary struct {
	InputTokens  int64
	OutputTokens int64
	TotalTokens  int64
	CacheTokens  int64
}
