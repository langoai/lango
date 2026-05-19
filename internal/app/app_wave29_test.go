package app

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/p2p/paygate"
	p2pproto "github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/payment/contracts"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/toolcatalog"
)

type wave29Submitter struct {
	taskID    string
	submitErr error
}

func (s *wave29Submitter) Submit(context.Context, string, background.Origin) (string, error) {
	if s.submitErr != nil {
		return "", s.submitErr
	}
	return s.taskID, nil
}

type wave29CancelFailingSubmitter struct {
	wave29Submitter
	cancelledID string
	cancelErr   error
}

func (s *wave29CancelFailingSubmitter) Cancel(taskID string) error {
	s.cancelledID = taskID
	return s.cancelErr
}

func TestWave29MissionAwareSubmitterPropagatesSubmitAndCancelErrors(t *testing.T) {
	t.Parallel()

	t.Run("submit error skips mission link", func(t *testing.T) {
		t.Parallel()

		linker := &failingMissionBackgroundLinker{}
		wantErr := errors.New("submit failed")
		submitter := &missionAwareSubmitter{
			base:   &wave29Submitter{submitErr: wantErr},
			linker: linker,
		}

		taskID, err := submitter.Submit(context.Background(), "run diagnostics", background.Origin{Channel: "test"})

		require.ErrorIs(t, err, wantErr)
		assert.Empty(t, taskID)
		assert.Zero(t, linker.calls)
	})

	t.Run("cancel error is included after link failure", func(t *testing.T) {
		t.Parallel()

		cancelErr := errors.New("cancel failed")
		base := &wave29CancelFailingSubmitter{
			wave29Submitter: wave29Submitter{taskID: "task-wave29"},
			cancelErr:       cancelErr,
		}
		submitter := &missionAwareSubmitter{
			base:   base,
			linker: &failingMissionBackgroundLinker{},
		}

		taskID, err := submitter.Submit(context.Background(), "run diagnostics", background.Origin{Channel: "test"})

		require.Error(t, err)
		assert.Empty(t, taskID)
		assert.Equal(t, "task-wave29", base.cancelledID)
		assert.ErrorContains(t, err, "attach spawned child execution to mission")
		assert.ErrorContains(t, err, `cancel submitted task "task-wave29"`)
		assert.ErrorContains(t, err, cancelErr.Error())
	})
}

func TestWave29NetworkModuleInitMarksP2PDisabledWithoutPaymentRequirement(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = false
	cfg.P2P.Enabled = false
	cfg.P2P.Workspace.Enabled = true
	cfg.Economy.Enabled = true

	result, err := (&networkModule{cfg: cfg}).Init(
		context.Background(),
		staticResolver{appinit.ProvidesSupervisor: &foundationValues{}},
	)
	require.NoError(t, err)

	assert.NotEmpty(t, result.Tools)
	assert.Nil(t, result.Values[appinit.ProvidesPayment])
	assert.Nil(t, result.Values[appinit.ProvidesP2P])
	assert.NotNil(t, result.Values[appinit.ProvidesEconomy])
	p2pEntry := requireCatalogEntry(t, result.CatalogEntries, "p2p")
	assert.False(t, p2pEntry.Enabled)
	assert.Equal(t, "P2P networking (disabled)", p2pEntry.Description)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "economy").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "workspace").Enabled)
}

func TestWave29CatalogSourceAdapterTruncatesVisibleToolsAndReportsDeferredCount(t *testing.T) {
	t.Parallel()

	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{
		Name:        "bulk",
		Description: "Bulk tools",
		Enabled:     true,
	})
	var tools []*agent.Tool
	for i := 0; i < 10; i++ {
		tools = append(tools, &agent.Tool{
			Name:        fmt.Sprintf("bulk_tool_%02d", i),
			Description: "visible",
			SafetyLevel: agent.SafetyLevelSafe,
		})
	}
	tools = append(tools, &agent.Tool{
		Name:        "bulk_deferred",
		Description: "deferred",
		SafetyLevel: agent.SafetyLevelSafe,
		Capability:  agent.ToolCapability{Exposure: agent.ExposureDeferred},
	})
	catalog.Register("bulk", tools)

	section := (&catalogSourceAdapter{catalog: catalog}).BuildToolCatalogSection("")

	assert.Contains(t, section, "bulk_tool_00")
	assert.Contains(t, section, "bulk_tool_07")
	assert.NotContains(t, section, "bulk_tool_08")
	assert.Contains(t, section, "... +2 more")
	assert.Contains(t, section, "Additional 1 specialized tools available via builtin_search.")
}

func TestWave29MetaToolsUseSettlementRuntimeAsPartialFallback(t *testing.T) {
	t.Parallel()

	tools := buildMetaToolsWithRuntimes(
		nil,
		nil,
		nil,
		config.SkillConfig{},
		nil,
		receipts.NewStore(),
		nil,
		&fakeSettlementExecutionRuntime{},
		nil,
		nil,
		nil,
		nil,
	)

	settlementTool := findTool(tools, "execute_settlement")
	partialTool := findTool(tools, "execute_partial_settlement")
	require.NotNil(t, settlementTool)
	require.NotNil(t, partialTool)
	assert.Equal(t, agent.ActivityWrite, partialTool.Capability.Activity)
	assert.Equal(t, "knowledge", partialTool.Capability.Category)
}

func TestWave29CreateDisputeReadyReceiptReportsValidationError(t *testing.T) {
	t.Parallel()

	got, err := createDisputeReadyReceipt(context.Background(), receipts.NewStore(), receipts.CreateSubmissionInput{})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "create dispute-ready receipt")
	assert.ErrorContains(t, err, "transaction_id is required")
}

func TestWave29PayGateAdapterMapsQuoteContractSellerAndExpiry(t *testing.T) {
	t.Parallel()

	const chainID int64 = 84532
	localAddr := "0x1234567890abcdef1234567890abcdef12345678"
	usdcAddr, err := contracts.LookupUSDC(chainID)
	require.NoError(t, err)
	gate := paygate.New(paygate.Config{
		PricingFn: func(string) (string, bool) {
			return "2.75", false
		},
		LocalAddr: localAddr,
		ChainID:   chainID,
		USDCAddr:  usdcAddr,
		Logger:    testLog(),
	})

	got, err := (&payGateAdapter{gate: gate}).Check("did:lango:peer", "paid_tool", map[string]interface{}{})

	require.NoError(t, err)
	assert.Equal(t, p2pproto.PayGateResult{
		Status: string(paygate.StatusPaymentRequired),
		PriceQuote: map[string]interface{}{
			"toolName":     "paid_tool",
			"price":        "2.75",
			"currency":     "USDC",
			"usdcContract": usdcAddr.Hex(),
			"chainId":      chainID,
			"sellerAddr":   localAddr,
			"quoteExpiry":  got.PriceQuote["quoteExpiry"],
			"isFree":       false,
		},
	}, got)
	assert.NotEmpty(t, got.PriceQuote["quoteExpiry"])
}

func TestWave29InitZKPDisabledReturnsNilAndAttestationCreatesProver(t *testing.T) {
	t.Parallel()

	disabled := config.DefaultConfig()
	disabled.P2P.ZKHandshake = false
	disabled.P2P.ZKAttestation = false
	assert.Nil(t, initZKP(disabled))

	cfg := config.DefaultConfig()
	cfg.P2P.ZKHandshake = false
	cfg.P2P.ZKAttestation = true
	cfg.P2P.ZKP.ProvingScheme = "unsupported-wave29-scheme"

	assert.NotNil(t, initZKP(cfg))
}

func TestWave29PayGateAdapterMapsVerifiedAuthWithoutQuote(t *testing.T) {
	t.Parallel()

	const chainID int64 = 84532
	localAddr := "0x1234567890abcdef1234567890abcdef12345678"
	usdcAddr, err := contracts.LookupUSDC(chainID)
	require.NoError(t, err)
	gate := paygate.New(paygate.Config{
		PricingFn: func(string) (string, bool) {
			return "0.50", false
		},
		LocalAddr: localAddr,
		ChainID:   chainID,
		USDCAddr:  usdcAddr,
		Logger:    testLog(),
	})

	got, err := (&payGateAdapter{gate: gate}).Check("did:lango:peer", "paid_tool", map[string]interface{}{
		"paymentAuth": makeAppP2PPaymentAuth(localAddr, big.NewInt(500000)),
	})

	require.NoError(t, err)
	assert.Equal(t, string(paygate.StatusVerified), got.Status)
	assert.NotNil(t, got.Auth)
	assert.Nil(t, got.PriceQuote)
}
