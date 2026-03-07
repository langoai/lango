package slack

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockClient implements Client interface
type MockClient struct {
	mu                sync.Mutex
	AuthTestFunc      func() (*slack.AuthTestResponse, error)
	PostMessageFunc   func(channelID string, options ...slack.MsgOption) (string, string, error)
	UpdateMessageFunc func(channelID, timestamp string, options ...slack.MsgOption) (string, string, string, error)
	PostMessages      []struct {
		ChannelID string
		Options   []slack.MsgOption
	}
	UpdateMessages []struct {
		ChannelID string
		Timestamp string
		Options   []slack.MsgOption
	}
}

func (m *MockClient) DeleteMessage(channelID, messageTimestamp string) (string, string, error) {
	return channelID, messageTimestamp, nil
}

func (m *MockClient) AuthTest() (*slack.AuthTestResponse, error) {
	if m.AuthTestFunc != nil {
		return m.AuthTestFunc()
	}
	return &slack.AuthTestResponse{UserID: "bot-123", Team: "TestTeam"}, nil
}

func (m *MockClient) UpdateMessage(channelID, timestamp string, options ...slack.MsgOption) (string, string, string, error) {
	m.mu.Lock()
	m.UpdateMessages = append(m.UpdateMessages, struct {
		ChannelID string
		Timestamp string
		Options   []slack.MsgOption
	}{ChannelID: channelID, Timestamp: timestamp, Options: options})
	m.mu.Unlock()
	if m.UpdateMessageFunc != nil {
		return m.UpdateMessageFunc(channelID, timestamp, options...)
	}
	return channelID, timestamp, "", nil
}

func (m *MockClient) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	m.mu.Lock()
	m.PostMessages = append(m.PostMessages, struct {
		ChannelID string
		Options   []slack.MsgOption
	}{ChannelID: channelID, Options: options})
	m.mu.Unlock()
	if m.PostMessageFunc != nil {
		return m.PostMessageFunc(channelID, options...)
	}
	return "ts-123", "chan-123", nil
}

func (m *MockClient) getPostMessages() []struct {
	ChannelID string
	Options   []slack.MsgOption
} {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]struct {
		ChannelID string
		Options   []slack.MsgOption
	}, len(m.PostMessages))
	copy(result, m.PostMessages)
	return result
}

func (m *MockClient) getUpdateMessages() []struct {
	ChannelID string
	Timestamp string
	Options   []slack.MsgOption
} {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]struct {
		ChannelID string
		Timestamp string
		Options   []slack.MsgOption
	}, len(m.UpdateMessages))
	copy(result, m.UpdateMessages)
	return result
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
	t.Parallel()

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
	require.NoError(t, err)

	channel.SetHandler(func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error) {
		assert.Equal(t, "Hello", msg.Text)
		return &OutgoingMessage{Text: "World"}, nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, channel.Start(ctx))

	// Simulate incoming message event via Socket Mode
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
	<-time.After(200 * time.Millisecond)
	// With thinking indicator: expect 1 PostMessage (thinking placeholder)
	// + 1 UpdateMessage (replace placeholder with response)
	assert.NotEmpty(t, mockClient.getPostMessages(), "expected PostMessage to be called (thinking placeholder)")
	assert.NotEmpty(t, mockClient.getUpdateMessages(), "expected UpdateMessage to be called (replace placeholder)")
}

func TestSlackThinkingPlaceholder(t *testing.T) {
	t.Parallel()

	mockClient := &MockClient{
		PostMessageFunc: func(channelID string, options ...slack.MsgOption) (string, string, error) {
			return channelID, "placeholder-ts", nil
		},
	}
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
	require.NoError(t, err)

	handlerDone := make(chan struct{})
	channel.SetHandler(func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error) {
		defer close(handlerDone)
		return &OutgoingMessage{Text: "response text"}, nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, channel.Start(ctx))

	innerEvent := &slackevents.MessageEvent{
		Type:        "message",
		Text:        "Hello",
		User:        "user-2",
		Channel:     "chan-2",
		ChannelType: "im",
	}
	eventsAPIEvent := slackevents.EventsAPIEvent{
		Type:       slackevents.CallbackEvent,
		InnerEvent: slackevents.EventsAPIInnerEvent{Data: innerEvent},
	}
	mockSocket.EventsCh <- socketmode.Event{
		Type:    socketmode.EventTypeEventsAPI,
		Request: &socketmode.Request{},
		Data:    eventsAPIEvent,
	}

	select {
	case <-handlerDone:
		// Allow goroutine to finish posting
		time.Sleep(50 * time.Millisecond)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for handler")
	}

	// Verify: first PostMessage is the thinking placeholder, then UpdateMessage replaces it
	postMsgs := mockClient.getPostMessages()
	require.NotEmpty(t, postMsgs, "expected at least 1 PostMessage call")

	updateMsgs := mockClient.getUpdateMessages()
	require.NotEmpty(t, updateMsgs, "expected at least 1 UpdateMessage call")

	// Verify UpdateMessage was called with the placeholder timestamp
	assert.Equal(t, "placeholder-ts", updateMsgs[0].Timestamp)
}
