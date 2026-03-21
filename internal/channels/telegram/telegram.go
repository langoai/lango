package telegram

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/langoai/lango/internal/logging"
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
	stopOnce sync.Once
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
			case update, ok := <-updates:
				if !ok {
					return
				}
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

	// Post a "Thinking..." placeholder and start progress updates.
	thinkingMsg, thinkingErr := c.postThinking(incoming.ChatID)
	var stopProgress func()
	if thinkingErr == nil {
		stopProgress = c.startProgressUpdates(incoming.ChatID, thinkingMsg.MessageID)
	} else {
		// Fall back to typing indicator if posting failed.
		stopFallback := c.startTyping(incoming.ChatID)
		stopProgress = stopFallback
	}

	response, err := c.handler(ctx, incoming)
	stopProgress()

	if err != nil {
		logger().Errorw("handler error", "error", err)
		// Update placeholder with error message if possible.
		if thinkingErr == nil {
			errText := fmt.Sprintf("❌ %s", formatChannelError(err))
			c.editMessage(incoming.ChatID, thinkingMsg.MessageID, errText)
		} else {
			c.sendError(incoming.ChatID, msg.MessageID, err)
		}
		return
	}

	if response != nil && response.Text != "" {
		// Replace placeholder with actual response.
		if thinkingErr == nil {
			c.editMessage(incoming.ChatID, thinkingMsg.MessageID, response.Text)
		} else {
			if err := c.Send(incoming.ChatID, response); err != nil {
				logger().Errorw("send error", "error", err)
			}
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
// The returned stop function is safe to call multiple times.
func (c *Channel) startTyping(chatID int64) func() {
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
			case <-ticker.C:
				if _, err := c.bot.Request(action); err != nil {
					logger().Warnw("typing indicator refresh error", "error", err)
				}
			}
		}
	}()

	return func() { once.Do(func() { close(done) }) }
}

// postThinking sends a "Thinking..." placeholder message and returns the sent message.
func (c *Channel) postThinking(chatID int64) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, "_Thinking..._")
	msg.ParseMode = "Markdown"
	return c.bot.Send(msg)
}

// editMessage edits an existing message with new text.
func (c *Channel) editMessage(chatID int64, messageID int, text string) {
	formatted := FormatMarkdown(text)
	edit := tgbotapi.NewEditMessageText(chatID, messageID, formatted)
	edit.ParseMode = "Markdown"
	if _, err := c.bot.Send(edit); err != nil {
		// Retry as plain text if Markdown fails.
		plainEdit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		if _, retryErr := c.bot.Send(plainEdit); retryErr != nil {
			logger().Warnw("edit message failed", "error", retryErr)
		}
	}
}

// startProgressUpdates periodically edits the thinking placeholder with elapsed time.
// Returns a stop function that must be called before the placeholder is replaced.
func (c *Channel) startProgressUpdates(chatID int64, messageID int) func() {
	start := time.Now()
	done := make(chan struct{})
	var once sync.Once

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				elapsed := time.Since(start).Truncate(time.Second)
				text := fmt.Sprintf("_Thinking... (%s)_", elapsed)
				edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
				edit.ParseMode = "Markdown"
				if _, err := c.bot.Send(edit); err != nil {
					logger().Warnw("progress update error", "error", err)
				}
			}
		}
	}()

	return func() { once.Do(func() { close(done) }) }
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

// sendError sends an error message with user-friendly formatting.
func (c *Channel) sendError(chatID int64, replyTo int, err error) {
	_ = c.Send(chatID, &OutgoingMessage{
		Text:      fmt.Sprintf("❌ %s", formatChannelError(err)),
		ReplyToID: replyTo,
	})
}

// formatChannelError returns a user-friendly error message.
// If the error implements UserMessage(), that is used; otherwise falls back to Error().
func formatChannelError(err error) string {
	type userMessager interface {
		UserMessage() string
	}
	var um userMessager
	if errors.As(err, &um) {
		return um.UserMessage()
	}
	return fmt.Sprintf("Error: %s", err.Error())
}

// downloadTimeout is the maximum time allowed for downloading a file.
const downloadTimeout = 30 * time.Second

// DownloadFile downloads a file by file ID
func (c *Channel) DownloadFile(fileID string) ([]byte, error) {
	file, err := c.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return nil, fmt.Errorf("get file: %w", err)
	}

	fileURL := file.Link(c.config.BotToken)

	ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create download request: %w", err)
	}

	client := c.config.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download file: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read file body: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("download file: empty response body")
	}

	return data, nil
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

// Stop stops the channel.
func (c *Channel) Stop(ctx context.Context) error {
	c.stopOnce.Do(func() {
		close(c.stopChan)
		c.bot.StopReceivingUpdates()
	})

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}

	logger().Info("telegram channel stopped")
	return nil
}
