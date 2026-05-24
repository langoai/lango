package eventbus

// EventModeChanged is published when a session's active mode changes via
// /mode, --mode, or any equivalent multi-channel command.
const EventModeChanged = "session.mode.changed"

// ModeChangedEvent carries the session key and old/new mode names.
// Subscribers (TUI, channel adapters) render this in their native format
// to keep multi-channel UX consistent.
type ModeChangedEvent struct {
	SessionKey string
	OldMode    string // "" when no mode was previously set
	NewMode    string // "" when the mode was cleared
}

// EventName implements Event.
func (e ModeChangedEvent) EventName() string { return EventModeChanged }
