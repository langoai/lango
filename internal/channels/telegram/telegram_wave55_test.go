package telegram

import (
	"context"
	"errors"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWave55NewRejectsMissingBotToken(t *testing.T) {
	t.Parallel()

	channel, err := New(Config{})

	require.Error(t, err)
	assert.Nil(t, channel)
	assert.Contains(t, err.Error(), "bot token is required")
}

func TestWave55StartRequiresMessageHandler(t *testing.T) {
	t.Parallel()

	channel, err := New(Config{BotToken: "TEST_TOKEN", Bot: &MockBotAPI{}})
	require.NoError(t, err)

	err = channel.Start(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "message handler not set")
}

func TestWave55EditMessageRetriesPlainTextWhenMarkdownFails(t *testing.T) {
	t.Parallel()

	var sendCount int
	mockBot := &MockBotAPI{
		SendFunc: func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			sendCount++
			if sendCount == 1 {
				return tgbotapi.Message{}, errors.New("markdown parse failed")
			}
			return tgbotapi.Message{MessageID: 20}, nil
		},
	}
	channel := &Channel{bot: mockBot}

	channel.editMessage(42, 7, "**bad markdown")

	sent := mockBot.getSentMessages()
	require.Len(t, sent, 2)
	formatted, ok := sent[0].(tgbotapi.EditMessageTextConfig)
	require.True(t, ok)
	assert.Equal(t, int64(42), formatted.ChatID)
	assert.Equal(t, 7, formatted.MessageID)
	assert.Equal(t, "Markdown", formatted.ParseMode)

	plain, ok := sent[1].(tgbotapi.EditMessageTextConfig)
	require.True(t, ok)
	assert.Equal(t, int64(42), plain.ChatID)
	assert.Equal(t, 7, plain.MessageID)
	assert.Equal(t, "**bad markdown", plain.Text)
	assert.Empty(t, plain.ParseMode)
}

func TestWave55HandleUpdateMapsReplyAndMediaBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		message    *tgbotapi.Message
		wantType   string
		wantFileID string
	}{
		{
			name: "photo",
			message: &tgbotapi.Message{
				Photo: []tgbotapi.PhotoSize{
					{FileID: "small-photo"},
					{FileID: "large-photo"},
				},
			},
			wantType:   "photo",
			wantFileID: "large-photo",
		},
		{
			name: "document",
			message: &tgbotapi.Message{
				Document: &tgbotapi.Document{FileID: "doc-file"},
			},
			wantType:   "document",
			wantFileID: "doc-file",
		},
		{
			name: "voice",
			message: &tgbotapi.Message{
				Voice: &tgbotapi.Voice{FileID: "voice-file"},
			},
			wantType:   "voice",
			wantFileID: "voice-file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockBot := &MockBotAPI{}
			channel := &Channel{bot: mockBot}
			seen := make(chan *IncomingMessage, 1)
			channel.SetHandler(func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error) {
				seen <- msg
				return &OutgoingMessage{Text: "ok"}, nil
			})

			tt.message.MessageID = 100
			tt.message.From = &tgbotapi.User{ID: 55, UserName: "tester"}
			tt.message.Chat = &tgbotapi.Chat{ID: 66, Type: "private"}
			tt.message.Text = "payload"
			tt.message.ReplyToMessage = &tgbotapi.Message{MessageID: 99}

			channel.handleUpdate(context.Background(), tgbotapi.Update{UpdateID: 1, Message: tt.message})

			var incoming *IncomingMessage
			select {
			case incoming = <-seen:
			default:
				t.Fatal("handler was not called")
			}
			assert.Equal(t, 100, incoming.MessageID)
			assert.Equal(t, int64(66), incoming.ChatID)
			assert.Equal(t, int64(55), incoming.UserID)
			assert.Equal(t, "tester", incoming.Username)
			assert.Equal(t, "payload", incoming.Text)
			assert.Equal(t, 99, incoming.ReplyToID)
			assert.True(t, incoming.HasMedia)
			assert.Equal(t, tt.wantType, incoming.MediaType)
			assert.Equal(t, tt.wantFileID, incoming.MediaFileID)
		})
	}
}

func TestWave55HandleUpdateFallsBackToTypingAndSendErrorWhenPlaceholderFails(t *testing.T) {
	t.Parallel()

	var sendCount int
	mockBot := &MockBotAPI{
		SendFunc: func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			sendCount++
			if sendCount == 1 {
				return tgbotapi.Message{}, errors.New("placeholder failed")
			}
			return tgbotapi.Message{MessageID: 200 + sendCount}, nil
		},
	}
	channel := &Channel{bot: mockBot}
	channel.SetHandler(func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error) {
		return nil, channelUserError{message: "try again later"}
	})

	channel.handleUpdate(context.Background(), tgbotapi.Update{
		UpdateID: 9,
		Message: &tgbotapi.Message{
			MessageID: 44,
			From:      &tgbotapi.User{ID: 55, UserName: "tester"},
			Chat:      &tgbotapi.Chat{ID: 66, Type: "private"},
			Text:      "fail",
		},
	})

	require.Len(t, mockBot.RequestCalls, 1)
	action, ok := mockBot.RequestCalls[0].(tgbotapi.ChatActionConfig)
	require.True(t, ok)
	assert.Equal(t, int64(66), action.ChatID)
	assert.Equal(t, tgbotapi.ChatTyping, action.Action)

	sent := mockBot.getSentMessages()
	require.Len(t, sent, 2)
	errMsg, ok := sent[1].(tgbotapi.MessageConfig)
	require.True(t, ok)
	assert.Equal(t, int64(66), errMsg.ChatID)
	assert.Equal(t, 44, errMsg.ReplyToMessageID)
	assert.Equal(t, "❌ try again later", errMsg.Text)
	assert.Equal(t, "Markdown", errMsg.ParseMode)
}
