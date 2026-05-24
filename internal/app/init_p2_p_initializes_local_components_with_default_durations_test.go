package app

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/automation"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/escrowrefund"
	"github.com/langoai/lango/internal/escrowrelease"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/p2p/zkp"
	"github.com/langoai/lango/internal/postadjudicationreplay"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/testutil"
)

func TestInitP2P_InitializesLocalComponentsWithDefaultDurations(t *testing.T) {
	key, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	cfg := initP2PInitializesLocalComponentsWithDefaultDurationsP2PConfig(t)
	cfg.P2P.SessionTokenTTL = 0
	cfg.P2P.HandshakeTimeout = 0
	cfg.P2P.GossipInterval = 0
	cfg.P2P.Team.HealthCheckInterval = 0
	cfg.P2P.Team.MaxMissedHeartbeats = 0
	cfg.P2P.Pricing.Enabled = true
	cfg.P2P.Pricing.PerQuery = "0.10"
	cfg.P2P.Pricing.ToolPrices = map[string]string{"premium_tool": "0.25"}

	components := initP2P(
		cfg,
		&wiringP2PWallet{publicKey: ethcrypto.CompressPubkey(&key.PublicKey)},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		"",
	)
	require.NotNil(t, components)
	t.Cleanup(func() { initP2PInitializesLocalComponentsWithDefaultDurationsStopP2PComponents(t, components) })

	assert.NotNil(t, components.node)
	assert.NotNil(t, components.sessions)
	assert.NotNil(t, components.handshaker)
	assert.NotNil(t, components.nonceCache)
	assert.NotNil(t, components.fw)
	assert.NotNil(t, components.gossip)
	assert.NotNil(t, components.handler)
	assert.NotNil(t, components.agentPool)
	assert.NotNil(t, components.selector)
	assert.NotNil(t, components.provider)
	assert.NotNil(t, components.coordinator)
	assert.NotNil(t, components.healthMonitor)
	assert.False(t, components.kemEnabled)
	assert.Nil(t, components.payGate)
	assert.IsType(t, &legacyLocalIdentity{}, components.identity)

	price, free := components.pricingFn("premium_tool")
	assert.Equal(t, "0.25", price)
	assert.False(t, free)
	price, free = components.pricingFn("unknown_tool")
	assert.Equal(t, "0.10", price)
	assert.False(t, free)
}

func TestInitP2P_UsesBundleIdentityWhenSigningKeyAndLangoDirAreConfigured(t *testing.T) {
	key, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	_, signingKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	cfg := initP2PInitializesLocalComponentsWithDefaultDurationsP2PConfig(t)

	components := initP2P(
		cfg,
		&wiringP2PWallet{publicKey: ethcrypto.CompressPubkey(&key.PublicKey)},
		nil,
		nil,
		nil,
		nil,
		signingKey,
		nil,
		t.TempDir(),
	)
	require.NotNil(t, components)
	t.Cleanup(func() { initP2PInitializesLocalComponentsWithDefaultDurationsStopP2PComponents(t, components) })

	bundleProvider, ok := components.identity.(*identity.BundleProvider)
	require.True(t, ok)
	assert.Equal(t, "ed25519", bundleProvider.Algorithm())
	assert.NotNil(t, bundleProvider.Bundle())
}

func TestInitZKP_ReturnsProverWhenEnabledEvenIfCircuitCompileFails(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.ZKHandshake = true
	cfg.P2P.ZKAttestation = false
	cfg.P2P.ZKP.ProofCacheDir = t.TempDir()
	cfg.P2P.ZKP.ProvingScheme = "unsupported-initP2PInitializesLocalComponentsWithDefaultDurations7"

	prover := initZKP(cfg)
	require.NotNil(t, prover)
	assert.Equal(t, zkp.ProofScheme("unsupported-initP2PInitializesLocalComponentsWithDefaultDurations7"), prover.Scheme())
}

func TestReplayDispatcherAdapter_DispatchesPromptWithSessionOrigin(t *testing.T) {
	dispatcher := &fakeAdjudicationBackgroundDispatcher{taskID: "task-initP2PInitializesLocalComponentsWithDefaultDurations7"}
	ctx := session.WithSessionKey(context.Background(), "telegram:room-7:user-42")
	req := postadjudicationreplay.BackgroundDispatchRequest{
		Prompt:               "retry adjudication",
		TransactionReceiptID: "tx-123",
		SubmissionReceiptID:  "sub-123",
		EscrowReference:      "escrow-123",
		Outcome:              receipts.EscrowAdjudicationRelease,
	}

	got, err := (replayDispatcherAdapter{dispatcher: dispatcher}).Dispatch(ctx, req)
	require.NoError(t, err)

	assert.Equal(t, postadjudicationreplay.BackgroundDispatchReceipt{
		Status:               "queued",
		TransactionReceiptID: "tx-123",
		SubmissionReceiptID:  "sub-123",
		EscrowReference:      "escrow-123",
		Outcome:              "release",
		DispatchReference:    "task-initP2PInitializesLocalComponentsWithDefaultDurations7",
	}, got)
	calls, prompt, origin := dispatcher.snapshot()
	assert.Equal(t, 1, calls)
	assert.Equal(t, "retry adjudication", prompt)
	assert.Equal(t, automation.DetectChannelFromContext(ctx), origin.Channel)
	assert.Equal(t, "telegram:room-7:user-42", origin.Session)
}

func TestEngineEscrowReleaseAndRefundAdapters_DelegateToEngine(t *testing.T) {
	ctx := context.Background()

	t.Run("release", func(t *testing.T) {
		engine := initP2PInitializesLocalComponentsWithDefaultDurationsEscrowEngine()
		entry := initP2PInitializesLocalComponentsWithDefaultDurationsCreateFundedEscrow(t, ctx, engine)
		entry, err := engine.Activate(ctx, entry.ID)
		require.NoError(t, err)
		entry, err = engine.CompleteMilestone(ctx, entry.ID, entry.Milestones[0].ID, "evidence")
		require.NoError(t, err)

		got, err := (engineEscrowReleaseRuntime{engine: engine}).Release(ctx, escrowrelease.ReleaseRequest{
			EscrowReference: entry.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, escrowrelease.ReleaseResult{Reference: entry.ID}, got)
	})

	t.Run("refund", func(t *testing.T) {
		engine := initP2PInitializesLocalComponentsWithDefaultDurationsEscrowEngine()
		entry := initP2PInitializesLocalComponentsWithDefaultDurationsCreateFundedEscrow(t, ctx, engine)
		entry, err := engine.Activate(ctx, entry.ID)
		require.NoError(t, err)
		entry, err = engine.Dispute(ctx, entry.ID, "needs refund")
		require.NoError(t, err)

		got, err := (engineEscrowRefundRuntime{engine: engine}).Refund(ctx, escrowrefund.RefundRequest{
			EscrowReference: entry.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, escrowrefund.RefundResult{Reference: entry.ID}, got)
	})
}

func TestEngineEscrowReleaseAndRefundAdapters_RejectNilEngine(t *testing.T) {
	ctx := context.Background()

	_, err := (engineEscrowReleaseRuntime{}).Release(ctx, escrowrelease.ReleaseRequest{
		EscrowReference: "escrow-123",
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "escrow engine is required")

	_, err = (engineEscrowRefundRuntime{}).Refund(ctx, escrowrefund.RefundRequest{
		EscrowReference: "escrow-123",
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "escrow engine is required")
}

func TestAutomationModuleInit_CronOnlyRegistersSchedulerAndDisabledCategories(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Cron.Enabled = true
	cfg.Background.Enabled = false
	cfg.Workflow.Enabled = false
	application := &App{Config: cfg}
	mod := &automationModule{cfg: cfg, app: application, bus: eventbus.New()}
	store := session.NewEntStoreWithClient(testutil.TestEntClient(t))

	result, err := mod.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{Store: store},
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	values, ok := result.Values[appinit.ProvidesAutomation].(*automationValues)
	require.True(t, ok)
	assert.NotNil(t, values.CronScheduler)
	assert.Nil(t, values.BackgroundManager)
	assert.Nil(t, values.WorkflowEngine)
	assert.NotNil(t, values.AgentRunStore)
	assert.NotEmpty(t, result.Tools)
	assert.NotEmpty(t, result.Components)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "cron").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "background").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "workflow").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "agent_control").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "task_tracking").Enabled)
}

func TestNetworkModuleInit_DisabledPaymentAndP2PRegisterPlainDisabledEntries(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = false
	cfg.P2P.Enabled = false
	cfg.P2P.Workspace.Enabled = false
	cfg.SmartAccount.Enabled = true
	mod := &networkModule{cfg: cfg, bus: eventbus.New()}

	result, err := mod.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{},
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Empty(t, result.Tools)
	assert.Empty(t, result.Components)
	assert.Nil(t, result.Values[appinit.ProvidesPayment])
	assert.Nil(t, result.Values[appinit.ProvidesP2P])
	assert.Nil(t, result.Values[appinit.ProvidesEconomy])
	assert.Nil(t, result.Values[appinit.ProvidesContract])
	assert.Nil(t, result.Values[appinit.ProvidesSmartAccount])
	assert.Nil(t, result.Values[appinit.ProvidesWorkspace])

	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "payment").Enabled)
	p2pEntry := requireCatalogEntry(t, result.CatalogEntries, "p2p")
	assert.False(t, p2pEntry.Enabled)
	assert.NotContains(t, p2pEntry.Description, "payment required")
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "contract").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "smartaccount").Enabled)
}

func initP2PInitializesLocalComponentsWithDefaultDurationsP2PConfig(t *testing.T) *config.Config {
	t.Helper()

	keyDir := filepath.Join(t.TempDir(), "p2p-keys")
	workspaceFile := filepath.Join(t.TempDir(), "workspace-file")
	require.NoError(t, os.WriteFile(workspaceFile, []byte("not a directory"), 0o600))

	cfg := config.DefaultConfig()
	cfg.A2A.AgentName = "initP2PInitializesLocalComponentsWithDefaultDurations7-agent"
	cfg.P2P.Enabled = true
	cfg.P2P.ListenAddrs = []string{"/ip4/127.0.0.1/tcp/0"}
	//nolint:staticcheck // intentional: regression test pins legacy KeyDir behavior.
	cfg.P2P.KeyDir = keyDir
	cfg.P2P.EnableRelay = false
	cfg.P2P.EnableMDNS = false
	cfg.P2P.ZKHandshake = false
	cfg.P2P.ZKAttestation = false
	cfg.P2P.EnablePQHandshake = false
	cfg.P2P.Workspace.DataDir = workspaceFile
	return cfg
}

func initP2PInitializesLocalComponentsWithDefaultDurationsStopP2PComponents(t *testing.T, components *p2pComponents) {
	t.Helper()

	if components == nil {
		return
	}
	if components.gossip != nil {
		components.gossip.Stop()
	}
	if components.nonceCache != nil {
		components.nonceCache.Stop()
	}
	if components.node != nil {
		require.NoError(t, components.node.Stop())
	}
}

func initP2PInitializesLocalComponentsWithDefaultDurationsEscrowEngine() *escrow.Engine {
	return escrow.NewEngine(escrow.NewMemoryStore(), escrow.NoopSettler{}, escrow.DefaultEngineConfig())
}

func initP2PInitializesLocalComponentsWithDefaultDurationsCreateFundedEscrow(t *testing.T, ctx context.Context, engine *escrow.Engine) *escrow.EscrowEntry {
	t.Helper()

	entry, err := engine.Create(ctx, escrow.CreateRequest{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    big.NewInt(100),
		Reason:    "initP2PInitializesLocalComponentsWithDefaultDurations7 escrow adapter test",
		Milestones: []escrow.MilestoneRequest{
			{Description: "deliverable", Amount: big.NewInt(100)},
		},
	})
	require.NoError(t, err)
	entry, err = engine.Fund(ctx, entry.ID)
	require.NoError(t, err)
	return entry
}
