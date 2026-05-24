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

type slackNewValidationAndMetadataBranchesSlackClient struct {
	mu sync.Mutex

	authErr    error
	postErrs   []error
	updateErrs []error

	posts   []slackNewValidationAndMetadataBranchesPostCall
	updates []slackNewValidationAndMetadataBranchesUpdateCall
	deletes []slackNewValidationAndMetadataBranchesDeleteCall
}

type slackNewValidationAndMetadataBranchesPostCall struct {
	channelID string
	options   []slack.MsgOption
}

type slackNewValidationAndMetadataBranchesUpdateCall struct {
	channelID string
	timestamp string
	options   []slack.MsgOption
}

type slackNewValidationAndMetadataBranchesDeleteCall struct {
	channelID string
	timestamp string
}

func (c *slackNewValidationAndMetadataBranchesSlackClient) AuthTest() (*slack.AuthTestResponse, error) {
	if c.authErr != nil {
		return nil, c.authErr
	}
	return &slack.AuthTestResponse{UserID: "bot-runAndCollectUsesLearnedFixRetryResponse4", Team: "runAndCollectUsesLearnedFixRetryResponse4"}, nil
}

func (c *slackNewValidationAndMetadataBranchesSlackClient) PostMessage(
	channelID string,
	options ...slack.MsgOption,
) (string, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.posts = append(c.posts, slackNewValidationAndMetadataBranchesPostCall{channelID: channelID, options: options})
	if len(c.postErrs) > 0 {
		err := c.postErrs[0]
		c.postErrs = c.postErrs[1:]
		if err != nil {
			return "", "", err
		}
	}
	return channelID, "ts-post", nil
}

func (c *slackNewValidationAndMetadataBranchesSlackClient) UpdateMessage(
	channelID string,
	timestamp string,
	options ...slack.MsgOption,
) (string, string, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.updates = append(c.updates, slackNewValidationAndMetadataBranchesUpdateCall{
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

func (c *slackNewValidationAndMetadataBranchesSlackClient) DeleteMessage(
	channelID string,
	messageTimestamp string,
) (string, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.deletes = append(c.deletes, slackNewValidationAndMetadataBranchesDeleteCall{
		channelID: channelID,
		timestamp: messageTimestamp,
	})
	return channelID, messageTimestamp, nil
}

func (c *slackNewValidationAndMetadataBranchesSlackClient) snapshot() (
	[]slackNewValidationAndMetadataBranchesPostCall,
	[]slackNewValidationAndMetadataBranchesUpdateCall,
	[]slackNewValidationAndMetadataBranchesDeleteCall,
) {
	c.mu.Lock()
	defer c.mu.Unlock()

	posts := append([]slackNewValidationAndMetadataBranchesPostCall(nil), c.posts...)
	updates := append([]slackNewValidationAndMetadataBranchesUpdateCall(nil), c.updates...)
	deletes := append([]slackNewValidationAndMetadataBranchesDeleteCall(nil), c.deletes...)
	return posts, updates, deletes
}

type slackNewValidationAndMetadataBranchesSocket struct {
	events chan socketmode.Event
	acks   int
}

func (s *slackNewValidationAndMetadataBranchesSocket) Run() error { return nil }

func (s *slackNewValidationAndMetadataBranchesSocket) Ack(socketmode.Request, ...interface{}) {
	s.acks++
}

func (s *slackNewValidationAndMetadataBranchesSocket) Events() <-chan socketmode.Event {
	return s.events
}

func TestSlackNewValidationAndMetadataBranches(t *testing.T) {
	_, err := New(Config{AppToken: "app"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bot token is required")

	_, err = New(Config{BotToken: "bot"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "app token is required")

	_, err = New(Config{BotToken: "bot", AppToken: "app", Client: &slackNewValidationAndMetadataBranchesSlackClient{}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must provide Socket if Client is mocked")

	channel, err := New(Config{
		BotToken: "bot",
		AppToken: "app",
		Client:   &slackNewValidationAndMetadataBranchesSlackClient{},
		Socket:   &slackNewValidationAndMetadataBranchesSocket{events: make(chan socketmode.Event)},
	})
	require.NoError(t, err)
	assert.Equal(t, "slack", channel.Name())
	assert.NotNil(t, channel.GetApprovalProvider())
}

func TestSlackStartValidationAuthAndStopBranches(t *testing.T) {
	channel, err := New(Config{
		BotToken: "bot",
		AppToken: "app",
		Client:   &slackNewValidationAndMetadataBranchesSlackClient{},
		Socket:   &slackNewValidationAndMetadataBranchesSocket{events: make(chan socketmode.Event)},
	})
	require.NoError(t, err)

	err = channel.Start(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "message handler not set")

	channel.SetHandler(func(context.Context, *IncomingMessage) (*OutgoingMessage, error) {
		return nil, nil
	})
	channel.api = &slackNewValidationAndMetadataBranchesSlackClient{authErr: errors.New("auth exploded")}
	err = channel.Start(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth test failed")

	stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, channel.Stop(stopCtx))
}

func TestSlackHandleInteractiveEventAcksAndRoutesBlockActions(t *testing.T) {
	socket := &slackNewValidationAndMetadataBranchesSocket{events: make(chan socketmode.Event)}
	channel, err := New(Config{
		BotToken: "bot",
		AppToken: "app",
		Client:   &slackNewValidationAndMetadataBranchesSlackClient{},
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

func TestSlackStartTypingDeletesPlaceholderOnceAndNoopsOnPostFailure(t *testing.T) {
	client := &slackNewValidationAndMetadataBranchesSlackClient{}
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

	failingClient := &slackNewValidationAndMetadataBranchesSlackClient{postErrs: []error{errors.New("post failed")}}
	failingChannel := &Channel{api: failingClient}
	failingChannel.StartTyping("chan-fail")()

	_, _, failingDeletes := failingClient.snapshot()
	assert.Empty(t, failingDeletes)
}

func TestSlackSendCoversThreadBlocksAndPostErrors(t *testing.T) {
	client := &slackNewValidationAndMetadataBranchesSlackClient{}
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

	failingClient := &slackNewValidationAndMetadataBranchesSlackClient{postErrs: []error{errors.New("post failed")}}
	failingChannel := &Channel{api: failingClient}
	err = failingChannel.Send("chan-send", &OutgoingMessage{Text: "text"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "post failed")
}

func TestSlackHandleMessageBranchesWithoutSocketLoop(t *testing.T) {
	t.Run("ignores bot user", func(t *testing.T) {
		client := &slackNewValidationAndMetadataBranchesSlackClient{}
		channel := &Channel{api: client, botID: "bot-runAndCollectUsesLearnedFixRetryResponse4"}
		channel.handleMessage(context.Background(), "message", "chan", "bot-runAndCollectUsesLearnedFixRetryResponse4", "ignored", "")

		posts, updates, _ := client.snapshot()
		assert.Empty(t, posts)
		assert.Empty(t, updates)
	})

	t.Run("placeholder update failure falls back to send", func(t *testing.T) {
		client := &slackNewValidationAndMetadataBranchesSlackClient{updateErrs: []error{errors.New("update failed")}}
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
		client := &slackNewValidationAndMetadataBranchesSlackClient{postErrs: []error{errors.New("placeholder failed"), nil}}
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
		client := &slackNewValidationAndMetadataBranchesSlackClient{}
		channel := &Channel{api: client}
		channel.SetHandler(func(context.Context, *IncomingMessage) (*OutgoingMessage, error) {
			return nil, slackNewValidationAndMetadataBranchesUserError{message: "safe error"}
		})

		channel.handleMessage(context.Background(), "message", "chan", "user", "hello", "thread-1")
		channel.wg.Wait()

		posts, updates, _ := client.snapshot()
		require.Len(t, posts, 2)
		require.Len(t, updates, 1)
	})
}

type slackNewValidationAndMetadataBranchesUserError struct {
	message string
}

func (e slackNewValidationAndMetadataBranchesUserError) Error() string {
	return "internal detail"
}

func (e slackNewValidationAndMetadataBranchesUserError) UserMessage() string {
	return e.message
}
