package app

import (
	"context"
	"math/big"
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/provenance"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/langoai/lango/internal/types"
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
