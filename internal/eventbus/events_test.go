package eventbus

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannelMessageReceivedEvent_EventName(t *testing.T) {
	t.Parallel()

	e := ChannelMessageReceivedEvent{
		Channel:    "telegram",
		SessionKey: "telegram:123:456",
		SenderName: "alice",
		SenderID:   "12345",
		Text:       "hello",
		Timestamp:  time.Now(),
		Metadata:   map[string]string{"thread_id": "t1"},
	}

	assert.Equal(t, "channel.message.received", e.EventName())
}

func TestChannelMessageSentEvent_EventName(t *testing.T) {
	t.Parallel()

	e := ChannelMessageSentEvent{
		Channel:      "discord",
		SessionKey:   "discord:guild:chan",
		ResponseText: "hi there",
		Timestamp:    time.Now(),
	}

	assert.Equal(t, "channel.message.sent", e.EventName())
}

func TestChannelMessageReceivedEvent_RoundTrip(t *testing.T) {
	t.Parallel()

	bus := New()

	var got ChannelMessageReceivedEvent
	SubscribeTyped(bus, func(e ChannelMessageReceivedEvent) {
		got = e
	})

	now := time.Now()
	bus.Publish(ChannelMessageReceivedEvent{
		Channel:    "slack",
		SessionKey: "slack:C123:U456",
		SenderName: "bob",
		SenderID:   "U456",
		Text:       "check this out",
		Timestamp:  now,
		Metadata:   map[string]string{"thread_ts": "1234567890.123456"},
	})

	assert.Equal(t, "slack", got.Channel)
	assert.Equal(t, "slack:C123:U456", got.SessionKey)
	assert.Equal(t, "bob", got.SenderName)
	assert.Equal(t, "U456", got.SenderID)
	assert.Equal(t, "check this out", got.Text)
	assert.Equal(t, now, got.Timestamp)
	require.Len(t, got.Metadata, 1)
	assert.Equal(t, "1234567890.123456", got.Metadata["thread_ts"])
}

func TestChannelMessageSentEvent_RoundTrip(t *testing.T) {
	t.Parallel()

	bus := New()

	var got ChannelMessageSentEvent
	SubscribeTyped(bus, func(e ChannelMessageSentEvent) {
		got = e
	})

	now := time.Now()
	bus.Publish(ChannelMessageSentEvent{
		Channel:      "telegram",
		SessionKey:   "telegram:123:456",
		ResponseText: "I can help with that",
		Timestamp:    now,
	})

	assert.Equal(t, "telegram", got.Channel)
	assert.Equal(t, "telegram:123:456", got.SessionKey)
	assert.Equal(t, "I can help with that", got.ResponseText)
	assert.Equal(t, now, got.Timestamp)
}
