package discord

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/langoai/lango/internal/logging"
)

var logger = logging.SubsystemSugar("channel.discord")

// Config holds Discord channel configuration
type Config struct {
	BotToken           string
	ApplicationID      string
	AllowedGuilds      []string     // empty = all
	ApprovalTimeoutSec int          // 0 = default 30s
	HTTPClient         *http.Client // optional, for testing
	Session            Session      // optional, for testing
}

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error)

// IncomingMessage represents a message from Discord
type IncomingMessage struct {
	MessageID  string
	ChannelID  string
	GuildID    string
	AuthorID   string
	AuthorName string
	Content    string
	IsDM       bool
	IsMention  bool
}

// OutgoingMessage represents a message to send
type OutgoingMessage struct {
	Content string
	Embed   *Embed
}

// Embed represents a Discord embed
type Embed struct {
	Title       string
	Description string
	Color       int
	Fields      []EmbedField
}

// EmbedField represents an embed field
type EmbedField struct {
	Name   string
	Value  string
	Inline bool
}

// Channel implements Discord bot
type Channel struct {
	config   Config
	session  Session
	handler  MessageHandler
	approval *ApprovalProvider
	ctx      context.Context
	botID    string
	stopChan chan struct{}
}

// New creates a new Discord channel
func New(cfg Config) (*Channel, error) {
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("bot token is required")
	}

	var sess Session
	if cfg.Session != nil {
		sess = cfg.Session
	} else {
		session, err := discordgo.New("Bot " + cfg.BotToken)
		if err != nil {
			return nil, fmt.Errorf("create session: %w", err)
		}

		session.Identify.Intents = discordgo.IntentsGuildMessages |
			discordgo.IntentsDirectMessages |
			discordgo.IntentMessageContent

		if cfg.HTTPClient != nil {
			session.Client = cfg.HTTPClient
		}
		sess = NewDiscordSession(session)
	}

	ch := &Channel{
		config:   cfg,
		session:  sess,
		stopChan: make(chan struct{}),
	}
	ch.approval = NewApprovalProvider(sess, time.Duration(cfg.ApprovalTimeoutSec)*time.Second)

	return ch, nil
}

// SetHandler sets the message handler
func (c *Channel) SetHandler(handler MessageHandler) {
	c.handler = handler
}

// GetApprovalProvider returns the channel's approval provider for composite registration.
func (c *Channel) GetApprovalProvider() *ApprovalProvider {
	return c.approval
}

// Start starts the Discord bot
func (c *Channel) Start(ctx context.Context) error {
	if c.handler == nil {
		return fmt.Errorf("message handler not set")
	}

	c.ctx = ctx

	c.session.AddHandler(c.onMessageCreate)
	c.session.AddHandler(c.onInteractionCreate)

	if err := c.session.Open(); err != nil {
		return fmt.Errorf("open session: %w", err)
	}

	c.botID = c.session.GetState().User.ID
	logger.Infow("discord bot connected", "botId", c.botID, "username", c.session.GetState().User.Username)

	// Register slash commands if application ID is set
	if c.config.ApplicationID != "" {
		c.registerCommands()
	}

	return nil
}

// onMessageCreate handles message events
func (c *Channel) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore bot's own messages
	if m.Author.ID == c.botID {
		return
	}

	// Check guild allowlist
	if m.GuildID != "" && !c.isGuildAllowed(m.GuildID) {
		return
	}

	// Check if it's a DM or mention
	isDM := m.GuildID == ""
	isMention := c.isBotMentioned(m)

	// Only respond to DMs or mentions
	if !isDM && !isMention {
		return
	}

	// Clean content (remove bot mention)
	content := c.cleanContent(m.Content)

	incoming := &IncomingMessage{
		MessageID:  m.ID,
		ChannelID:  m.ChannelID,
		GuildID:    m.GuildID,
		AuthorID:   m.Author.ID,
		AuthorName: m.Author.Username,
		Content:    content,
		IsDM:       isDM,
		IsMention:  isMention,
	}

	logger.Infow("received message",
		"messageId", m.ID,
		"channelId", m.ChannelID,
		"authorId", m.Author.ID,
	)

	// Show typing indicator while processing
	stopThinking := c.startTyping(m.ChannelID)
	response, err := c.handler(c.ctx, incoming)
	stopThinking()

	if err != nil {
		logger.Errorw("handler error", "error", err)
		c.sendError(m.ChannelID, err)
		return
	}

	if response != nil {
		if err := c.Send(m.ChannelID, response); err != nil {
			logger.Errorw("send error", "error", err)
		}
	}
}

// StartTyping sends a typing indicator to the channel and refreshes it
// periodically until the returned stop function is called or ctx is cancelled.
// The returned stop function is safe to call multiple times.
func (c *Channel) StartTyping(ctx context.Context, channelID string) func() {
	if err := c.session.ChannelTyping(channelID); err != nil {
		logger.Warnw("typing indicator error", "error", err)
	}

	done := make(chan struct{})
	var once sync.Once
	go func() {
		ticker := time.NewTicker(8 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := c.session.ChannelTyping(channelID); err != nil {
					logger.Warnw("typing indicator refresh error", "error", err)
				}
			}
		}
	}()

	return func() { once.Do(func() { close(done) }) }
}

// startTyping sends a typing indicator to the channel and refreshes it
// periodically until the returned stop function is called.
// The returned stop function is safe to call multiple times.
func (c *Channel) startTyping(channelID string) func() {
	if err := c.session.ChannelTyping(channelID); err != nil {
		logger.Warnw("typing indicator error", "error", err)
	}

	done := make(chan struct{})
	var once sync.Once
	go func() {
		ticker := time.NewTicker(8 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if err := c.session.ChannelTyping(channelID); err != nil {
					logger.Warnw("typing indicator refresh error", "error", err)
				}
			}
		}
	}()

	return func() { once.Do(func() { close(done) }) }
}

// Send sends a message
func (c *Channel) Send(channelID string, msg *OutgoingMessage) error {
	// Split long messages (Discord limit is 2000)
	chunks := splitMessage(msg.Content, 2000)

	for _, chunk := range chunks {
		if msg.Embed != nil {
			embed := &discordgo.MessageEmbed{
				Title:       msg.Embed.Title,
				Description: msg.Embed.Description,
				Color:       msg.Embed.Color,
			}
			for _, f := range msg.Embed.Fields {
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:   f.Name,
					Value:  f.Value,
					Inline: f.Inline,
				})
			}
			_, err := c.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
				Content: chunk,
				Embed:   embed,
			})
			if err != nil {
				return err
			}
		} else {
			_, err := c.session.ChannelMessageSend(channelID, chunk)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// registerCommands registers slash commands
func (c *Channel) registerCommands() {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "ask",
			Description: "Ask the AI assistant",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "Your message",
					Required:    true,
				},
			},
		},
		{
			Name:        "clear",
			Description: "Clear conversation history",
		},
	}

	for _, cmd := range commands {
		_, err := c.session.ApplicationCommandCreate(c.config.ApplicationID, "", cmd)
		if err != nil {
			logger.Warnw("create command", "command", cmd.Name, "error", err)
		}
	}

	logger.Info("slash commands registered")
}

// onInteractionCreate handles interaction events (button clicks)
func (c *Channel) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionMessageComponent {
		c.approval.HandleInteraction(i)
	}
}

// isBotMentioned checks if the bot is mentioned
func (c *Channel) isBotMentioned(m *discordgo.MessageCreate) bool {
	for _, mention := range m.Mentions {
		if mention.ID == c.botID {
			return true
		}
	}
	return false
}

// cleanContent removes bot mention from content
func (c *Channel) cleanContent(content string) string {
	// Remove <@botID> or <@!botID> mentions
	content = strings.ReplaceAll(content, "<@"+c.botID+">", "")
	content = strings.ReplaceAll(content, "<@!"+c.botID+">", "")
	return strings.TrimSpace(content)
}

// isGuildAllowed checks if a guild is in the allowlist
func (c *Channel) isGuildAllowed(guildID string) bool {
	if len(c.config.AllowedGuilds) == 0 {
		return true
	}
	for _, id := range c.config.AllowedGuilds {
		if id == guildID {
			return true
		}
	}
	return false
}

// sendError sends an error message
func (c *Channel) sendError(channelID string, err error) {
	_, _ = c.session.ChannelMessageSend(channelID, fmt.Sprintf("❌ Error: %s", err.Error()))
}

// splitMessage splits a message into chunks
func splitMessage(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	var chunks []string
	for len(text) > maxLen {
		// Find last newline before limit
		idx := strings.LastIndex(text[:maxLen], "\n")
		if idx == -1 {
			idx = maxLen
		}
		chunks = append(chunks, text[:idx])
		text = text[idx:]
		if len(text) > 0 && text[0] == '\n' {
			text = text[1:]
		}
	}
	if len(text) > 0 {
		chunks = append(chunks, text)
	}
	return chunks
}

// Stop stops the Discord bot
func (c *Channel) Stop() {
	c.session.Close()
	logger.Info("discord channel stopped")
}
