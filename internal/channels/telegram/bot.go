package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BotAPI defines the interface for Telegram Bot API operations.
type BotAPI interface {
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error)
	StopReceivingUpdates()
	GetSelf() tgbotapi.User
}

// TelegramBot is an adapter for *tgbotapi.BotAPI.
type TelegramBot struct {
	*tgbotapi.BotAPI
}

var _ BotAPI = (*TelegramBot)(nil)

// NewTelegramBot creates a new TelegramBot adapter.
func NewTelegramBot(b *tgbotapi.BotAPI) *TelegramBot {
	return &TelegramBot{BotAPI: b}
}

// GetSelf returns the bot user.
func (b *TelegramBot) GetSelf() tgbotapi.User {
	return b.Self
}
