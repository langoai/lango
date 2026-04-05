package cockpit

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/eventbus"
)

type mockSender struct {
	msgs []tea.Msg
}

func (m *mockSender) Send(msg tea.Msg) {
	m.msgs = append(m.msgs, msg)
}

func TestSubscribeChannelEvents_NilBus(t *testing.T) {
	sender := &mockSender{}
	// Must not panic.
	SubscribeChannelEvents(nil, sender)
	assert.Empty(t, sender.msgs)
}

func TestSubscribeChannelEvents_NilSender(t *testing.T) {
	bus := eventbus.New()
	// Must not panic.
	SubscribeChannelEvents(bus, nil)
}

func TestSubscribeChannelEvents_ReceivesMessage(t *testing.T) {
	bus := eventbus.New()
	sender := &mockSender{}
	SubscribeChannelEvents(bus, sender)

	now := time.Now()
	bus.Publish(eventbus.ChannelMessageReceivedEvent{
		Channel:    "telegram",
		SessionKey: "telegram:123:456",
		SenderName: "alice",
		SenderID:   "uid-001",
		Text:       "hello world",
		Timestamp:  now,
		Metadata:   map[string]string{"thread": "t1"},
	})

	require.Len(t, sender.msgs, 1)
	msg, ok := sender.msgs[0].(chat.ChannelMessageMsg)
	require.True(t, ok, "expected chat.ChannelMessageMsg, got %T", sender.msgs[0])

	assert.Equal(t, "telegram", msg.Channel)
	assert.Equal(t, "telegram:123:456", msg.SessionKey)
	assert.Equal(t, "alice", msg.SenderName)
	assert.Equal(t, "hello world", msg.Text)
	assert.Equal(t, now, msg.Timestamp)
	assert.Equal(t, map[string]string{"thread": "t1"}, msg.Metadata)
}

func TestSubscribeChannelEvents_PreservesMetadata(t *testing.T) {
	bus := eventbus.New()
	sender := &mockSender{}
	SubscribeChannelEvents(bus, sender)

	meta := map[string]string{
		"ThreadTS": "1234567890.123456",
		"GuildID":  "guild-abc",
		"ReplyTo":  "msg-xyz",
	}

	bus.Publish(eventbus.ChannelMessageReceivedEvent{
		Channel:    "discord",
		SessionKey: "discord:guild-abc:chan-1",
		SenderName: "bob",
		SenderID:   "uid-002",
		Text:       "metadata test",
		Timestamp:  time.Now(),
		Metadata:   meta,
	})

	require.Len(t, sender.msgs, 1)
	msg, ok := sender.msgs[0].(chat.ChannelMessageMsg)
	require.True(t, ok)
	assert.Equal(t, meta, msg.Metadata)
	assert.Equal(t, "1234567890.123456", msg.Metadata["ThreadTS"])
	assert.Equal(t, "guild-abc", msg.Metadata["GuildID"])
	assert.Equal(t, "msg-xyz", msg.Metadata["ReplyTo"])
}

func TestSubscribeChannelEvents_MultipleEvents(t *testing.T) {
	bus := eventbus.New()
	sender := &mockSender{}
	SubscribeChannelEvents(bus, sender)

	channels := []string{"telegram", "discord", "slack"}
	for _, ch := range channels {
		bus.Publish(eventbus.ChannelMessageReceivedEvent{
			Channel:    ch,
			SessionKey: ch + ":session",
			SenderName: "user-" + ch,
			SenderID:   "id-" + ch,
			Text:       "msg from " + ch,
			Timestamp:  time.Now(),
		})
	}

	require.Len(t, sender.msgs, 3)
	for i, ch := range channels {
		msg, ok := sender.msgs[i].(chat.ChannelMessageMsg)
		require.True(t, ok, "message %d: expected chat.ChannelMessageMsg", i)
		assert.Equal(t, ch, msg.Channel)
		assert.Equal(t, "user-"+ch, msg.SenderName)
		assert.Equal(t, "msg from "+ch, msg.Text)
	}
}

// --- ChannelTracker tests ---

func TestNewChannelTracker_NilBus(t *testing.T) {
	tracker := NewChannelTracker(nil)
	assert.NotNil(t, tracker)
	assert.Empty(t, tracker.Snapshot())
}

func TestChannelTracker_SeedChannel(t *testing.T) {
	tracker := NewChannelTracker(nil)
	tracker.SeedChannel("telegram", true)
	tracker.SeedChannel("discord", false)

	snap := tracker.Snapshot()
	require.Len(t, snap, 2)

	// Snapshot is sorted by name: discord, telegram.
	assert.Equal(t, "discord", snap[0].Name)
	assert.False(t, snap[0].Connected)
	assert.Equal(t, 0, snap[0].MessageCount)

	assert.Equal(t, "telegram", snap[1].Name)
	assert.True(t, snap[1].Connected)
	assert.Equal(t, 0, snap[1].MessageCount)
}

func TestChannelTracker_SeedChannelUpdatesExisting(t *testing.T) {
	tracker := NewChannelTracker(nil)
	tracker.SeedChannel("telegram", false)

	snap := tracker.Snapshot()
	require.Len(t, snap, 1)
	assert.False(t, snap[0].Connected)

	// Seed again with connected=true — should update, not duplicate.
	tracker.SeedChannel("telegram", true)

	snap = tracker.Snapshot()
	require.Len(t, snap, 1)
	assert.True(t, snap[0].Connected)
}

func TestChannelTracker_MessageCounting(t *testing.T) {
	bus := eventbus.New()
	tracker := NewChannelTracker(bus)

	now := time.Now()
	bus.Publish(eventbus.ChannelMessageReceivedEvent{
		Channel:   "telegram",
		Timestamp: now,
	})
	bus.Publish(eventbus.ChannelMessageReceivedEvent{
		Channel:   "telegram",
		Timestamp: now.Add(time.Second),
	})

	snap := tracker.Snapshot()
	require.Len(t, snap, 1)
	assert.Equal(t, "telegram", snap[0].Name)
	assert.Equal(t, 2, snap[0].MessageCount)
	assert.True(t, snap[0].Connected)
	assert.Equal(t, now.Add(time.Second), snap[0].LastActivity)
}

func TestChannelTracker_SentEventUpdatesLastActivity(t *testing.T) {
	bus := eventbus.New()
	tracker := NewChannelTracker(bus)

	recvTime := time.Now()
	bus.Publish(eventbus.ChannelMessageReceivedEvent{
		Channel:   "telegram",
		Timestamp: recvTime,
	})

	sentTime := recvTime.Add(2 * time.Second)
	bus.Publish(eventbus.ChannelMessageSentEvent{
		Channel:   "telegram",
		Timestamp: sentTime,
	})

	snap := tracker.Snapshot()
	require.Len(t, snap, 1)
	assert.Equal(t, sentTime, snap[0].LastActivity)
	// Sent events do not increment message count.
	assert.Equal(t, 1, snap[0].MessageCount)
}

func TestChannelTracker_SentEventIgnoresUnknownChannel(t *testing.T) {
	bus := eventbus.New()
	tracker := NewChannelTracker(bus)

	// Sent event for a channel that was never received/seeded — should not create entry.
	bus.Publish(eventbus.ChannelMessageSentEvent{
		Channel:   "unknown",
		Timestamp: time.Now(),
	})

	snap := tracker.Snapshot()
	assert.Empty(t, snap)
}

func TestChannelTracker_SnapshotSorted(t *testing.T) {
	tracker := NewChannelTracker(nil)
	tracker.SeedChannel("telegram", true)
	tracker.SeedChannel("discord", true)
	tracker.SeedChannel("slack", true)

	snap := tracker.Snapshot()
	require.Len(t, snap, 3)
	assert.Equal(t, "discord", snap[0].Name)
	assert.Equal(t, "slack", snap[1].Name)
	assert.Equal(t, "telegram", snap[2].Name)
}

func TestChannelTracker_MultipleChannels(t *testing.T) {
	bus := eventbus.New()
	tracker := NewChannelTracker(bus)

	bus.Publish(eventbus.ChannelMessageReceivedEvent{
		Channel:   "telegram",
		Timestamp: time.Now(),
	})
	bus.Publish(eventbus.ChannelMessageReceivedEvent{
		Channel:   "discord",
		Timestamp: time.Now(),
	})
	bus.Publish(eventbus.ChannelMessageReceivedEvent{
		Channel:   "telegram",
		Timestamp: time.Now(),
	})

	snap := tracker.Snapshot()
	require.Len(t, snap, 2)

	// Sorted: discord, telegram.
	assert.Equal(t, "discord", snap[0].Name)
	assert.Equal(t, 1, snap[0].MessageCount)

	assert.Equal(t, "telegram", snap[1].Name)
	assert.Equal(t, 2, snap[1].MessageCount)
}
