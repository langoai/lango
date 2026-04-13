package cockpit

import (
	"sort"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/eventbus"
)

// msgSender abstracts tea.Program.Send for testability.
// *tea.Program satisfies this interface via duck typing.
type msgSender interface {
	Send(msg tea.Msg)
}

// SubscribeChannelEvents subscribes to channel events on the EventBus
// and forwards them as tea.Msg to the TUI program.
func SubscribeChannelEvents(bus *eventbus.Bus, sender msgSender) {
	if bus == nil || sender == nil {
		return
	}

	eventbus.SubscribeTyped(bus, func(e eventbus.ChannelMessageReceivedEvent) {
		sender.Send(chat.ChannelMessageMsg{
			Channel:    e.Channel,
			SessionKey: e.SessionKey,
			SenderName: e.SenderName,
			Text:       e.Text,
			Timestamp:  e.Timestamp,
			Metadata:   e.Metadata,
		})
	})
}

// ChannelTracker aggregates channel status from EventBus events.
// It is safe for concurrent use.
type ChannelTracker struct {
	mu       sync.RWMutex
	channels map[string]*channelStatusEntry
	bus      *eventbus.Bus
}

type channelStatusEntry struct {
	name         string
	connected    bool
	messageCount int
	lastActivity time.Time
}

// NewChannelTracker creates a tracker and subscribes to channel events.
// If bus is nil, the tracker still works for manual seeding but receives
// no events.
func NewChannelTracker(bus *eventbus.Bus) *ChannelTracker {
	t := &ChannelTracker{
		channels: make(map[string]*channelStatusEntry),
		bus:      bus,
	}
	if bus != nil {
		eventbus.SubscribeTyped(bus, func(e eventbus.ChannelMessageReceivedEvent) {
			t.mu.Lock()
			defer t.mu.Unlock()
			entry, ok := t.channels[e.Channel]
			if !ok {
				entry = &channelStatusEntry{name: e.Channel, connected: true}
				t.channels[e.Channel] = entry
			}
			entry.messageCount++
			entry.lastActivity = e.Timestamp
		})
		eventbus.SubscribeTyped(bus, func(e eventbus.ChannelMessageSentEvent) {
			t.mu.Lock()
			defer t.mu.Unlock()
			if entry, ok := t.channels[e.Channel]; ok {
				entry.lastActivity = e.Timestamp
			}
		})
	}
	return t
}

// SeedChannel registers a channel's initial connection status.
// Called after channel Start() to distinguish "connected but no messages yet"
// from "start failed".
func (t *ChannelTracker) SeedChannel(name string, connected bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if entry, ok := t.channels[name]; ok {
		entry.connected = connected
	} else {
		t.channels[name] = &channelStatusEntry{
			name:      name,
			connected: connected,
		}
	}
}

// Snapshot returns current channel statuses sorted by name.
// The returned slice matches the channelStatus type expected by
// ContextPanel.SetChannelStatuses.
func (t *ChannelTracker) Snapshot() []channelStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.channels) == 0 {
		return nil
	}

	result := make([]channelStatus, 0, len(t.channels))
	for _, entry := range t.channels {
		result = append(result, channelStatus{
			Name:         entry.name,
			Connected:    entry.connected,
			MessageCount: entry.messageCount,
			LastActivity: entry.lastActivity,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}
