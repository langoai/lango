package discord

import (
	"context"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSession implements Session interface for testing
type MockSession struct {
	Handlers      []interface{}
	SentMessages  []string
	EditedMessages []string
	State         *discordgo.State
	TypingCalls   []string
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
	return &discordgo.Message{ID: "mock-msg-id", Content: content}, nil
}

func (m *MockSession) ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	m.SentMessages = append(m.SentMessages, data.Content)
	return &discordgo.Message{Content: data.Content}, nil
}

func (m *MockSession) ChannelMessageEditComplex(edit *discordgo.MessageEdit, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	if edit.Content != nil {
		m.EditedMessages = append(m.EditedMessages, *edit.Content)
	}
	return &discordgo.Message{}, nil
}

func (m *MockSession) ChannelTyping(channelID string, options ...discordgo.RequestOption) error {
	m.TypingCalls = append(m.TypingCalls, channelID)
	return nil
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
	t.Parallel()

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
	require.NoError(t, err)

	// Set a handler that replies
	channel.SetHandler(func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error) {
		assert.Equal(t, "Hello", msg.Content)
		return &OutgoingMessage{Content: "World"}, nil
	})

	// Start (registers handler)
	require.NoError(t, channel.Start(context.Background()))

	// Retrieve registered message handler (first one registered)
	var handlerFunc func(*discordgo.Session, *discordgo.MessageCreate)
	for _, h := range mockSession.Handlers {
		if fn, ok := h.(func(*discordgo.Session, *discordgo.MessageCreate)); ok {
			handlerFunc = fn
			break
		}
	}
	require.NotNil(t, handlerFunc, "message handler not registered or wrong type")

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

	// Verify thinking placeholder was sent
	require.NotEmpty(t, mockSession.SentMessages, "expected thinking placeholder to be sent")
	assert.Equal(t, "_Thinking..._", mockSession.SentMessages[0])

	// Verify response was sent via edit (placeholder replaced with response)
	require.NotEmpty(t, mockSession.EditedMessages, "expected response via edit")
	assert.Equal(t, "World", mockSession.EditedMessages[0])
}

func TestDiscordTypingIndicator(t *testing.T) {
	t.Parallel()

	state := &discordgo.State{}
	state.User = &discordgo.User{ID: "bot-123", Username: "TestBot"}
	mockSession := &MockSession{State: state}

	cfg := Config{BotToken: "TEST_TOKEN", Session: mockSession}
	channel, err := New(cfg)
	require.NoError(t, err)

	handlerCalled := make(chan struct{})
	channel.SetHandler(func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error) {
		close(handlerCalled)
		return &OutgoingMessage{Content: "done"}, nil
	})

	require.NoError(t, channel.Start(context.Background()))

	// Find the message handler
	var handlerFunc func(*discordgo.Session, *discordgo.MessageCreate)
	for _, h := range mockSession.Handlers {
		if fn, ok := h.(func(*discordgo.Session, *discordgo.MessageCreate)); ok {
			handlerFunc = fn
			break
		}
	}
	require.NotNil(t, handlerFunc, "message handler not registered")

	handlerFunc(nil, &discordgo.MessageCreate{
		Message: &discordgo.Message{
			ID:        "msg-2",
			ChannelID: "chan-typing",
			Content:   "test",
			Author:    &discordgo.User{ID: "user-2", Username: "User"},
		},
	})

	// Thinking placeholder should have been posted
	require.NotEmpty(t, mockSession.SentMessages, "expected thinking placeholder to be sent")
	found := false
	for _, msg := range mockSession.SentMessages {
		if msg == "_Thinking..._" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected thinking placeholder message")
}
