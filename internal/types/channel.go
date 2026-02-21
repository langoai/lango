package types

// ChannelType represents a messaging platform channel type.
type ChannelType string

const (
	ChannelTelegram ChannelType = "telegram"
	ChannelDiscord  ChannelType = "discord"
	ChannelSlack    ChannelType = "slack"
)

// Valid reports whether c is a known channel type.
func (c ChannelType) Valid() bool {
	switch c {
	case ChannelTelegram, ChannelDiscord, ChannelSlack:
		return true
	}
	return false
}

// Values returns all known channel types.
func (c ChannelType) Values() []ChannelType {
	return []ChannelType{ChannelTelegram, ChannelDiscord, ChannelSlack}
}
