package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/receipts"
)

func TestInitP2PEarlyExitBranchesAvoidNetworkState(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.KeyDir = filepath.Join(t.TempDir(), "p2p-keys")
	cfg.P2P.ListenAddrs = []string{"/ip4/127.0.0.1/tcp/0"}

	cfg.P2P.Enabled = false
	assert.Nil(t, initP2P(cfg, nil, nil, nil, nil, nil, nil, nil, ""))
	assert.NoDirExists(t, cfg.P2P.KeyDir)

	cfg.P2P.Enabled = true
	assert.Nil(t, initP2P(cfg, nil, nil, nil, nil, nil, nil, nil, ""))
	assert.NoDirExists(t, cfg.P2P.KeyDir)
}

func TestNetworkModuleP2PEnabledWithoutPaymentReportsPaymentRequired(t *testing.T) {
	t.Parallel()

	cfg := foundationModuleInitContinuesAfterSanitizerPatternErrorModuleConfig(t)
	cfg.P2P.Enabled = true
	module := &networkModule{cfg: cfg, bus: eventbus.New(), app: &App{ctx: context.Background()}}
	require.True(t, module.Enabled())

	result, err := module.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{
			ReceiptStore: receipts.NewStore(),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Nil(t, result.Values[appinit.ProvidesPayment])
	assert.Nil(t, result.Values[appinit.ProvidesP2P])
	assert.Nil(t, result.Values[appinit.ProvidesWorkspace])

	p2pEntry := requireCatalogEntry(t, result.CatalogEntries, "p2p")
	assert.False(t, p2pEntry.Enabled)
	assert.Contains(t, p2pEntry.Description, "payment required")
	assert.Nil(t, foundationModuleInitContinuesAfterSanitizerPatternErrorCatalogEntry(result.CatalogEntries, "workspace"))
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "smartaccount").Enabled)
}

func TestParseEscrowExecutionInputReportsRequiredFieldErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		give      map[string]interface{}
		wantError string
	}{
		{
			name:      "missing buyer DID",
			give:      map[string]interface{}{},
			wantError: "missing escrow_buyer_did parameter",
		},
		{
			name: "missing seller DID",
			give: map[string]interface{}{
				"escrow_buyer_did": "did:lango:buyer",
			},
			wantError: "missing escrow_seller_did parameter",
		},
		{
			name: "missing reason",
			give: map[string]interface{}{
				"escrow_buyer_did":  "did:lango:buyer",
				"escrow_seller_did": "did:lango:seller",
			},
			wantError: "missing escrow_reason parameter",
		},
		{
			name: "milestones must be array",
			give: map[string]interface{}{
				"escrow_buyer_did":  "did:lango:buyer",
				"escrow_seller_did": "did:lango:seller",
				"escrow_reason":     "initP2PEarlyExitBranchesAvoidNetworkState2 execution",
				"escrow_milestones": "not-array",
			},
			wantError: "escrow_milestones must be an array",
		},
		{
			name: "milestone must be object",
			give: map[string]interface{}{
				"escrow_buyer_did":  "did:lango:buyer",
				"escrow_seller_did": "did:lango:seller",
				"escrow_reason":     "initP2PEarlyExitBranchesAvoidNetworkState2 execution",
				"escrow_milestones": []interface{}{"not-object"},
			},
			wantError: "escrow_milestones[0] must be an object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseEscrowExecutionInput(tt.give, "1.25")

			require.Error(t, err)
			assert.Empty(t, got)
			assert.ErrorContains(t, err, tt.wantError)
		})
	}
}

func TestWireBudgetManagerInvalidAllocationReturnsWithoutAdapter(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Context.ModelWindow = 8192
	cfg.Context.ResponseReserve = 1024
	cfg.Context.Allocation = config.ContextAllocationConfig{
		Knowledge: 0.50,
		RAG:       0.50,
		Memory:    0.50,
	}

	require.NotPanics(t, func() {
		wireBudgetManager(cfg, buildPromptBuilder(&cfg.Agent), nil)
	})
}
