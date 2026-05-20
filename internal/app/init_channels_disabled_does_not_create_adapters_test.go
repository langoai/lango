package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/channels/discord"
	"github.com/langoai/lango/internal/channels/slack"
	"github.com/langoai/lango/internal/channels/telegram"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/turnrunner"
)

func TestInitChannelsDisabledDoesNotCreateAdapters(t *testing.T) {
	t.Parallel()

	app := &App{Config: &config.Config{}}

	require.NoError(t, app.initChannels())
	assert.Empty(t, app.Channels)
}

func TestRunAgentRequiresTurnRunner(t *testing.T) {
	t.Parallel()

	app := &App{Config: initChannelsDisabledDoesNotCreateAdaptersChannelConfig()}

	got, err := app.runAgent(context.Background(), "telegram:1:2", "hello")

	require.Error(t, err)
	assert.Empty(t, got)
	assert.Contains(t, err.Error(), "turn runner is not initialized")
}

func TestHandleTelegramMessagePublishesReceiveAndSentEvents(t *testing.T) {
	t.Parallel()

	bus, received, sent := initChannelsDisabledDoesNotCreateAdaptersSubscribeChannelEvents()
	executor := &initChannelsDisabledDoesNotCreateAdaptersChannelExecutor{response: "telegram response"}
	app := initChannelsDisabledDoesNotCreateAdaptersChannelApp(bus, executor)

	out, err := app.handleTelegramMessage(context.Background(), &telegram.IncomingMessage{
		ChatID:   1001,
		UserID:   2002,
		Username: "alice",
		Text:     "hello telegram",
	})

	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "telegram response", out.Text)
	require.Len(t, executor.calls, 1)
	assert.Equal(t, initChannelsDisabledDoesNotCreateAdaptersChannelCall{sessionID: "telegram:1001:2002", input: "hello telegram"}, executor.calls[0])
	require.Len(t, *received, 1)
	assert.Equal(t, eventbus.ChannelMessageReceivedEvent{
		Channel:    "telegram",
		SessionKey: "telegram:1001:2002",
		SenderName: "alice",
		SenderID:   "2002",
		Text:       "hello telegram",
		Metadata:   map[string]string{"chatID": "1001"},
	}, initChannelsDisabledDoesNotCreateAdaptersWithoutReceivedTimestamp((*received)[0]))
	require.Len(t, *sent, 1)
	assert.Equal(t, eventbus.ChannelMessageSentEvent{
		Channel:      "telegram",
		SessionKey:   "telegram:1001:2002",
		ResponseText: "telegram response",
	}, initChannelsDisabledDoesNotCreateAdaptersWithoutSentTimestamp((*sent)[0]))
}

func TestHandleDiscordMessageOmitsEmptyGuildMetadata(t *testing.T) {
	t.Parallel()

	bus, received, sent := initChannelsDisabledDoesNotCreateAdaptersSubscribeChannelEvents()
	executor := &initChannelsDisabledDoesNotCreateAdaptersChannelExecutor{response: "discord response"}
	app := initChannelsDisabledDoesNotCreateAdaptersChannelApp(bus, executor)

	out, err := app.handleDiscordMessage(context.Background(), &discord.IncomingMessage{
		ChannelID:  "channel-1",
		AuthorID:   "author-2",
		AuthorName: "bob",
		Content:    "hello discord",
	})

	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "discord response", out.Content)
	require.Len(t, executor.calls, 1)
	assert.Equal(t, initChannelsDisabledDoesNotCreateAdaptersChannelCall{sessionID: "discord:channel-1:author-2", input: "hello discord"}, executor.calls[0])
	require.Len(t, *received, 1)
	assert.Equal(t, eventbus.ChannelMessageReceivedEvent{
		Channel:    "discord",
		SessionKey: "discord:channel-1:author-2",
		SenderName: "bob",
		SenderID:   "author-2",
		Text:       "hello discord",
		Metadata:   map[string]string{"channelID": "channel-1"},
	}, initChannelsDisabledDoesNotCreateAdaptersWithoutReceivedTimestamp((*received)[0]))
	require.Len(t, *sent, 1)
	assert.Equal(t, eventbus.ChannelMessageSentEvent{
		Channel:      "discord",
		SessionKey:   "discord:channel-1:author-2",
		ResponseText: "discord response",
	}, initChannelsDisabledDoesNotCreateAdaptersWithoutSentTimestamp((*sent)[0]))
}

func TestHandleSlackMessageIncludesThreadMetadataWhenPresent(t *testing.T) {
	t.Parallel()

	bus, received, sent := initChannelsDisabledDoesNotCreateAdaptersSubscribeChannelEvents()
	executor := &initChannelsDisabledDoesNotCreateAdaptersChannelExecutor{response: "slack response"}
	app := initChannelsDisabledDoesNotCreateAdaptersChannelApp(bus, executor)

	out, err := app.handleSlackMessage(context.Background(), &slack.IncomingMessage{
		ChannelID: "channel-3",
		UserID:    "user-4",
		Text:      "hello slack",
		ThreadTS:  "1710000000.000100",
	})

	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "slack response", out.Text)
	require.Len(t, executor.calls, 1)
	assert.Equal(t, initChannelsDisabledDoesNotCreateAdaptersChannelCall{sessionID: "slack:channel-3:user-4", input: "hello slack"}, executor.calls[0])
	require.Len(t, *received, 1)
	assert.Equal(t, eventbus.ChannelMessageReceivedEvent{
		Channel:    "slack",
		SessionKey: "slack:channel-3:user-4",
		SenderName: "user-4",
		SenderID:   "user-4",
		Text:       "hello slack",
		Metadata:   map[string]string{"channelID": "channel-3", "threadTS": "1710000000.000100"},
	}, initChannelsDisabledDoesNotCreateAdaptersWithoutReceivedTimestamp((*received)[0]))
	require.Len(t, *sent, 1)
	assert.Equal(t, eventbus.ChannelMessageSentEvent{
		Channel:      "slack",
		SessionKey:   "slack:channel-3:user-4",
		ResponseText: "slack response",
	}, initChannelsDisabledDoesNotCreateAdaptersWithoutSentTimestamp((*sent)[0]))
}

func TestHandleTelegramMessagePublishesClassifiedRunnerFailureResponse(t *testing.T) {
	t.Parallel()

	bus, received, sent := initChannelsDisabledDoesNotCreateAdaptersSubscribeChannelEvents()
	app := initChannelsDisabledDoesNotCreateAdaptersChannelApp(bus, &initChannelsDisabledDoesNotCreateAdaptersChannelExecutor{err: errors.New("runner failed")})

	out, err := app.handleTelegramMessage(context.Background(), &telegram.IncomingMessage{
		ChatID:   3003,
		UserID:   4004,
		Username: "carol",
		Text:     "runner failure",
	})

	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "An error occurred: runner failed", out.Text)
	require.Len(t, *received, 1)
	assert.Equal(t, "telegram:3003:4004", (*received)[0].SessionKey)
	require.Len(t, *sent, 1)
	assert.Equal(t, eventbus.ChannelMessageSentEvent{
		Channel:      "telegram",
		SessionKey:   "telegram:3003:4004",
		ResponseText: "An error occurred: runner failed",
	}, initChannelsDisabledDoesNotCreateAdaptersWithoutSentTimestamp((*sent)[0]))
}

func initChannelsDisabledDoesNotCreateAdaptersChannelConfig() *config.Config {
	return &config.Config{
		Agent: config.AgentConfig{
			RequestTimeout: time.Second,
		},
	}
}

func initChannelsDisabledDoesNotCreateAdaptersChannelApp(bus *eventbus.Bus, executor *initChannelsDisabledDoesNotCreateAdaptersChannelExecutor) *App {
	return &App{
		Config:   initChannelsDisabledDoesNotCreateAdaptersChannelConfig(),
		EventBus: bus,
		TurnRunner: turnrunner.New(
			turnrunner.Config{HardCeiling: time.Second, StaleTimeout: -1},
			executor,
			nil,
			nil,
		),
	}
}

func initChannelsDisabledDoesNotCreateAdaptersSubscribeChannelEvents() (*eventbus.Bus, *[]eventbus.ChannelMessageReceivedEvent, *[]eventbus.ChannelMessageSentEvent) {
	bus := eventbus.New()
	received := make([]eventbus.ChannelMessageReceivedEvent, 0, 1)
	sent := make([]eventbus.ChannelMessageSentEvent, 0, 1)
	eventbus.SubscribeTyped(bus, func(e eventbus.ChannelMessageReceivedEvent) {
		received = append(received, e)
	})
	eventbus.SubscribeTyped(bus, func(e eventbus.ChannelMessageSentEvent) {
		sent = append(sent, e)
	})
	return bus, &received, &sent
}

func initChannelsDisabledDoesNotCreateAdaptersWithoutReceivedTimestamp(e eventbus.ChannelMessageReceivedEvent) eventbus.ChannelMessageReceivedEvent {
	e.Timestamp = time.Time{}
	return e
}

func initChannelsDisabledDoesNotCreateAdaptersWithoutSentTimestamp(e eventbus.ChannelMessageSentEvent) eventbus.ChannelMessageSentEvent {
	e.Timestamp = time.Time{}
	return e
}

type initChannelsDisabledDoesNotCreateAdaptersChannelCall struct {
	sessionID string
	input     string
}

type initChannelsDisabledDoesNotCreateAdaptersChannelExecutor struct {
	response string
	err      error
	calls    []initChannelsDisabledDoesNotCreateAdaptersChannelCall
}

func (e *initChannelsDisabledDoesNotCreateAdaptersChannelExecutor) RunStreamingDetailed(
	_ context.Context,
	sessionID string,
	input string,
	_ adk.ChunkCallback,
	_ ...adk.RunOption,
) (adk.RunReport, error) {
	e.calls = append(e.calls, initChannelsDisabledDoesNotCreateAdaptersChannelCall{sessionID: sessionID, input: input})
	if e.err != nil {
		return adk.RunReport{}, e.err
	}
	return adk.RunReport{Response: e.response}, nil
}
