package app

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/economy/escrow/hub"
	"github.com/langoai/lango/internal/economy/negotiation"
	p2pproto "github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWave23HandleNegotiateProtocolProposeInvalidPriceDefaultsToZero(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	engine := negotiation.New(config.NegotiationConfig{})

	got, err := handleNegotiateProtocol(ctx, engine, "did:local", "did:peer", p2pproto.NegotiatePayload{
		Action:   string(negotiation.ActionPropose),
		ToolName: "search",
		Price:    "not-a-number",
	})

	require.NoError(t, err)
	sessionID, ok := got["sessionId"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, sessionID)
	assert.Equal(t, string(negotiation.PhaseProposed), got["phase"])

	session, err := engine.Get(sessionID)
	require.NoError(t, err)
	require.NotNil(t, session.CurrentTerms)
	assert.Equal(t, "did:peer", session.InitiatorDID)
	assert.Equal(t, "did:local", session.ResponderDID)
	assert.Equal(t, "search", session.CurrentTerms.ToolName)
	assert.Zero(t, session.CurrentTerms.Price.Sign())
}

func TestWave23HandleNegotiateProtocolCounterReturnsRoundAndReason(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	engine := negotiation.New(config.NegotiationConfig{})
	session, err := engine.Propose(ctx, "did:peer", "did:local", negotiation.Terms{
		ToolName: "code_review",
		Price:    big.NewInt(100),
	})
	require.NoError(t, err)

	got, err := handleNegotiateProtocol(ctx, engine, "did:local", "did:peer", p2pproto.NegotiatePayload{
		Action:    string(negotiation.ActionCounter),
		SessionID: session.ID,
		ToolName:  "code_review",
		Price:     "75",
		Reason:    "preferred price",
	})

	require.NoError(t, err)
	assert.Equal(t, session.ID, got["sessionId"])
	assert.Equal(t, string(negotiation.PhaseCountered), got["phase"])
	assert.Equal(t, 2, got["round"])

	updated, err := engine.Get(session.ID)
	require.NoError(t, err)
	require.Len(t, updated.Proposals, 2)
	assert.Equal(t, big.NewInt(75), updated.CurrentTerms.Price)
	assert.Equal(t, "preferred price", updated.Proposals[1].Reason)
}

func TestWave23HandleNegotiateProtocolAcceptAndRejectTerminalPhases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	acceptEngine := negotiation.New(config.NegotiationConfig{})
	acceptSession, err := acceptEngine.Propose(ctx, "did:peer", "did:local", negotiation.Terms{
		ToolName: "summarize",
		Price:    big.NewInt(42),
	})
	require.NoError(t, err)

	accepted, err := handleNegotiateProtocol(ctx, acceptEngine, "did:local", "did:peer", p2pproto.NegotiatePayload{
		Action:    string(negotiation.ActionAccept),
		SessionID: acceptSession.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, acceptSession.ID, accepted["sessionId"])
	assert.Equal(t, string(negotiation.PhaseAccepted), accepted["phase"])

	rejectEngine := negotiation.New(config.NegotiationConfig{})
	rejectSession, err := rejectEngine.Propose(ctx, "did:peer", "did:local", negotiation.Terms{
		ToolName: "summarize",
		Price:    big.NewInt(42),
	})
	require.NoError(t, err)

	rejected, err := handleNegotiateProtocol(ctx, rejectEngine, "did:local", "did:peer", p2pproto.NegotiatePayload{
		Action:    string(negotiation.ActionReject),
		SessionID: rejectSession.ID,
		Reason:    "too expensive",
	})
	require.NoError(t, err)
	assert.Equal(t, rejectSession.ID, rejected["sessionId"])
	assert.Equal(t, string(negotiation.PhaseRejected), rejected["phase"])
}

func TestWave23HandleNegotiateProtocolErrorsForUnknownActionAndMissingSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	engine := negotiation.New(config.NegotiationConfig{})

	got, err := handleNegotiateProtocol(ctx, engine, "did:local", "did:peer", p2pproto.NegotiatePayload{
		Action: "unknown",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, negotiation.ErrSessionNotFound)

	got, err = handleNegotiateProtocol(ctx, engine, "did:local", "did:peer", p2pproto.NegotiatePayload{
		Action:    string(negotiation.ActionAccept),
		SessionID: "missing-session",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, negotiation.ErrSessionNotFound))
}

func TestWave23SelectSettlerReturnsNoopWithoutPaymentComponents(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Economy.Escrow.OnChain.Enabled = true
	cfg.Economy.Escrow.OnChain.Mode = "hub"
	cfg.Economy.Escrow.OnChain.HubAddress = "0x1111111111111111111111111111111111111111"

	settler := selectSettler(cfg, nil, nil)

	_, ok := settler.(escrow.NoopSettler)
	assert.True(t, ok)
}

func TestWave23SelectSettlerFallsBackToCustodianWhenOnChainAddressMissing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		giveMode string
	}{
		{giveMode: "hub"},
		{giveMode: "vault"},
	}

	for _, tt := range tests {
		t.Run(tt.giveMode, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Economy.Escrow.OnChain.Enabled = true
			cfg.Economy.Escrow.OnChain.Mode = tt.giveMode

			settler := selectSettler(cfg, &paymentComponents{chainID: 8453}, nil)

			_, ok := settler.(*escrow.USDCSettler)
			assert.True(t, ok)
		})
	}
}

func TestWave23SelectSettlerUsesConfiguredOnChainSettlers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		giveName  string
		giveMode  string
		configure func(*config.EscrowOnChainConfig)
		wantHub   bool
	}{
		{
			giveName: "hub",
			giveMode: "hub",
			configure: func(oc *config.EscrowOnChainConfig) {
				oc.HubAddress = "0x1111111111111111111111111111111111111111"
			},
			wantHub: true,
		},
		{
			giveName: "vault",
			giveMode: "vault",
			configure: func(oc *config.EscrowOnChainConfig) {
				oc.VaultFactoryAddress = "0x2222222222222222222222222222222222222222"
				oc.VaultImplementation = "0x3333333333333333333333333333333333333333"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.giveName, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Economy.Escrow.OnChain.Enabled = true
			cfg.Economy.Escrow.OnChain.Mode = tt.giveMode
			cfg.Economy.Escrow.OnChain.TokenAddress = "0x4444444444444444444444444444444444444444"
			tt.configure(&cfg.Economy.Escrow.OnChain)

			settler := selectSettler(cfg, &paymentComponents{chainID: 8453}, nil)

			if tt.wantHub {
				_, ok := settler.(*hub.HubSettler)
				assert.True(t, ok)
				return
			}
			_, ok := settler.(*hub.VaultSettler)
			assert.True(t, ok)
		})
	}
}

func TestWave23CatalogSourceAdapterCoversNilAndModeFilteredSections(t *testing.T) {
	t.Parallel()

	assert.Empty(t, (&catalogSourceAdapter{}).BuildToolCatalogSection(""))

	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{Name: "alpha", Description: "alpha tools", Enabled: true})
	catalog.RegisterCategory(toolcatalog.Category{Name: "disabled", Description: "disabled tools", ConfigKey: "tools.disabled", Enabled: false})
	catalog.Register("alpha", []*agent.Tool{
		wave23Tool("alpha_one", agent.ExposureDefault),
		wave23Tool("alpha_two", agent.ExposureAlwaysVisible),
		wave23Tool("alpha_deferred", agent.ExposureDeferred),
	})
	catalog.Register("disabled", []*agent.Tool{wave23Tool("disabled_one", agent.ExposureDefault)})

	cfg := config.DefaultConfig()
	cfg.Modes = map[string]config.SessionMode{
		"wave23": {
			Name:  "wave23",
			Tools: []string{"alpha_two", "alpha_deferred"},
		},
	}
	adapter := &catalogSourceAdapter{catalog: catalog, cfg: cfg}

	defaultSection := adapter.BuildToolCatalogSection("")
	assert.Contains(t, defaultSection, "## Available Tool Categories")
	assert.Contains(t, defaultSection, "alpha_one")
	assert.Contains(t, defaultSection, "alpha_two")
	assert.NotContains(t, defaultSection, "alpha_deferred")
	assert.Contains(t, defaultSection, "Disabled categories (enable via config): disabled (tools.disabled)")
	assert.Contains(t, defaultSection, "Additional 1 specialized tools available via builtin_search")

	modeSection := adapter.BuildToolCatalogSection("wave23")
	assert.Contains(t, modeSection, "## Tools Available in `wave23` Mode")
	assert.NotContains(t, modeSection, "alpha_one")
	assert.Contains(t, modeSection, "alpha_two")
	assert.NotContains(t, modeSection, "alpha_deferred")
	assert.Contains(t, modeSection, "Only tools in this mode's allowlist are available")

	missingModeSection := adapter.BuildToolCatalogSection("missing-mode")
	assert.Contains(t, missingModeSection, "## Tools Available in `missing-mode` Mode")
	assert.Contains(t, missingModeSection, "alpha_one")
	assert.Contains(t, missingModeSection, "alpha_two")
}

func TestWave23CatalogSourceAdapterTruncatesLongCategoryToolLists(t *testing.T) {
	t.Parallel()

	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{Name: "bulk", Description: "bulk tools", Enabled: true})
	tools := make([]*agent.Tool, 0, 10)
	for i := 0; i < 10; i++ {
		tools = append(tools, wave23Tool("bulk_"+string(rune('a'+i)), agent.ExposureDefault))
	}
	catalog.Register("bulk", tools)

	section := (&catalogSourceAdapter{catalog: catalog}).BuildToolCatalogSection("")

	assert.Contains(t, section, "... +2 more")
	assert.Contains(t, section, "bulk_a")
	assert.NotContains(t, section, "bulk_j")
}

func TestWave23AutomationPromptSectionIncludesOnlyEnabledCapabilities(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Cron.Enabled = true
	cfg.Background.Enabled = false
	cfg.Workflow.Enabled = true

	section := buildAutomationPromptSection(cfg)
	content := section.Render()

	assert.Contains(t, content, "### Cron Scheduling")
	assert.Contains(t, content, "### Workflow Pipelines")
	assert.NotContains(t, content, "### Background Tasks")
	assert.True(t, strings.Contains(content, "NEVER use exec"))
}

func wave23Tool(name string, exposure agent.ExposurePolicy) *agent.Tool {
	return &agent.Tool{
		Name:        name,
		Description: name + " description",
		Capability:  agent.ToolCapability{Exposure: exposure},
	}
}
