package provider

import (
	"context"
	"iter"
)

// StreamEventType defines the type of event in a generation stream.
type StreamEventType string

const (
	StreamEventPlainText StreamEventType = "text_delta"
	StreamEventToolCall  StreamEventType = "tool_call"
	StreamEventError     StreamEventType = "error"
	StreamEventDone      StreamEventType = "done"
)

// Valid reports whether t is a known stream event type.
func (t StreamEventType) Valid() bool {
	switch t {
	case StreamEventPlainText, StreamEventToolCall, StreamEventError, StreamEventDone:
		return true
	}
	return false
}

// Values returns all known stream event types.
func (t StreamEventType) Values() []StreamEventType {
	return []StreamEventType{StreamEventPlainText, StreamEventToolCall, StreamEventError, StreamEventDone}
}

// StreamEvent represents a single event in the generation stream.
type StreamEvent struct {
	Type     StreamEventType
	Text     string
	ToolCall *ToolCall
	Error    error
}

// ToolCall represents a request for tool execution.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string // JSON string
}

// Message represents a chat message.
type Message struct {
	Role      string
	Content   string
	ToolCalls []ToolCall
	Metadata  map[string]interface{}
}

// Tool represents a tool definition.
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]interface{} // JSON schema
}

// GenerateParams contains parameters for generation.
type GenerateParams struct {
	Model       string
	Messages    []Message
	Tools       []Tool
	Temperature float64
	MaxTokens   int
}

// ModelInfo describes an available model.
type ModelInfo struct {
	ID             string
	Name           string
	ContextWindow  int
	SupportsVision bool
	SupportsTools  bool
	IsReasoning    bool
}

// Provider defines the interface for LLM providers.
type Provider interface {
	// ID returns the unique identifier of the provider (e.g., "openai", "anthropic").
	ID() string

	// Generate streams responses for the given conversation.
	// It returns an iterator that yields StreamEvents.
	Generate(ctx context.Context, params GenerateParams) (iter.Seq2[StreamEvent, error], error)

	// ListModels returns a list of available models.
	ListModels(ctx context.Context) ([]ModelInfo, error)
}
