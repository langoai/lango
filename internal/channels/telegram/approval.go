package telegram

import (
	"context"
	"fmt"
	"strings"
	"strconv"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/langowarny/lango/internal/approval"
)

// ApprovalProvider implements approval.Provider for Telegram using InlineKeyboard buttons.
type ApprovalProvider struct {
	bot     BotAPI
	pending sync.Map // map[requestID]chan bool
	timeout time.Duration
}

var _ approval.Provider = (*ApprovalProvider)(nil)

// NewApprovalProvider creates a Telegram approval provider.
func NewApprovalProvider(bot BotAPI, timeout time.Duration) *ApprovalProvider {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &ApprovalProvider{
		bot:     bot,
		timeout: timeout,
	}
}

// RequestApproval sends an InlineKeyboard message to the chat and waits for a callback.
func (p *ApprovalProvider) RequestApproval(ctx context.Context, req approval.ApprovalRequest) (bool, error) {
	chatID, err := parseTelegramChatID(req.SessionKey)
	if err != nil {
		return false, fmt.Errorf("parse session key: %w", err)
	}

	respChan := make(chan bool, 1)
	p.pending.Store(req.ID, respChan)
	defer p.pending.Delete(req.ID)

	// Build inline keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Approve", "approve:"+req.ID),
			tgbotapi.NewInlineKeyboardButtonData("❌ Deny", "deny:"+req.ID),
		),
	)

	text := fmt.Sprintf("🔐 Tool '%s' requires approval", req.ToolName)
	if req.Summary != "" {
		text += "\n\n" + req.Summary
	}
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	if _, err := p.bot.Send(msg); err != nil {
		return false, fmt.Errorf("send approval message: %w", err)
	}

	select {
	case approved := <-respChan:
		return approved, nil
	case <-ctx.Done():
		return false, ctx.Err()
	case <-time.After(p.timeout):
		return false, fmt.Errorf("approval timeout")
	}
}

// HandleCallback processes an InlineKeyboard callback query for approval.
func (p *ApprovalProvider) HandleCallback(query *tgbotapi.CallbackQuery) {
	if query == nil || query.Data == "" {
		return
	}

	var requestID string
	var approved bool

	if strings.HasPrefix(query.Data, "approve:") {
		requestID = strings.TrimPrefix(query.Data, "approve:")
		approved = true
	} else if strings.HasPrefix(query.Data, "deny:") {
		requestID = strings.TrimPrefix(query.Data, "deny:")
		approved = false
	} else {
		return
	}

	// Answer callback to dismiss the loading indicator
	callback := tgbotapi.NewCallback(query.ID, "")
	if _, err := p.bot.Request(callback); err != nil {
		logger().Warnw("answer callback error", "error", err)
	}

	// Edit original message to remove the keyboard
	if query.Message != nil {
		status := "❌ Denied"
		if approved {
			status = "✅ Approved"
		}
		edit := tgbotapi.NewEditMessageText(
			query.Message.Chat.ID,
			query.Message.MessageID,
			fmt.Sprintf("%s — %s", query.Message.Text, status),
		)
		if _, err := p.bot.Send(edit); err != nil {
			logger().Warnw("edit approval message error", "error", err)
		}
	}

	// Send result to waiting goroutine
	if ch, ok := p.pending.LoadAndDelete(requestID); ok {
		respChan, ok := ch.(chan bool)
		if !ok {
			logger().Warnw("unexpected pending type", "requestId", requestID)
			return
		}
		select {
		case respChan <- approved:
		default:
		}
	}
}

// CanHandle returns true for session keys starting with "telegram:".
func (p *ApprovalProvider) CanHandle(sessionKey string) bool {
	return strings.HasPrefix(sessionKey, "telegram:")
}

// parseTelegramChatID extracts the chatID from a session key like "telegram:<chatID>:<userID>".
func parseTelegramChatID(sessionKey string) (int64, error) {
	parts := strings.SplitN(sessionKey, ":", 3)
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid telegram session key: %s", sessionKey)
	}
	return strconv.ParseInt(parts[1], 10, 64)
}
