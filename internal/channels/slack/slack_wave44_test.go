package slack

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type wave44SlackClient struct {
	mu sync.Mutex

	authErr    error
	postErrs   []error
	updateErrs []error

	posts   []wave44PostCall
	updates []wave44UpdateCall
	deletes []wave44DeleteCall
}

type wave44PostCall struct {
	channelID string
	options   []slack.MsgOption
}

type wave44UpdateCall struct {
	channelID string
	timestamp string
	options   []slack.MsgOption
}

type wave44DeleteCall struct {
	channelID string
	timestamp string
}

func (c *wave44SlackClient) AuthTest() (*slack.AuthTestResponse, error) {
	if c.authErr != nil {
		return nil, c.authErr
	}
	return &slack.AuthTestResponse{UserID: "bot-wave44", Team: "wave44"}, nil
}

func (c *wave44SlackClient) PostMessage(
	channelID string,
	options ...slack.MsgOption,
) (string, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.posts = append(c.posts, wave44PostCall{channelID: channelID, options: options})
	if len(c.postErrs) > 0 {
		err := c.postErrs[0]
		c.postErrs = c.postErrs[1:]
		if err != nil {
			return "", "", err
		}
	}
	return channelID, "ts-post", nil
}

func (c *wave44SlackClient) UpdateMessage(
	channelID string,
	timestamp string,
	options ...slack.MsgOption,
) (string, string, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.updates = append(c.updates, wave44UpdateCall{
		channelID: channelID,
		timestamp: timestamp,
		options:   options,
	})
	if len(c.updateErrs) > 0 {
		err := c.updateErrs[0]
		c.updateErrs = c.updateErrs[1:]
		if err != nil {
			return "", "", "", err
		}
	}
	return channelID, timestamp, "", nil
}

func (c *wave44SlackClient) DeleteMessage(
	channelID string,
	messageTimestamp string,
) (string, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.deletes = append(c.deletes, wave44DeleteCall{
		channelID: channelID,
		timestamp: messageTimestamp,
	})
	return channelID, messageTimestamp, nil
}

func (c *wave44SlackClient) snapshot() (
	[]wave44PostCall,
	[]wave44UpdateCall,
	[]wave44DeleteCall,
) {
	c.mu.Lock()
	defer c.mu.Unlock()

	posts := append([]wave44PostCall(nil), c.posts...)
	updates := append([]wave44UpdateCall(nil), c.updates...)
	deletes := append([]wave44DeleteCall(nil), c.deletes...)
	return posts, updates, deletes
}

type wave44Socket struct {
	events chan socketmode.Event
	acks   int
}

func (s *wave44Socket) Run() error { return nil }

func (s *wave44Socket) Ack(socketmode.Request, ...interface{}) {
	s.acks++
}

func (s *wave44Socket) Events() <-chan socketmode.Event {
	return s.events
}

func TestSlackWave44_NewValidationAndMetadataBranches(t *testing.T) {
	_, err := New(Config{AppToken: "app"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bot token is required")

	_, err = New(Config{BotToken: "bot"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "app token is required")

	_, err = New(Config{BotToken: "bot", AppToken: "app", Client: &wave44SlackClient{}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must provide Socket if Client is mocked")

	channel, err := New(Config{
		BotToken: "bot",
		AppToken: "app",
		Client:   &wave44SlackClient{},
		Socket:   &wave44Socket{events: make(chan socketmode.Event)},
	})
	require.NoError(t, err)
	assert.Equal(t, "slack", channel.Name())
	assert.NotNil(t, channel.GetApprovalProvider())
}

func TestSlackWave44_StartValidationAuthAndStopBranches(t *testing.T) {
	channel, err := New(Config{
		BotToken: "bot",
		AppToken: "app",
		Client:   &wave44SlackClient{},
		Socket:   &wave44Socket{events: make(chan socketmode.Event)},
	})
	require.NoError(t, err)

	err = channel.Start(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "message handler not set")

	channel.SetHandler(func(context.Context, *IncomingMessage) (*OutgoingMessage, error) {
		return nil, nil
	})
	channel.api = &wave44SlackClient{authErr: errors.New("auth exploded")}
	err = channel.Start(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth test failed")

	stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, channel.Stop(stopCtx))
}

func TestSlackWave44_HandleInteractiveEventAcksAndRoutesBlockActions(t *testing.T) {
	socket := &wave44Socket{events: make(chan socketmode.Event)}
	channel, err := New(Config{
		BotToken: "bot",
		AppToken: "app",
		Client:   &wave44SlackClient{},
		Socket:   socket,
	})
	require.NoError(t, err)

	channel.handleInteractiveEvent(socketmode.Event{Data: "not a callback"})
	assert.Equal(t, 0, socket.acks)

	channel.handleInteractiveEvent(socketmode.Event{
		Request: &socketmode.Request{},
		Data: slack.InteractionCallback{
			Type: slack.InteractionTypeBlockActions,
			ActionCallback: slack.ActionCallbacks{
				BlockActions: []*slack.BlockAction{{ActionID: "missing-action"}},
			},
		},
	})
	assert.Equal(t, 1, socket.acks)
}

func TestSlackWave44_StartTypingDeletesPlaceholderOnceAndNoopsOnPostFailure(t *testing.T) {
	client := &wave44SlackClient{}
	channel := &Channel{api: client}

	stop := channel.StartTyping("chan-typing")
	stop()
	stop()

	posts, _, deletes := client.snapshot()
	require.Len(t, posts, 1)
	assert.Equal(t, "chan-typing", posts[0].channelID)
	require.Len(t, deletes, 1)
	assert.Equal(t, "chan-typing", deletes[0].channelID)
	assert.Equal(t, "ts-post", deletes[0].timestamp)

	failingClient := &wave44SlackClient{postErrs: []error{errors.New("post failed")}}
	failingChannel := &Channel{api: failingClient}
	failingChannel.StartTyping("chan-fail")()

	_, _, failingDeletes := failingClient.snapshot()
	assert.Empty(t, failingDeletes)
}

func TestSlackWave44_SendCoversThreadBlocksAndPostErrors(t *testing.T) {
	client := &wave44SlackClient{}
	channel := &Channel{api: client}

	err := channel.Send("chan-send", &OutgoingMessage{
		Text:     "**bold**",
		ThreadTS: "thread-1",
		Blocks: []Block{
			{
				Type: "section",
				Text: &TextBlock{Type: "mrkdwn", Text: "*block stays as supplied*"},
			},
		},
	})
	require.NoError(t, err)

	posts, _, _ := client.snapshot()
	require.Len(t, posts, 1)
	assert.Equal(t, "chan-send", posts[0].channelID)
	assert.Len(t, posts[0].options, 3)

	failingClient := &wave44SlackClient{postErrs: []error{errors.New("post failed")}}
	failingChannel := &Channel{api: failingClient}
	err = failingChannel.Send("chan-send", &OutgoingMessage{Text: "text"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "post failed")
}

func TestSlackWave44_HandleMessageBranchesWithoutSocketLoop(t *testing.T) {
	t.Run("ignores bot user", func(t *testing.T) {
		client := &wave44SlackClient{}
		channel := &Channel{api: client, botID: "bot-wave44"}
		channel.handleMessage(context.Background(), "message", "chan", "bot-wave44", "ignored", "")

		posts, updates, _ := client.snapshot()
		assert.Empty(t, posts)
		assert.Empty(t, updates)
	})

	t.Run("placeholder update failure falls back to send", func(t *testing.T) {
		client := &wave44SlackClient{updateErrs: []error{errors.New("update failed")}}
		channel := &Channel{api: client}
		channel.SetHandler(func(context.Context, *IncomingMessage) (*OutgoingMessage, error) {
			return &OutgoingMessage{Text: "final response"}, nil
		})

		channel.handleMessage(context.Background(), "message", "chan", "user", "hello", "thread-1")
		channel.wg.Wait()

		posts, updates, _ := client.snapshot()
		require.Len(t, posts, 2)
		require.Len(t, updates, 1)
		assert.Equal(t, "ts-post", updates[0].timestamp)
	})

	t.Run("placeholder post failure sends normal response", func(t *testing.T) {
		client := &wave44SlackClient{postErrs: []error{errors.New("placeholder failed"), nil}}
		channel := &Channel{api: client}
		channel.SetHandler(func(context.Context, *IncomingMessage) (*OutgoingMessage, error) {
			return &OutgoingMessage{Text: "fallback response"}, nil
		})

		channel.handleMessage(context.Background(), "message", "chan", "user", "hello", "")
		channel.wg.Wait()

		posts, updates, _ := client.snapshot()
		require.Len(t, posts, 2)
		assert.Empty(t, updates)
	})

	t.Run("handler error updates placeholder and sends error", func(t *testing.T) {
		client := &wave44SlackClient{}
		channel := &Channel{api: client}
		channel.SetHandler(func(context.Context, *IncomingMessage) (*OutgoingMessage, error) {
			return nil, wave44UserError{message: "safe error"}
		})

		channel.handleMessage(context.Background(), "message", "chan", "user", "hello", "thread-1")
		channel.wg.Wait()

		posts, updates, _ := client.snapshot()
		require.Len(t, posts, 2)
		require.Len(t, updates, 1)
	})
}

type wave44UserError struct {
	message string
}

func (e wave44UserError) Error() string {
	return "internal detail"
}

func (e wave44UserError) UserMessage() string {
	return e.message
}
