package telegram

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockBotAPI implements BotAPI interface
type MockBotAPI struct {
	mu                       sync.Mutex
	GetUpdatesChanFunc       func(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
	SendFunc                 func(c tgbotapi.Chattable) (tgbotapi.Message, error)
	GetSelfFunc              func() tgbotapi.User
	StopReceivingUpdatesFunc func()
	SentMessages             []tgbotapi.Chattable
	RequestCalls             []tgbotapi.Chattable
	StopCalls                int
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

func (m *MockBotAPI) StopReceivingUpdates() {
	m.mu.Lock()
	m.StopCalls++
	m.mu.Unlock()
	if m.StopReceivingUpdatesFunc != nil {
		m.StopReceivingUpdatesFunc()
	}
}

func (m *MockBotAPI) GetSelf() tgbotapi.User {
	if m.GetSelfFunc != nil {
		return m.GetSelfFunc()
	}
	return tgbotapi.User{ID: 12345, UserName: "TestBot"}
}

func TestTelegramChannel(t *testing.T) {
	t.Parallel()

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
	require.NoError(t, err)

	msgProcessed := make(chan bool)

	channel.SetHandler(func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error) {
		assert.Equal(t, "Hello Bot", msg.Text)
		assert.Equal(t, int64(999), msg.UserID)
		msgProcessed <- true
		return &OutgoingMessage{Text: "Reply"}, nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, channel.Start(ctx))
	defer func() { require.NoError(t, channel.Stop(context.Background())) }()

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

		// Check thinking placeholder was posted via Send
		sentMsgs := mockBot.getSentMessages()
		require.NotEmpty(t, sentMsgs, "expected Send to be called")

		// First send: thinking placeholder
		placeholder, ok := sentMsgs[0].(tgbotapi.MessageConfig)
		require.True(t, ok, "expected MessageConfig for placeholder, got %T", sentMsgs[0])
		assert.Contains(t, placeholder.Text, "Thinking")

		// Second send: edit with response
		require.True(t, len(sentMsgs) >= 2, "expected at least 2 Send calls (placeholder + edit)")
		_, isEdit := sentMsgs[1].(tgbotapi.EditMessageTextConfig)
		assert.True(t, isEdit, "expected EditMessageTextConfig for response, got %T", sentMsgs[1])
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message processing")
	}
}

func TestTelegramHTTPClient_DefaultHasBoundedTimeout(t *testing.T) {
	t.Parallel()

	client := telegramHTTPClient(Config{})
	require.NotNil(t, client)
	assert.Equal(t, defaultTelegramHTTPClientTimeout, client.Timeout)
	assert.Greater(t, client.Timeout, 60*time.Second)
}

func TestTelegramHTTPClient_PreservesInjectedClient(t *testing.T) {
	t.Parallel()

	injected := &http.Client{Timeout: 5 * time.Second}

	client := telegramHTTPClient(Config{HTTPClient: injected})

	assert.Same(t, injected, client)
}

func TestTelegramChannelIdentityAndApprovalProvider(t *testing.T) {
	t.Parallel()

	mockBot := &MockBotAPI{}
	channel, err := New(Config{
		BotToken:           "TEST_TOKEN",
		Bot:                mockBot,
		ApprovalTimeoutSec: 7,
	})
	require.NoError(t, err)

	provider := channel.GetApprovalProvider()
	require.NotNil(t, provider)
	assert.Equal(t, "telegram", channel.Name())
	assert.Same(t, provider, channel.approval)
	assert.Same(t, mockBot, provider.bot)
	assert.Equal(t, 7*time.Second, provider.timeout)
	assert.Equal(t, "telegram", provider.Name())
	assert.True(t, provider.CanHandle("telegram:123:456"))
}

func TestTelegramSendFormatsSplitsAndFallsBackToPlainText(t *testing.T) {
	t.Parallel()

	sendErr := errors.New("telegram markdown parse failed")
	var sendCount int
	mockBot := &MockBotAPI{
		SendFunc: func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			sendCount++
			if sendCount == 1 {
				return tgbotapi.Message{}, sendErr
			}
			return tgbotapi.Message{MessageID: 200 + sendCount}, nil
		},
	}
	channel := &Channel{bot: mockBot}

	err := channel.Send(42, &OutgoingMessage{
		Text:           "**hello**",
		ReplyToID:      99,
		DisablePreview: true,
	})
	require.NoError(t, err)

	sent := mockBot.getSentMessages()
	require.Len(t, sent, 2)

	formatted, ok := sent[0].(tgbotapi.MessageConfig)
	require.True(t, ok)
	assert.Equal(t, int64(42), formatted.ChatID)
	assert.Equal(t, "*hello*", formatted.Text)
	assert.Equal(t, "Markdown", formatted.ParseMode)
	assert.Equal(t, 99, formatted.ReplyToMessageID)
	assert.True(t, formatted.DisableWebPagePreview)

	plain, ok := sent[1].(tgbotapi.MessageConfig)
	require.True(t, ok)
	assert.Equal(t, "**hello**", plain.Text)
	assert.Empty(t, plain.ParseMode)
	assert.Equal(t, 99, plain.ReplyToMessageID)
	assert.True(t, plain.DisableWebPagePreview)
}

func TestTelegramSendSplitsLongMessages(t *testing.T) {
	t.Parallel()

	mockBot := &MockBotAPI{}
	channel := &Channel{bot: mockBot}
	longText := strings.Repeat("a", 4097)

	err := channel.Send(42, &OutgoingMessage{
		Text:      longText,
		ParseMode: "HTML",
		ReplyToID: 10,
	})
	require.NoError(t, err)

	sent := mockBot.getSentMessages()
	require.Len(t, sent, 2)
	first, ok := sent[0].(tgbotapi.MessageConfig)
	require.True(t, ok)
	assert.Equal(t, "HTML", first.ParseMode)
	assert.Equal(t, 10, first.ReplyToMessageID)
	assert.Len(t, first.Text, 4096)

	second, ok := sent[1].(tgbotapi.MessageConfig)
	require.True(t, ok)
	assert.Equal(t, "HTML", second.ParseMode)
	assert.Zero(t, second.ReplyToMessageID)
	assert.Equal(t, "a", second.Text)
}

func TestTelegramSendReturnsPlainTextFallbackError(t *testing.T) {
	t.Parallel()

	mockBot := &MockBotAPI{
		SendFunc: func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			return tgbotapi.Message{}, errors.New("telegram unavailable")
		},
	}
	channel := &Channel{bot: mockBot}

	err := channel.Send(42, &OutgoingMessage{Text: "**hello**"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "send plain text fallback")
	assert.Contains(t, err.Error(), "send chunk 0")
}

func TestTelegramSendErrorFormatsUserMessage(t *testing.T) {
	t.Parallel()

	mockBot := &MockBotAPI{}
	channel := &Channel{bot: mockBot}

	channel.sendError(42, 99, channelUserError{message: "retry later"})

	sent := mockBot.getSentMessages()
	require.Len(t, sent, 1)
	msg, ok := sent[0].(tgbotapi.MessageConfig)
	require.True(t, ok)
	assert.Equal(t, int64(42), msg.ChatID)
	assert.Equal(t, 99, msg.ReplyToMessageID)
	assert.Equal(t, "❌ retry later", msg.Text)
	assert.Equal(t, "Markdown", msg.ParseMode)
}

func TestFormatChannelErrorFallsBackToErrorText(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "retry later", formatChannelError(channelUserError{message: "retry later"}))
	assert.Equal(t, "Error: raw failure", formatChannelError(errors.New("raw failure")))
}

func TestTelegramSplitMessage(t *testing.T) {
	t.Parallel()

	channel := &Channel{}

	tests := []struct {
		name   string
		text   string
		maxLen int
		want   []string
	}{
		{
			name:   "short message is unchanged",
			text:   "short",
			maxLen: 10,
			want:   []string{"short"},
		},
		{
			name:   "splits on line boundary",
			text:   "alpha\nbeta\ngamma",
			maxLen: 10,
			want:   []string{"alpha\nbeta", "gamma"},
		},
		{
			name:   "splits long line into fixed chunks",
			text:   "abcdefghijkl",
			maxLen: 5,
			want:   []string{"abcde", "fghij", "kl"},
		},
		{
			name:   "keeps trailing short line after long line",
			text:   "abcdefghij\nxy",
			maxLen: 4,
			want:   []string{"abcd", "efgh", "ij", "xy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, channel.splitMessage(tt.text, tt.maxLen))
			for _, chunk := range channel.splitMessage(tt.text, tt.maxLen) {
				assert.LessOrEqual(t, len(chunk), tt.maxLen)
			}
		})
	}
}

func TestTelegramStartTypingStopIsIdempotent(t *testing.T) {
	t.Parallel()

	mockBot := &MockBotAPI{}
	channel := &Channel{bot: mockBot}

	stop := channel.StartTyping(context.Background(), 42)
	stop()
	require.NotPanics(t, stop)

	require.Len(t, mockBot.RequestCalls, 1)
	action, ok := mockBot.RequestCalls[0].(tgbotapi.ChatActionConfig)
	require.True(t, ok)
	assert.Equal(t, int64(42), action.ChatID)
	assert.Equal(t, tgbotapi.ChatTyping, action.Action)
}

func TestTelegramStartTypingReturnsStopWhenInitialRequestFails(t *testing.T) {
	t.Parallel()

	requestErr := errors.New("request failed")
	mockBot := &MockApprovalBotAPI{
		RequestFunc: func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
			return nil, requestErr
		},
	}
	channel := &Channel{bot: mockBot}

	stop := channel.StartTyping(context.Background(), 42)

	require.NotNil(t, stop)
	require.NotPanics(t, func() {
		stop()
		stop()
	})
}

func TestTelegramAllowlistBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		allowlist []int64
		chatID    int64
		userID    int64
		want      bool
	}{
		{name: "empty allowlist permits all", chatID: 1, userID: 2, want: true},
		{name: "matching chat id permits", allowlist: []int64{10}, chatID: 10, userID: 20, want: true},
		{name: "matching user id permits", allowlist: []int64{20}, chatID: 10, userID: 20, want: true},
		{name: "no match blocks", allowlist: []int64{30}, chatID: 10, userID: 20, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			channel := &Channel{config: Config{Allowlist: tt.allowlist}}
			assert.Equal(t, tt.want, channel.isAllowed(tt.chatID, tt.userID))
		})
	}
}

func TestTelegramStartSkipsBlockedAllowlistMessage(t *testing.T) {
	t.Parallel()

	updatesCh := make(chan tgbotapi.Update)
	mockBot := &MockBotAPI{
		GetUpdatesChanFunc: func(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
			return updatesCh
		},
		StopReceivingUpdatesFunc: func() {
			close(updatesCh)
		},
	}
	channel, err := New(Config{
		BotToken:  "TEST_TOKEN",
		Bot:       mockBot,
		Allowlist: []int64{12345},
	})
	require.NoError(t, err)

	handled := make(chan struct{}, 1)
	channel.SetHandler(func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error) {
		handled <- struct{}{}
		return &OutgoingMessage{Text: "should not send"}, nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, channel.Start(ctx))

	sent := make(chan struct{})
	go func() {
		updatesCh <- tgbotapi.Update{
			UpdateID: 4,
			Message: &tgbotapi.Message{
				MessageID: 400,
				From:      &tgbotapi.User{ID: 999, UserName: "blocked"},
				Chat:      &tgbotapi.Chat{ID: 888, Type: "private"},
				Text:      "blocked",
			},
		}
		close(sent)
	}()

	select {
	case <-sent:
	case <-time.After(time.Second):
		t.Fatal("blocked update was not consumed")
	}

	require.NoError(t, channel.Stop(context.Background()))
	select {
	case <-handled:
		t.Fatal("blocked message reached handler")
	default:
	}
	assert.Empty(t, mockBot.getSentMessages())
}

type channelUserError struct {
	message string
}

func (e channelUserError) Error() string {
	return fmt.Sprintf("internal: %s", e.message)
}

func (e channelUserError) UserMessage() string {
	return e.message
}

func TestTelegramTypingIndicator(t *testing.T) {
	t.Parallel()

	updatesCh := make(chan tgbotapi.Update, 1)

	mockBot := &MockBotAPI{
		GetUpdatesChanFunc: func(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
			return updatesCh
		},
	}

	cfg := Config{BotToken: "TEST_TOKEN", Bot: mockBot}
	channel, err := New(cfg)
	require.NoError(t, err)

	done := make(chan struct{})
	channel.SetHandler(func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error) {
		close(done)
		return &OutgoingMessage{Text: "ok"}, nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, channel.Start(ctx))
	defer func() { require.NoError(t, channel.Stop(context.Background())) }()

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

		// Verify thinking placeholder was posted
		sentMsgs := mockBot.getSentMessages()
		found := false
		for _, msg := range sentMsgs {
			if msgCfg, ok := msg.(tgbotapi.MessageConfig); ok {
				if msgCfg.Text == "_Thinking..._" {
					found = true
					break
				}
			}
		}
		assert.True(t, found, "expected thinking placeholder message")
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for handler")
	}
}

func TestTelegramStopStopsReceivingUpdatesBeforeWait(t *testing.T) {
	t.Parallel()

	updatesCh := make(chan tgbotapi.Update)
	stoppedUpdates := make(chan struct{})

	mockBot := &MockBotAPI{
		GetUpdatesChanFunc: func(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
			return updatesCh
		},
		StopReceivingUpdatesFunc: func() {
			close(stoppedUpdates)
			close(updatesCh)
		},
	}

	cfg := Config{BotToken: "TEST_TOKEN", Bot: mockBot}
	channel, err := New(cfg)
	require.NoError(t, err)
	channel.SetHandler(func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error) {
		return &OutgoingMessage{Text: "ok"}, nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, channel.Start(ctx))

	require.NoError(t, channel.Stop(context.Background()))

	select {
	case <-stoppedUpdates:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected StopReceivingUpdates to be called")
	}
}

func TestTelegramStopReturnsContextErrorWhenWorkersDoNotExit(t *testing.T) {
	t.Parallel()

	updatesCh := make(chan tgbotapi.Update, 1)
	release := make(chan struct{})

	mockBot := &MockBotAPI{
		GetUpdatesChanFunc: func(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
			return updatesCh
		},
		StopReceivingUpdatesFunc: func() {
			close(updatesCh)
		},
	}

	cfg := Config{BotToken: "TEST_TOKEN", Bot: mockBot}
	channel, err := New(cfg)
	require.NoError(t, err)

	channel.SetHandler(func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error) {
		<-release
		return &OutgoingMessage{Text: "done"}, nil
	})

	parentCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, channel.Start(parentCtx))

	updatesCh <- tgbotapi.Update{
		UpdateID: 3,
		Message: &tgbotapi.Message{
			MessageID: 300,
			From:      &tgbotapi.User{ID: 777, UserName: "tester"},
			Chat:      &tgbotapi.Chat{ID: 777, Type: "private"},
			Text:      "block",
		},
	}

	time.Sleep(50 * time.Millisecond)

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer stopCancel()

	err = channel.Stop(stopCtx)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	close(release)
	require.NoError(t, channel.Stop(context.Background()))
}
