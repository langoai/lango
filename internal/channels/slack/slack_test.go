package slack

import (
	"context"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// MockClient implements Client interface
type MockClient struct {
	AuthTestFunc    func() (*slack.AuthTestResponse, error)
	PostMessageFunc func(channelID string, options ...slack.MsgOption) (string, string, error)
	PostMessages    []struct {
		ChannelID string
		Options   []slack.MsgOption
	}
}

func (m *MockClient) AuthTest() (*slack.AuthTestResponse, error) {
	if m.AuthTestFunc != nil {
		return m.AuthTestFunc()
	}
	return &slack.AuthTestResponse{UserID: "bot-123", Team: "TestTeam"}, nil
}

func (m *MockClient) UpdateMessage(channelID, timestamp string, options ...slack.MsgOption) (string, string, string, error) {
	return channelID, timestamp, "", nil
}

func (m *MockClient) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	m.PostMessages = append(m.PostMessages, struct {
		ChannelID string
		Options   []slack.MsgOption
	}{ChannelID: channelID, Options: options})
	if m.PostMessageFunc != nil {
		return m.PostMessageFunc(channelID, options...)
	}
	return "ts-123", "chan-123", nil
}

// MockSocket implements Socket interface
type MockSocket struct {
	EventsCh chan socketmode.Event
}

func (m *MockSocket) Run() error {
	return nil
}

func (m *MockSocket) Ack(req socketmode.Request, payload ...interface{}) {
	// No-op
}

func (m *MockSocket) Events() <-chan socketmode.Event {
	return m.EventsCh
}

func TestSlackChannel(t *testing.T) {
	mockClient := &MockClient{}
	mockSocket := &MockSocket{
		EventsCh: make(chan socketmode.Event, 1),
	}

	cfg := Config{
		BotToken: "TEST_TOKEN",
		AppToken: "APP_TOKEN",
		Client:   mockClient,
		Socket:   mockSocket,
	}

	channel, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}

	channel.SetHandler(func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error) {
		if msg.Text != "Hello" {
			t.Errorf("expected 'Hello', got '%s'", msg.Text)
		}
		return &OutgoingMessage{Text: "World"}, nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := channel.Start(ctx); err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	// Simulate incoming message event via Socket Mode
	// We need to construct strict structure expected by handler
	innerEvent := &slackevents.MessageEvent{
		Type:        "message",
		Text:        "Hello",
		User:        "user-1",
		Channel:     "chan-1",
		ChannelType: "im",
	}

	// Events API Event
	eventsAPIEvent := slackevents.EventsAPIEvent{
		Type:       slackevents.CallbackEvent,
		InnerEvent: slackevents.EventsAPIInnerEvent{Data: innerEvent},
	}

	// Socket Mode Event
	mockSocket.EventsCh <- socketmode.Event{
		Type:    socketmode.EventTypeEventsAPI,
		Request: &socketmode.Request{},
		Data:    eventsAPIEvent,
	}

	// Wait for processing (async)
	// We can check if PostMessage was called
	select {
	case <-time.After(100 * time.Millisecond):
		if len(mockClient.PostMessages) == 0 {
			t.Error("expected PostMessage to be called")
		} else {
			// Verification
			// Slack MsgOption is opaque function, hard to verify content directly without applying it to a struct
			// But simpler: just checking call count is enough for basic unit test of flow.
		}
	}
}
