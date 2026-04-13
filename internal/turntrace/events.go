package turntrace

// EventType classifies a trace event. Defined as a type alias (not a new type)
// for backward compatibility with existing Ent schema string fields and DB rows.
type EventType = string

const (
	EventToolCall         EventType = "tool_call"
	EventToolResult       EventType = "tool_result"
	EventDelegation       EventType = "delegation"
	EventDelegationReturn EventType = "delegation_return"
	EventText             EventType = "text"
	EventTerminalError    EventType = "terminal_error"
	EventBudgetWarning    EventType = "budget_warning"
	EventRecoveryAttempt  EventType = "recovery_attempt"
	EventPolicyDecision  EventType = "policy_decision"
)
