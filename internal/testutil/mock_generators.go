package testutil

import (
	"context"
	"sync"
)

// MockTextGenerator is a thread-safe mock for the TextGenerator interface
// used in memory, graph, and learning packages.
type MockTextGenerator struct {
	mu sync.Mutex

	Response string
	Err      error
	calls    int
	lastArgs []string
}

// NewMockTextGenerator creates a MockTextGenerator with the given response.
func NewMockTextGenerator(response string) *MockTextGenerator {
	return &MockTextGenerator{Response: response}
}

func (m *MockTextGenerator) GenerateText(_ context.Context, systemPrompt, userPrompt string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	m.lastArgs = []string{systemPrompt, userPrompt}
	if m.Err != nil {
		return "", m.Err
	}
	return m.Response, nil
}

// Calls returns the number of GenerateText calls.
func (m *MockTextGenerator) Calls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

// LastArgs returns the last system and user prompts.
func (m *MockTextGenerator) LastArgs() (string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.lastArgs) < 2 {
		return "", ""
	}
	return m.lastArgs[0], m.lastArgs[1]
}

// MockAgentRunner is a thread-safe mock for cron.AgentRunner.
type MockAgentRunner struct {
	mu sync.Mutex

	Response string
	Err      error
	calls    int
	lastKey  string
}

// NewMockAgentRunner creates a MockAgentRunner with the given response.
func NewMockAgentRunner(response string) *MockAgentRunner {
	return &MockAgentRunner{Response: response}
}

func (m *MockAgentRunner) Run(_ context.Context, sessionKey string, _ string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	m.lastKey = sessionKey
	if m.Err != nil {
		return "", m.Err
	}
	return m.Response, nil
}

// Calls returns the number of Run calls.
func (m *MockAgentRunner) Calls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

// LastSessionKey returns the last session key passed to Run.
func (m *MockAgentRunner) LastSessionKey() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastKey
}

// MockChannelSender is a thread-safe mock for cron.ChannelSender.
type MockChannelSender struct {
	mu sync.Mutex

	Err      error
	calls    int
	messages []SentMessage
}

// SentMessage records a message sent via MockChannelSender.
type SentMessage struct {
	Channel string
	Message string
}

// NewMockChannelSender creates a MockChannelSender.
func NewMockChannelSender() *MockChannelSender {
	return &MockChannelSender{}
}

func (m *MockChannelSender) SendMessage(_ context.Context, channel string, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	m.messages = append(m.messages, SentMessage{Channel: channel, Message: message})
	if m.Err != nil {
		return m.Err
	}
	return nil
}

// Calls returns the number of SendMessage calls.
func (m *MockChannelSender) Calls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

// Messages returns all sent messages.
func (m *MockChannelSender) Messages() []SentMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]SentMessage, len(m.messages))
	copy(result, m.messages)
	return result
}
