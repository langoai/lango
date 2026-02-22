package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/langoai/lango/internal/approval"
)

// approvalPending holds the response channel and message metadata for a pending approval.
type approvalPending struct {
	ch        chan approval.ApprovalResponse
	chatID    int64
	messageID int
}

// ApprovalProvider implements approval.Provider for Telegram using InlineKeyboard buttons.
type ApprovalProvider struct {
	bot     BotAPI
	pending sync.Map // map[requestID]*approvalPending
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
func (p *ApprovalProvider) RequestApproval(ctx context.Context, req approval.ApprovalRequest) (approval.ApprovalResponse, error) {
	chatID, err := parseTelegramChatID(req.SessionKey)
	if err != nil {
		return approval.ApprovalResponse{}, fmt.Errorf("parse session key: %w", err)
	}

	respChan := make(chan approval.ApprovalResponse, 1)

	// Build inline keyboard with Always Allow on a second row
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Approve", "approve:"+req.ID),
			tgbotapi.NewInlineKeyboardButtonData("❌ Deny", "deny:"+req.ID),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔓 Always Allow", "always:"+req.ID),
		),
	)

	text := fmt.Sprintf("🔐 Tool '%s' requires approval", req.ToolName)
	if req.Summary != "" {
		text += "\n\n" + req.Summary
	}
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	sentMsg, err := p.bot.Send(msg)
	if err != nil {
		return approval.ApprovalResponse{}, fmt.Errorf("send approval message: %w", err)
	}

	p.pending.Store(req.ID, &approvalPending{
		ch:        respChan,
		chatID:    chatID,
		messageID: sentMsg.MessageID,
	})
	defer p.pending.Delete(req.ID)

	select {
	case resp := <-respChan:
		return resp, nil
	case <-ctx.Done():
		p.editApprovalMessage(chatID, sentMsg.MessageID, "🔐 Tool approval — ⏱ Expired")
		return approval.ApprovalResponse{}, ctx.Err()
	case <-time.After(p.timeout):
		p.editApprovalMessage(chatID, sentMsg.MessageID, "🔐 Tool approval — ⏱ Expired")
		return approval.ApprovalResponse{}, fmt.Errorf("approval timeout")
	}
}

// HandleCallback processes an InlineKeyboard callback query for approval.
func (p *ApprovalProvider) HandleCallback(query *tgbotapi.CallbackQuery) {
	if query == nil || query.Data == "" {
		return
	}

	var requestID string
	var resp approval.ApprovalResponse

	if strings.HasPrefix(query.Data, "approve:") {
		requestID = strings.TrimPrefix(query.Data, "approve:")
		resp = approval.ApprovalResponse{Approved: true}
	} else if strings.HasPrefix(query.Data, "deny:") {
		requestID = strings.TrimPrefix(query.Data, "deny:")
		resp = approval.ApprovalResponse{}
	} else if strings.HasPrefix(query.Data, "always:") {
		requestID = strings.TrimPrefix(query.Data, "always:")
		resp = approval.ApprovalResponse{Approved: true, AlwaysAllow: true}
	} else {
		return
	}

	// LoadAndDelete first to prevent duplicate clicks (TOCTOU)
	val, ok := p.pending.LoadAndDelete(requestID)
	if !ok {
		// Already processed or expired — answer callback silently
		callback := tgbotapi.NewCallback(query.ID, "")
		if _, err := p.bot.Request(callback); err != nil {
			if !isCallbackExpiredErr(err) {
				logger().Debugw("answer expired callback error", "error", err)
			}
		}
		return
	}

	pending, ok := val.(*approvalPending)
	if !ok {
		logger().Warnw("unexpected pending type", "requestId", requestID)
		return
	}

	// Answer callback to dismiss the loading indicator
	callback := tgbotapi.NewCallback(query.ID, "")
	if _, err := p.bot.Request(callback); err != nil {
		if !isCallbackExpiredErr(err) {
			logger().Debugw("answer callback error", "error", err)
		}
	}

	// Unblock the waiting agent immediately before editing the message.
	// This prevents Telegram API latency from delaying the agent's next action
	// and avoids message ordering issues where the final response arrives before
	// the approval status edit.
	select {
	case pending.ch <- resp:
	default:
	}

	// Edit original message to remove the keyboard
	var status string
	switch {
	case resp.AlwaysAllow:
		status = "🔓 Always Allowed"
	case resp.Approved:
		status = "✅ Approved"
	default:
		status = "❌ Denied"
	}
	p.editApprovalMessage(pending.chatID, pending.messageID, fmt.Sprintf("🔐 Tool approval — %s", status))
}

// CanHandle returns true for session keys starting with "telegram:".
func (p *ApprovalProvider) CanHandle(sessionKey string) bool {
	return strings.HasPrefix(sessionKey, "telegram:")
}

// editApprovalMessage edits a message with new text and removes inline keyboard buttons.
func (p *ApprovalProvider) editApprovalMessage(chatID int64, messageID int, newText string) {
	emptyMarkup := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
	}
	edit := tgbotapi.NewEditMessageTextAndMarkup(chatID, messageID, newText, emptyMarkup)
	if _, err := p.bot.Send(edit); err != nil {
		if !isMessageNotModifiedErr(err) {
			logger().Warnw("edit approval message error", "error", err)
		}
	}
}

func isCallbackExpiredErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "query is too old")
}

func isMessageNotModifiedErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "message is not modified")
}

// parseTelegramChatID extracts the chatID from a session key like "telegram:<chatID>:<userID>".
func parseTelegramChatID(sessionKey string) (int64, error) {
	parts := strings.SplitN(sessionKey, ":", 3)
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid telegram session key: %s", sessionKey)
	}
	return strconv.ParseInt(parts[1], 10, 64)
}
