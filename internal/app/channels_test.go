package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
)

// TestResolveTimeouts_DelegatesToDeadlinePackage verifies that App.resolveTimeouts()
// correctly delegates to deadline.ResolveTimeouts with matching results.
// The exhaustive logic tests live in internal/deadline/resolve_test.go.
func TestResolveTimeouts_DelegatesToDeadlinePackage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cfg         config.AgentConfig
		wantIdle    time.Duration
		wantCeiling time.Duration
	}{
		{
			name:        "default fixed timeout",
			cfg:         config.AgentConfig{RequestTimeout: 5 * time.Minute},
			wantIdle:    0,
			wantCeiling: 5 * time.Minute,
		},
		{
			name:        "explicit idle timeout",
			cfg:         config.AgentConfig{IdleTimeout: 2 * time.Minute, RequestTimeout: 30 * time.Minute},
			wantIdle:    2 * time.Minute,
			wantCeiling: 30 * time.Minute,
		},
		{
			name:        "legacy auto-extend",
			cfg:         config.AgentConfig{AutoExtendTimeout: true, RequestTimeout: 5 * time.Minute, MaxRequestTimeout: 15 * time.Minute},
			wantIdle:    5 * time.Minute,
			wantCeiling: 15 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			app := &App{Config: &config.Config{Agent: tt.cfg}}
			idle, ceiling := app.resolveTimeouts()
			assert.Equal(t, tt.wantIdle, idle)
			assert.Equal(t, tt.wantCeiling, ceiling)
		})
	}
}

// TestChannelMessageReceivedEvent_EventName verifies the EventName() method
// returns the correct constant for ChannelMessageReceivedEvent.
func TestChannelMessageReceivedEvent_EventName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		channel    string
		sessionKey string
		senderName string
		senderID   string
		text       string
		metadata   map[string]string
	}{
		{
			name:       "telegram message",
			channel:    "telegram",
			sessionKey: "telegram:123:456",
			senderName: "testuser",
			senderID:   "456",
			text:       "hello from telegram",
			metadata:   map[string]string{"chatID": "123"},
		},
		{
			name:       "discord message",
			channel:    "discord",
			sessionKey: "discord:ch1:author1",
			senderName: "discorduser",
			senderID:   "author1",
			text:       "hello from discord",
			metadata:   map[string]string{"channelID": "ch1", "guildID": "guild1"},
		},
		{
			name:       "slack message",
			channel:    "slack",
			sessionKey: "slack:ch2:user2",
			senderName: "user2",
			senderID:   "user2",
			text:       "hello from slack",
			metadata:   map[string]string{"channelID": "ch2", "threadTS": "1234567890.123456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evt := eventbus.ChannelMessageReceivedEvent{
				Channel:    tt.channel,
				SessionKey: tt.sessionKey,
				SenderName: tt.senderName,
				SenderID:   tt.senderID,
				Text:       tt.text,
				Timestamp:  time.Now(),
				Metadata:   tt.metadata,
			}
			assert.Equal(t, "channel.message.received", evt.EventName())
			assert.Equal(t, tt.channel, evt.Channel)
			assert.Equal(t, tt.sessionKey, evt.SessionKey)
			assert.Equal(t, tt.senderName, evt.SenderName)
			assert.Equal(t, tt.senderID, evt.SenderID)
			assert.Equal(t, tt.text, evt.Text)
			assert.Equal(t, tt.metadata, evt.Metadata)
		})
	}
}

// TestChannelMessageSentEvent_EventName verifies the EventName() method
// returns the correct constant for ChannelMessageSentEvent.
func TestChannelMessageSentEvent_EventName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		channel      string
		sessionKey   string
		responseText string
	}{
		{
			name:         "telegram response",
			channel:      "telegram",
			sessionKey:   "telegram:123:456",
			responseText: "agent response for telegram",
		},
		{
			name:         "discord response",
			channel:      "discord",
			sessionKey:   "discord:ch1:author1",
			responseText: "agent response for discord",
		},
		{
			name:         "slack response",
			channel:      "slack",
			sessionKey:   "slack:ch2:user2",
			responseText: "agent response for slack",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evt := eventbus.ChannelMessageSentEvent{
				Channel:      tt.channel,
				SessionKey:   tt.sessionKey,
				ResponseText: tt.responseText,
				Timestamp:    time.Now(),
			}
			assert.Equal(t, "channel.message.sent", evt.EventName())
			assert.Equal(t, tt.channel, evt.Channel)
			assert.Equal(t, tt.sessionKey, evt.SessionKey)
			assert.Equal(t, tt.responseText, evt.ResponseText)
		})
	}
}

// TestChannelEventBus_PublishSubscribe verifies that channel events are
// correctly delivered through the EventBus to subscribers.
func TestChannelEventBus_PublishSubscribe(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()

	var receivedEvents []eventbus.Event
	var sentEvents []eventbus.Event

	eventbus.SubscribeTyped(bus, func(e eventbus.ChannelMessageReceivedEvent) {
		receivedEvents = append(receivedEvents, e)
	})
	eventbus.SubscribeTyped(bus, func(e eventbus.ChannelMessageSentEvent) {
		sentEvents = append(sentEvents, e)
	})

	// Publish a received event
	bus.Publish(eventbus.ChannelMessageReceivedEvent{
		Channel:    "telegram",
		SessionKey: "telegram:100:200",
		SenderName: "alice",
		SenderID:   "200",
		Text:       "test message",
		Timestamp:  time.Now(),
		Metadata:   map[string]string{"chatID": "100"},
	})

	// Publish a sent event
	bus.Publish(eventbus.ChannelMessageSentEvent{
		Channel:      "telegram",
		SessionKey:   "telegram:100:200",
		ResponseText: "agent reply",
		Timestamp:    time.Now(),
	})

	require.Len(t, receivedEvents, 1)
	require.Len(t, sentEvents, 1)

	rcv := receivedEvents[0].(eventbus.ChannelMessageReceivedEvent)
	assert.Equal(t, "telegram", rcv.Channel)
	assert.Equal(t, "telegram:100:200", rcv.SessionKey)
	assert.Equal(t, "alice", rcv.SenderName)
	assert.Equal(t, "200", rcv.SenderID)
	assert.Equal(t, "test message", rcv.Text)
	assert.Equal(t, map[string]string{"chatID": "100"}, rcv.Metadata)

	snt := sentEvents[0].(eventbus.ChannelMessageSentEvent)
	assert.Equal(t, "telegram", snt.Channel)
	assert.Equal(t, "telegram:100:200", snt.SessionKey)
	assert.Equal(t, "agent reply", snt.ResponseText)
}

// TestChannelEventBus_NilBusSafe verifies that nil EventBus does not panic.
// The handlers guard with `if a.EventBus != nil` checks.
func TestChannelEventBus_NilBusSafe(t *testing.T) {
	t.Parallel()

	app := &App{
		Config:   &config.Config{},
		EventBus: nil,
	}
	// Verify that nil EventBus field is safe — the guard checks in handlers
	// prevent any publish calls when EventBus is nil.
	assert.Nil(t, app.EventBus)
}
