package discord

import "github.com/bwmarrin/discordgo"

// Session defines the interface for Discord session operations.
type Session interface {
	Open() error
	Close() error
	AddHandler(handler interface{}) func()
	ChannelMessageSend(channelID string, content string, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ChannelMessageEditComplex(edit *discordgo.MessageEdit, options ...discordgo.RequestOption) (*discordgo.Message, error)
	InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error
	ApplicationCommandCreate(appID string, guildID string, cmd *discordgo.ApplicationCommand, options ...discordgo.RequestOption) (*discordgo.ApplicationCommand, error)
	GetState() *discordgo.State
}

// DiscordSession is an adapter for *discordgo.Session that implements Session.
type DiscordSession struct {
	*discordgo.Session
}

var _ Session = (*DiscordSession)(nil)

// NewDiscordSession creates a new DiscordSession adapter.
func NewDiscordSession(s *discordgo.Session) *DiscordSession {
	return &DiscordSession{Session: s}
}

// GetState returns the session state.
func (s *DiscordSession) GetState() *discordgo.State {
	return s.State
}
