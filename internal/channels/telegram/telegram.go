package telegram

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/langowarny/lango/internal/logging"
)

func logger() *zap.SugaredLogger { return logging.Channel().Named("telegram") }

// Config holds Telegram channel configuration
type Config struct {
	BotToken           string
	Allowlist          []int64      // allowed user/chat IDs (empty = all)
	ApprovalTimeoutSec int          // 0 = default 30s
	APIEndpoint        string       // optional, for testing
	HTTPClient         *http.Client // optional, for testing
	Bot                BotAPI       // optional, for testing
}

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error)

// IncomingMessage represents a message from Telegram
type IncomingMessage struct {
	MessageID   int
	ChatID      int64
	UserID      int64
	Username    string
	Text        string
	ReplyToID   int
	HasMedia    bool
	MediaType   string
	MediaFileID string
}

// OutgoingMessage represents a message to send
type OutgoingMessage struct {
	Text           string
	ReplyToID      int
	ParseMode      string // "Markdown", "HTML"
	DisablePreview bool
}

// Channel implements Telegram bot
type Channel struct {
	config   Config
	bot      BotAPI
	handler  MessageHandler
	approval *ApprovalProvider
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// New creates a new Telegram channel
func New(cfg Config) (*Channel, error) {
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("bot token is required")
	}

	endpoint := cfg.APIEndpoint
	if endpoint == "" {
		endpoint = tgbotapi.APIEndpoint
	}

	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{}
	}

	var botAPI BotAPI
	if cfg.Bot != nil {
		botAPI = cfg.Bot
	} else {
		bot, err := tgbotapi.NewBotAPIWithClient(cfg.BotToken, endpoint, client)
		if err != nil {
			return nil, fmt.Errorf("create bot: %w", err)
		}
		botAPI = NewTelegramBot(bot)
	}

	logger().Infow("telegram bot authorized", "username", botAPI.GetSelf().UserName)

	ch := &Channel{
		config:   cfg,
		bot:      botAPI,
		stopChan: make(chan struct{}),
	}
	ch.approval = NewApprovalProvider(botAPI, time.Duration(cfg.ApprovalTimeoutSec)*time.Second)

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

// Start starts listening for updates
func (c *Channel) Start(ctx context.Context) error {
	if c.handler == nil {
		return fmt.Errorf("message handler not set")
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := c.bot.GetUpdatesChan(u)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case <-c.stopChan:
				return
			case update := <-updates:
				if update.CallbackQuery != nil {
					c.approval.HandleCallback(update.CallbackQuery)
					continue
				}

				if update.Message == nil {
					continue
				}

				// Check allowlist
				if !c.isAllowed(update.Message.Chat.ID, update.Message.From.ID) {
					logger().Warnw("blocked message from non-allowed user",
						"userId", update.Message.From.ID,
						"chatId", update.Message.Chat.ID,
					)
					continue
				}

				c.wg.Add(1)
				go func() {
					defer c.wg.Done()
					c.handleUpdate(ctx, update)
				}()
			}
		}
	}()

	logger().Infow("telegram channel started", "bot", c.bot.GetSelf().UserName)
	return nil
}

// handleUpdate processes a single update
func (c *Channel) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	msg := update.Message

	incoming := &IncomingMessage{
		MessageID: msg.MessageID,
		ChatID:    msg.Chat.ID,
		UserID:    msg.From.ID,
		Username:  msg.From.UserName,
		Text:      msg.Text,
	}

	if msg.ReplyToMessage != nil {
		incoming.ReplyToID = msg.ReplyToMessage.MessageID
	}

	// Check for media
	if len(msg.Photo) > 0 {
		incoming.HasMedia = true
		incoming.MediaType = "photo"
		incoming.MediaFileID = msg.Photo[len(msg.Photo)-1].FileID
	} else if msg.Document != nil {
		incoming.HasMedia = true
		incoming.MediaType = "document"
		incoming.MediaFileID = msg.Document.FileID
	} else if msg.Voice != nil {
		incoming.HasMedia = true
		incoming.MediaType = "voice"
		incoming.MediaFileID = msg.Voice.FileID
	}

	logger().Infow("received message",
		"messageId", incoming.MessageID,
		"chatId", incoming.ChatID,
		"userId", incoming.UserID,
	)

	// Show typing indicator while processing
	stopThinking := c.startTyping(incoming.ChatID)
	response, err := c.handler(ctx, incoming)
	stopThinking()

	if err != nil {
		logger().Errorw("handler error", "error", err)
		c.sendError(incoming.ChatID, msg.MessageID, err)
		return
	}

	if response != nil {
		if err := c.Send(incoming.ChatID, response); err != nil {
			logger().Errorw("send error", "error", err)
		}
	}
}

// StartTyping sends a typing indicator to the chat and refreshes it
// periodically until the returned stop function is called or ctx is cancelled.
// The returned stop function is safe to call multiple times.
func (c *Channel) StartTyping(ctx context.Context, chatID int64) func() {
	action := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	if _, err := c.bot.Request(action); err != nil {
		logger().Warnw("typing indicator error", "error", err)
	}

	done := make(chan struct{})
	var once sync.Once
	go func() {
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := c.bot.Request(action); err != nil {
					logger().Warnw("typing indicator refresh error", "error", err)
				}
			}
		}
	}()

	return func() { once.Do(func() { close(done) }) }
}

// startTyping sends a typing action to the chat and refreshes it
// periodically until the returned stop function is called.
func (c *Channel) startTyping(chatID int64) func() {
	action := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	if _, err := c.bot.Request(action); err != nil {
		logger().Warnw("typing indicator error", "error", err)
	}

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if _, err := c.bot.Request(action); err != nil {
					logger().Warnw("typing indicator refresh error", "error", err)
				}
			}
		}
	}()

	return func() { close(done) }
}

// Send sends a message.
// When ParseMode is not set, standard Markdown is auto-converted to Telegram v1
// and sent with ParseMode "Markdown". If the API rejects the formatted text,
// the original text is re-sent as plain text.
func (c *Channel) Send(chatID int64, msg *OutgoingMessage) error {
	text := msg.Text
	parseMode := msg.ParseMode

	// Auto-format: standard Markdown → Telegram v1
	if parseMode == "" {
		text = FormatMarkdown(msg.Text)
		parseMode = "Markdown"
	}

	// Split long messages (Telegram limit is 4096)
	chunks := c.splitMessage(text, 4096)

	for i, chunk := range chunks {
		tgMsg := tgbotapi.NewMessage(chatID, chunk)

		if i == 0 && msg.ReplyToID > 0 {
			tgMsg.ReplyToMessageID = msg.ReplyToID
		}

		tgMsg.ParseMode = parseMode
		tgMsg.DisableWebPagePreview = msg.DisablePreview

		if _, err := c.bot.Send(tgMsg); err != nil {
			// Fallback: re-send as plain text if Markdown parsing failed
			logger().Warnw("markdown send failed, retrying as plain text", "error", err)
			if fallbackErr := c.sendPlainText(chatID, msg, i); fallbackErr != nil {
				return fmt.Errorf("send plain text fallback: %w", fallbackErr)
			}
			return nil
		}
	}

	return nil
}

// sendPlainText re-sends the original message text without any parse mode,
// starting from the given chunk index.
func (c *Channel) sendPlainText(chatID int64, msg *OutgoingMessage, fromChunk int) error {
	chunks := c.splitMessage(msg.Text, 4096)

	for i := fromChunk; i < len(chunks); i++ {
		tgMsg := tgbotapi.NewMessage(chatID, chunks[i])

		if i == 0 && msg.ReplyToID > 0 {
			tgMsg.ReplyToMessageID = msg.ReplyToID
		}

		tgMsg.DisableWebPagePreview = msg.DisablePreview

		if _, err := c.bot.Send(tgMsg); err != nil {
			return fmt.Errorf("send chunk %d: %w", i, err)
		}
	}

	return nil
}

// splitMessage splits a message into chunks
func (c *Channel) splitMessage(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	var chunks []string
	lines := strings.Split(text, "\n")
	var current strings.Builder

	for _, line := range lines {
		if current.Len()+len(line)+1 > maxLen {
			if current.Len() > 0 {
				chunks = append(chunks, current.String())
				current.Reset()
			}
			// Handle very long lines
			for len(line) > maxLen {
				chunks = append(chunks, line[:maxLen])
				line = line[maxLen:]
			}
		}
		if current.Len() > 0 {
			current.WriteString("\n")
		}
		current.WriteString(line)
	}

	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}

	return chunks
}

// sendError sends an error message
func (c *Channel) sendError(chatID int64, replyTo int, err error) {
	c.Send(chatID, &OutgoingMessage{
		Text:      fmt.Sprintf("❌ Error: %s", err.Error()),
		ReplyToID: replyTo,
	})
}

// DownloadFile downloads a file by file ID
func (c *Channel) DownloadFile(fileID string) ([]byte, error) {
	file, err := c.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return nil, fmt.Errorf("get file: %w", err)
	}

	url := file.Link(c.config.BotToken)
	_ = url // Would download from URL

	// Note: actual download implementation would fetch from url
	return nil, fmt.Errorf("download not implemented")
}

// isAllowed checks if a user/chat is allowed
func (c *Channel) isAllowed(chatID, userID int64) bool {
	if len(c.config.Allowlist) == 0 {
		return true
	}

	for _, id := range c.config.Allowlist {
		if id == chatID || id == userID {
			return true
		}
	}

	return false
}

// Stop stops the channel
func (c *Channel) Stop() {
	close(c.stopChan)
	c.wg.Wait()
	c.bot.StopReceivingUpdates()
	logger().Info("telegram channel stopped")
}
