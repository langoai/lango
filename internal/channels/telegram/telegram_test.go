package telegram

import (
	"context"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MockBotAPI implements BotAPI interface
type MockBotAPI struct {
	GetUpdatesChanFunc func(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
	SendFunc           func(c tgbotapi.Chattable) (tgbotapi.Message, error)
	GetSelfFunc        func() tgbotapi.User
	SentMessages       []tgbotapi.Chattable
}

func (m *MockBotAPI) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	if m.GetUpdatesChanFunc != nil {
		return m.GetUpdatesChanFunc(config)
	}
	ch := make(chan tgbotapi.Update)
	return ch
}

func (m *MockBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.SentMessages = append(m.SentMessages, c)
	if m.SendFunc != nil {
		return m.SendFunc(c)
	}
	return tgbotapi.Message{MessageID: 101}, nil
}

func (m *MockBotAPI) GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error) {
	return tgbotapi.File{}, nil
}

func (m *MockBotAPI) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return &tgbotapi.APIResponse{Ok: true}, nil
}

func (m *MockBotAPI) StopReceivingUpdates() {
}

func (m *MockBotAPI) GetSelf() tgbotapi.User {
	if m.GetSelfFunc != nil {
		return m.GetSelfFunc()
	}
	return tgbotapi.User{ID: 12345, UserName: "TestBot"}
}

func TestTelegramChannel(t *testing.T) {
	updatesCh := make(chan tgbotapi.Update, 1)

	mockBot := &MockBotAPI{
		GetUpdatesChanFunc: func(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
			return updatesCh
		},
	}

	cfg := Config{
		BotToken: "TEST_TOKEN",
		Bot:      mockBot,
	}

	channel, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}

	msgProcessed := make(chan bool)

	channel.SetHandler(func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error) {
		if msg.Text != "Hello Bot" {
			t.Errorf("expected 'Hello Bot', got '%s'", msg.Text)
		}
		if msg.UserID != 999 {
			t.Errorf("expected user ID 999, got %d", msg.UserID)
		}
		msgProcessed <- true
		return &OutgoingMessage{Text: "Reply"}, nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := channel.Start(ctx); err != nil {
		t.Fatalf("failed to start channel: %v", err)
	}
	defer channel.Stop()

	// Simulate incoming message
	updatesCh <- tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			MessageID: 100,
			From: &tgbotapi.User{
				ID:       999,
				UserName: "user",
			},
			Chat: &tgbotapi.Chat{
				ID:   999,
				Type: "private",
			},
			Text: "Hello Bot",
		},
	}

	select {
	case <-msgProcessed:
		// Check response
		if len(mockBot.SentMessages) == 0 {
			t.Error("expected Send to be called")
		} else {
			sent := mockBot.SentMessages[0].(tgbotapi.MessageConfig)
			if sent.Text != "Reply" {
				t.Errorf("expected 'Reply', got '%s'", sent.Text)
			}
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for message processing")
	}
}
