package eventbus

// Event name constant for observability domain events.
const EventTokenUsage = "token.usage"

// TokenUsageEvent is published when an LLM provider returns token usage data.
// The observability TokenTracker subscribes to this event.
type TokenUsageEvent struct {
	Provider         string
	Model            string
	SessionKey       string
	AgentName        string
	InputTokens      int64
	OutputTokens     int64
	TotalTokens      int64
	CacheTokens      int64
	EstimatedCostUSD float64 // 0 when model has no pricing entry; populated by emitter
}

// EventName implements Event.
func (e TokenUsageEvent) EventName() string { return EventTokenUsage }
