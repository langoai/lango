package discord

import (
	"context"
	"testing"

	"github.com/bwmarrin/discordgo"
)

// MockSession implements Session interface for testing
type MockSession struct {
	Handlers     []interface{}
	SentMessages []string
	State        *discordgo.State
}

func (m *MockSession) Open() error {
	return nil
}

func (m *MockSession) Close() error {
	return nil
}

func (m *MockSession) AddHandler(handler interface{}) func() {
	m.Handlers = append(m.Handlers, handler)
	return func() {}
}

func (m *MockSession) ChannelMessageSend(channelID string, content string, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	m.SentMessages = append(m.SentMessages, content)
	return &discordgo.Message{Content: content}, nil
}

func (m *MockSession) ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	m.SentMessages = append(m.SentMessages, data.Content)
	return &discordgo.Message{Content: data.Content}, nil
}

func (m *MockSession) ChannelMessageEditComplex(edit *discordgo.MessageEdit, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	return &discordgo.Message{}, nil
}

func (m *MockSession) InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
	return nil
}

func (m *MockSession) ApplicationCommandCreate(appID string, guildID string, cmd *discordgo.ApplicationCommand, options ...discordgo.RequestOption) (*discordgo.ApplicationCommand, error) {
	return cmd, nil
}

func (m *MockSession) GetState() *discordgo.State {
	return m.State
}

func TestDiscordChannel(t *testing.T) {
	// Setup Mock
	state := &discordgo.State{}
	state.User = &discordgo.User{
		ID:       "bot-123",
		Username: "TestBot",
	}
	mockSession := &MockSession{
		State: state,
	}

	cfg := Config{
		BotToken: "TEST_TOKEN",
		Session:  mockSession,
	}

	channel, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}

	// Set a handler that replies
	channel.SetHandler(func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error) {
		if msg.Content != "Hello" {
			t.Errorf("expected 'Hello', got '%s'", msg.Content)
		}
		return &OutgoingMessage{Content: "World"}, nil
	})

	// Start (registers handler)
	if err := channel.Start(context.Background()); err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	// Retrieve registered message handler (first one registered)
	var handlerFunc func(*discordgo.Session, *discordgo.MessageCreate)
	for _, h := range mockSession.Handlers {
		if fn, ok := h.(func(*discordgo.Session, *discordgo.MessageCreate)); ok {
			handlerFunc = fn
			break
		}
	}
	if handlerFunc == nil {
		t.Fatalf("message handler not registered or wrong type")
	}

	// Simulate incoming message
	handlerFunc(nil, &discordgo.MessageCreate{
		Message: &discordgo.Message{
			ID:        "msg-1",
			ChannelID: "chan-1",
			Content:   "Hello",
			Author: &discordgo.User{
				ID:       "user-1",
				Username: "User",
			},
		},
	})

	// Verify response was sent
	if len(mockSession.SentMessages) != 1 {
		t.Errorf("expected 1 sent message, got %d", len(mockSession.SentMessages))
	} else if mockSession.SentMessages[0] != "World" {
		t.Errorf("expected 'World', got '%s'", mockSession.SentMessages[0])
	}
}
