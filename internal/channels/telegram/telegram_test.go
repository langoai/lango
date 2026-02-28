package telegram

import (
	"context"
	"sync"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MockBotAPI implements BotAPI interface
type MockBotAPI struct {
	mu                 sync.Mutex
	GetUpdatesChanFunc func(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
	SendFunc           func(c tgbotapi.Chattable) (tgbotapi.Message, error)
	GetSelfFunc        func() tgbotapi.User
	SentMessages       []tgbotapi.Chattable
	RequestCalls       []tgbotapi.Chattable
}

func (m *MockBotAPI) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	if m.GetUpdatesChanFunc != nil {
		return m.GetUpdatesChanFunc(config)
	}
	ch := make(chan tgbotapi.Update)
	return ch
}

func (m *MockBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.mu.Lock()
	m.SentMessages = append(m.SentMessages, c)
	m.mu.Unlock()
	if m.SendFunc != nil {
		return m.SendFunc(c)
	}
	return tgbotapi.Message{MessageID: 101}, nil
}

func (m *MockBotAPI) GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error) {
	return tgbotapi.File{}, nil
}

func (m *MockBotAPI) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	m.mu.Lock()
	m.RequestCalls = append(m.RequestCalls, c)
	m.mu.Unlock()
	return &tgbotapi.APIResponse{Ok: true}, nil
}

func (m *MockBotAPI) getSentMessages() []tgbotapi.Chattable {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]tgbotapi.Chattable, len(m.SentMessages))
	copy(result, m.SentMessages)
	return result
}

func (m *MockBotAPI) getRequestCalls() []tgbotapi.Chattable {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]tgbotapi.Chattable, len(m.RequestCalls))
	copy(result, m.RequestCalls)
	return result
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
		// Allow goroutine to finish posting
		time.Sleep(50 * time.Millisecond)

		// Check typing indicator was sent via Request
		reqCalls := mockBot.getRequestCalls()
		if len(reqCalls) == 0 {
			t.Error("expected typing indicator via Request")
		} else {
			action, ok := reqCalls[0].(tgbotapi.ChatActionConfig)
			if !ok {
				t.Errorf("expected ChatActionConfig, got %T", reqCalls[0])
			} else if action.Action != tgbotapi.ChatTyping {
				t.Errorf("expected action 'typing', got '%s'", action.Action)
			}
		}

		// Check response
		sentMsgs := mockBot.getSentMessages()
		if len(sentMsgs) == 0 {
			t.Error("expected Send to be called")
		} else {
			sent := sentMsgs[0].(tgbotapi.MessageConfig)
			if sent.Text != "Reply" {
				t.Errorf("expected 'Reply', got '%s'", sent.Text)
			}
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for message processing")
	}
}

func TestTelegramTypingIndicator(t *testing.T) {
	updatesCh := make(chan tgbotapi.Update, 1)

	mockBot := &MockBotAPI{
		GetUpdatesChanFunc: func(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
			return updatesCh
		},
	}

	cfg := Config{BotToken: "TEST_TOKEN", Bot: mockBot}
	channel, err := New(cfg)
	if err != nil {
		t.Fatalf("new channel: %v", err)
	}

	done := make(chan struct{})
	channel.SetHandler(func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error) {
		close(done)
		return &OutgoingMessage{Text: "ok"}, nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := channel.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer channel.Stop()

	updatesCh <- tgbotapi.Update{
		UpdateID: 2,
		Message: &tgbotapi.Message{
			MessageID: 200,
			From:      &tgbotapi.User{ID: 888, UserName: "tester"},
			Chat:      &tgbotapi.Chat{ID: 888, Type: "private"},
			Text:      "ping",
		},
	}

	select {
	case <-done:
		// Allow goroutine to finish posting
		time.Sleep(50 * time.Millisecond)

		// Verify at least one Request call with ChatTyping action
		found := false
		for _, call := range mockBot.getRequestCalls() {
			if action, ok := call.(tgbotapi.ChatActionConfig); ok && action.Action == tgbotapi.ChatTyping {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected at least one typing action via Request")
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for handler")
	}
}
