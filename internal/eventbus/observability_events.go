package eventbus

// TokenUsageEvent is published when an LLM provider returns token usage data.
// The observability TokenTracker subscribes to this event.
type TokenUsageEvent struct {
	Provider     string
	Model        string
	SessionKey   string
	AgentName    string
	InputTokens  int64
	OutputTokens int64
	TotalTokens  int64
	CacheTokens  int64
}

// EventName implements Event.
func (e TokenUsageEvent) EventName() string { return "token.usage" }
