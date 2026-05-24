package app

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/langoai/lango/internal/channels/discord"
	slackchannel "github.com/langoai/lango/internal/channels/slack"
	"github.com/langoai/lango/internal/channels/telegram"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/provenance"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/langoai/lango/internal/types"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDeliveryTarget(t *testing.T) {
	tests := []struct {
		give       string
		wantType   types.ChannelType
		wantTarget string
	}{
		{
			give:       "telegram:123456789",
			wantType:   types.ChannelTelegram,
			wantTarget: "123456789",
		},
		{
			give:       "discord:channel-id-here",
			wantType:   types.ChannelDiscord,
			wantTarget: "channel-id-here",
		},
		{
			give:       "slack:C12345",
			wantType:   types.ChannelSlack,
			wantTarget: "C12345",
		},
		{
			give:       "telegram",
			wantType:   types.ChannelTelegram,
			wantTarget: "",
		},
		{
			give:       "discord",
			wantType:   types.ChannelDiscord,
			wantTarget: "",
		},
		{
			give:       "slack",
			wantType:   types.ChannelSlack,
			wantTarget: "",
		},
		{
			give:       "  TELEGRAM:999  ",
			wantType:   types.ChannelTelegram,
			wantTarget: "999",
		},
		{
			give:       "  Discord  ",
			wantType:   types.ChannelDiscord,
			wantTarget: "",
		},
		{
			give:       "unknown:abc",
			wantType:   types.ChannelType("unknown"),
			wantTarget: "abc",
		},
		{
			give:       "",
			wantType:   types.ChannelType(""),
			wantTarget: "",
		},
		{
			give:       "telegram:chat:extra",
			wantType:   types.ChannelTelegram,
			wantTarget: "chat:extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			gotType, gotTarget := parseDeliveryTarget(tt.give)
			assert.Equal(t, tt.wantType, gotType, "channel type")
			assert.Equal(t, tt.wantTarget, gotTarget, "target ID")
		})
	}
}

func TestChannelSender_NoConfiguredChannelReportsUnavailable(t *testing.T) {
	sender := newChannelSender(&App{
		Config: &config.Config{},
	})

	err := sender.SendMessage(context.Background(), "telegram:1234", "hello")

	require.Error(t, err)
	assert.Contains(t, err.Error(), `channel "telegram:1234" not available`)
}

func TestChannelSender_StartTypingReturnsNoopWhenUnavailable(t *testing.T) {
	sender := newChannelSender(&App{
		Config: &config.Config{},
	})

	stop, err := sender.StartTyping(context.Background(), "slack:C123")

	require.NoError(t, err)
	require.NotNil(t, stop)
	assert.NotPanics(t, stop)
}

func TestChannelSender_TelegramDeliveryAndTypingBranches(t *testing.T) {
	t.Parallel()

	bot := &recordingTelegramBot{}
	tg, err := telegram.New(telegram.Config{BotToken: "token", Bot: bot})
	require.NoError(t, err)
	sender := newChannelSender(&App{
		Config: &config.Config{
			Channels: config.ChannelsConfig{
				Telegram: config.TelegramConfig{Allowlist: []int64{4321}},
			},
		},
		Channels: []Channel{tg},
	})

	require.ErrorContains(t, sender.SendMessage(context.Background(), "telegram:not-a-number", "hello"), "parse telegram chat ID")
	require.NoError(t, sender.SendMessage(context.Background(), "telegram", "hello"))
	require.Len(t, bot.sent, 1)
	assert.Equal(t, int64(4321), bot.sent[0].ChatID)
	assert.Equal(t, "hello", bot.sent[0].Text)

	stop, err := sender.StartTyping(context.Background(), "telegram:bad")
	require.ErrorContains(t, err, "parse telegram chat ID")
	assert.NotPanics(t, stop)

	stop, err = sender.StartTyping(context.Background(), "telegram")
	require.NoError(t, err)
	stop()
	stop()
	assert.NotEmpty(t, bot.requests)
}

func TestChannelSender_AvailableChannelsValidateRequiredTargets(t *testing.T) {
	t.Parallel()

	dc, err := discord.New(discord.Config{BotToken: "token", Session: &recordingDiscordSession{state: discordgo.NewState()}})
	require.NoError(t, err)
	sl, err := slackchannel.New(slackchannel.Config{
		BotToken: "xoxb-token",
		AppToken: "xapp-token",
		Client:   &recordingSlackClient{},
		Socket:   recordingSlackSocket{events: make(chan socketmode.Event)},
	})
	require.NoError(t, err)
	tg, err := telegram.New(telegram.Config{BotToken: "token", Bot: &recordingTelegramBot{sendErr: errors.New("send failed")}})
	require.NoError(t, err)

	sender := newChannelSender(&App{
		Config: &config.Config{},
		Channels: []Channel{
			dc,
			sl,
			tg,
		},
	})

	require.ErrorContains(t, sender.SendMessage(context.Background(), "discord", "hello"), "discord delivery requires a channel ID")
	require.ErrorContains(t, sender.SendMessage(context.Background(), "slack", "hello"), "slack delivery requires a channel ID")
	require.ErrorContains(t, sender.SendMessage(context.Background(), "telegram", "hello"), "telegram delivery requires a chat ID")

	stop, err := sender.StartTyping(context.Background(), "discord")
	require.NoError(t, err)
	assert.NotPanics(t, stop)

	stop, err = sender.StartTyping(context.Background(), "slack")
	require.NoError(t, err)
	assert.NotPanics(t, stop)

	stop, err = sender.StartTyping(context.Background(), "telegram")
	require.NoError(t, err)
	assert.NotPanics(t, stop)
}

func TestChannelSender_FirstTelegramChatIDUsesAllowlist(t *testing.T) {
	sender := newChannelSender(&App{
		Config: &config.Config{
			Channels: config.ChannelsConfig{
				Telegram: config.TelegramConfig{
					Allowlist: []int64{42, 99},
				},
			},
		},
	})

	assert.Equal(t, int64(42), sender.firstTelegramChatID())

	sender.app.Config.Channels.Telegram.Allowlist = nil
	assert.Zero(t, sender.firstTelegramChatID())
}

func TestProvenanceAuthorFromContext(t *testing.T) {
	t.Run("agent identity wins over fallback DID", func(t *testing.T) {
		ctx := toolchain.WithAgentName(context.Background(), "planner")

		authorType, authorID := provenanceAuthorFromContext(ctx, "did:lango:fallback")

		assert.Equal(t, provenance.AuthorAgent, authorType)
		assert.Equal(t, "planner", authorID)
	})

	t.Run("user agent name falls back to remote peer DID", func(t *testing.T) {
		ctx := toolchain.WithAgentName(context.Background(), "user")

		authorType, authorID := provenanceAuthorFromContext(ctx, "did:lango:remote")

		assert.Equal(t, provenance.AuthorRemotePeer, authorType)
		assert.Equal(t, "did:lango:remote", authorID)
	})

	t.Run("empty context and fallback identify unknown human", func(t *testing.T) {
		authorType, authorID := provenanceAuthorFromContext(context.Background(), "")

		assert.Equal(t, provenance.AuthorHuman, authorType)
		assert.Equal(t, "unknown", authorID)
	})
}

func TestWalletBundleSignerDelegatesAndReportsAlgorithm(t *testing.T) {
	wallet := &recordingWalletProvider{
		signature: []byte("signed-payload"),
	}
	signer := &walletBundleSigner{wp: wallet}

	got, err := signer.Sign(context.Background(), []byte("payload"))

	require.NoError(t, err)
	assert.Equal(t, []byte("signed-payload"), got)
	assert.Equal(t, []byte("payload"), wallet.signedMessage)
	assert.Equal(t, security.AlgorithmSecp256k1Keccak256, signer.Algorithm())
}

type recordingWalletProvider struct {
	signature     []byte
	signedMessage []byte
}

func (w *recordingWalletProvider) Address(context.Context) (string, error) {
	return "", nil
}

func (w *recordingWalletProvider) Balance(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (w *recordingWalletProvider) SignTransaction(_ context.Context, rawTx []byte) ([]byte, error) {
	return append([]byte(nil), rawTx...), nil
}

func (w *recordingWalletProvider) SignMessage(_ context.Context, message []byte) ([]byte, error) {
	w.signedMessage = append([]byte(nil), message...)
	return append([]byte(nil), w.signature...), nil
}

func (w *recordingWalletProvider) PublicKey(context.Context) ([]byte, error) {
	return nil, nil
}

type recordingTelegramBot struct {
	sent     []tgbotapi.MessageConfig
	requests []tgbotapi.Chattable
	sendErr  error
}

func (b *recordingTelegramBot) GetUpdatesChan(tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return make(tgbotapi.UpdatesChannel)
}

func (b *recordingTelegramBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	if b.sendErr != nil {
		return tgbotapi.Message{}, b.sendErr
	}
	if msg, ok := c.(tgbotapi.MessageConfig); ok {
		b.sent = append(b.sent, msg)
	}
	return tgbotapi.Message{}, nil
}

func (b *recordingTelegramBot) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	b.requests = append(b.requests, c)
	return &tgbotapi.APIResponse{Ok: true}, nil
}

func (b *recordingTelegramBot) GetFile(tgbotapi.FileConfig) (tgbotapi.File, error) {
	return tgbotapi.File{}, nil
}

func (b *recordingTelegramBot) StopReceivingUpdates() {}

func (b *recordingTelegramBot) GetSelf() tgbotapi.User {
	return tgbotapi.User{UserName: "test-bot"}
}

type recordingDiscordSession struct {
	state *discordgo.State
}

func (s *recordingDiscordSession) Open() error  { return nil }
func (s *recordingDiscordSession) Close() error { return nil }
func (s *recordingDiscordSession) AddHandler(interface{}) func() {
	return func() {}
}
func (s *recordingDiscordSession) ChannelMessageSend(string, string, ...discordgo.RequestOption) (*discordgo.Message, error) {
	return &discordgo.Message{}, nil
}
func (s *recordingDiscordSession) ChannelMessageSendComplex(string, *discordgo.MessageSend, ...discordgo.RequestOption) (*discordgo.Message, error) {
	return &discordgo.Message{}, nil
}
func (s *recordingDiscordSession) ChannelMessageEditComplex(*discordgo.MessageEdit, ...discordgo.RequestOption) (*discordgo.Message, error) {
	return &discordgo.Message{}, nil
}
func (s *recordingDiscordSession) ChannelTyping(string, ...discordgo.RequestOption) error {
	return nil
}
func (s *recordingDiscordSession) InteractionRespond(*discordgo.Interaction, *discordgo.InteractionResponse, ...discordgo.RequestOption) error {
	return nil
}
func (s *recordingDiscordSession) ApplicationCommandCreate(string, string, *discordgo.ApplicationCommand, ...discordgo.RequestOption) (*discordgo.ApplicationCommand, error) {
	return &discordgo.ApplicationCommand{}, nil
}
func (s *recordingDiscordSession) GetState() *discordgo.State {
	return s.state
}

type recordingSlackClient struct{}

func (recordingSlackClient) AuthTest() (*slack.AuthTestResponse, error) {
	return &slack.AuthTestResponse{}, nil
}

func (recordingSlackClient) PostMessage(string, ...slack.MsgOption) (string, string, error) {
	return "", "ts", nil
}

func (recordingSlackClient) UpdateMessage(string, string, ...slack.MsgOption) (string, string, string, error) {
	return "", "", "", nil
}

func (recordingSlackClient) DeleteMessage(string, string) (string, string, error) {
	return "", "", nil
}

type recordingSlackSocket struct {
	events chan socketmode.Event
}

func (recordingSlackSocket) Run() error { return nil }

func (recordingSlackSocket) Ack(socketmode.Request, ...interface{}) {}

func (s recordingSlackSocket) Events() <-chan socketmode.Event {
	return s.events
}
