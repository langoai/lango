package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/p2p/handshake"
	"github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/postadjudicationreplay"
	"github.com/langoai/lango/internal/receipts"
)

func TestNetworkModuleMetadataAndEnablementBranches(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = false
	cfg.P2P.Enabled = false
	cfg.Economy.Enabled = false

	module := &networkModule{cfg: cfg}
	assert.Equal(t, "network", module.Name())
	assert.Equal(t, []appinit.Provides{
		appinit.ProvidesPayment,
		appinit.ProvidesP2P,
		appinit.ProvidesEconomy,
		appinit.ProvidesContract,
		appinit.ProvidesSmartAccount,
		appinit.ProvidesWorkspace,
	}, module.Provides())
	assert.Equal(t, []appinit.Provides{appinit.ProvidesSecurity, appinit.ProvidesSessionStore}, module.DependsOn())
	assert.False(t, module.Enabled())

	cfg.P2P.Enabled = true
	assert.True(t, module.Enabled())
	cfg.P2P.Enabled = false
	cfg.Economy.Enabled = true
	assert.True(t, module.Enabled())
}

func TestP2PPaymentToolEarlyValidationAvoidsChainRPC(t *testing.T) {
	t.Parallel()

	assert.Nil(t, buildP2PPaymentTool(&p2pComponents{}, nil, nil, nil))
	assert.Nil(t, buildP2PPaymentTool(&p2pComponents{}, &paymentComponents{}, nil, nil))

	validDID := p2PToolsMetadataAndMissingDependencyBranchesP2PPeerDID(t)
	sessions, err := handshake.NewSessionStore(time.Hour)
	require.NoError(t, err)
	_, err = sessions.Create("not-a-did", false)
	require.NoError(t, err)

	tool := findP2PTool(t, buildP2PPaymentTool(
		&p2pComponents{sessions: sessions},
		&paymentComponents{service: &payment.Service{}},
		receipts.NewStore(),
		&fakeP2PAuditor{},
	), "p2p_pay")

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"peer_did":               validDID.ID,
		"transaction_receipt_id": "tx-networkModuleMetadataAndEnablementBranches5-missing-session",
		"amount":                 "0.50",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "no active session for peer "+validDID.ID)

	got, err = tool.Handler(context.Background(), map[string]interface{}{
		"peer_did":               "not-a-did",
		"transaction_receipt_id": "tx-networkModuleMetadataAndEnablementBranches5-bad-did",
		"amount":                 "0.50",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "parse peer DID")
}

func TestMetaPostAdjudicationAdaptersHandleNilDependencies(t *testing.T) {
	t.Parallel()

	snapshots, err := (postAdjudicationStatusBackgroundTaskReader{}).ListTaskSnapshots(context.Background())
	require.NoError(t, err)
	assert.Nil(t, snapshots)

	receipt, err := (replayDispatcherAdapter{}).Dispatch(context.Background(), postadjudicationreplay.BackgroundDispatchRequest{
		TransactionReceiptID: "tx-networkModuleMetadataAndEnablementBranches5",
		SubmissionReceiptID:  "sub-networkModuleMetadataAndEnablementBranches5",
		EscrowReference:      "escrow-networkModuleMetadataAndEnablementBranches5",
		Outcome:              receipts.EscrowAdjudicationRelease,
		Prompt:               "retry settlement",
	})
	require.Error(t, err)
	assert.Equal(t, postadjudicationreplay.BackgroundDispatchReceipt{}, receipt)
	assert.ErrorContains(t, err, "background manager is not configured")
}

func TestP2PWiringPureAdapterBranchesAvoidNodeStartup(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.ZKHandshake = false
	cfg.P2P.ZKAttestation = false
	assert.Nil(t, initZKP(cfg))

	wallet := &wiringP2PWallet{}
	legacy := &legacyLocalIdentity{
		prov: nil,
		wp:   wallet,
	}
	assert.Nil(t, legacy.Bundle())
	assert.Equal(t, "secp256k1-keccak256", legacy.Algorithm())

	dispatcher := &networkModuleMetadataAndEnablementBranchesBackgroundDispatcher{taskID: "task-networkModuleMetadataAndEnablementBranches5"}
	dispatchReceipt, err := (replayDispatcherAdapter{dispatcher: dispatcher}).Dispatch(context.Background(), postadjudicationreplay.BackgroundDispatchRequest{
		TransactionReceiptID: "tx-networkModuleMetadataAndEnablementBranches5-dispatch",
		SubmissionReceiptID:  "sub-networkModuleMetadataAndEnablementBranches5-dispatch",
		EscrowReference:      "escrow-networkModuleMetadataAndEnablementBranches5-dispatch",
		Outcome:              receipts.EscrowAdjudicationRefund,
		Prompt:               "dispatch refund",
	})
	require.NoError(t, err)
	assert.Equal(t, "queued", dispatchReceipt.Status)
	assert.Equal(t, "task-networkModuleMetadataAndEnablementBranches5", dispatchReceipt.DispatchReference)
	assert.Equal(t, "dispatch refund", dispatcher.prompt)
}

func TestWiringHelpersResolveModeHintAndAutomationPrompt(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Modes = map[string]config.SessionMode{
		"networkModuleMetadataAndEnablementBranches5": {
			Name:       "networkModuleMetadataAndEnablementBranches5",
			SystemHint: "stay deterministic",
		},
	}

	resolver := &modeResolverAdapter{cfg: cfg}
	assert.Equal(t, "stay deterministic", resolver.LookupModeHint("networkModuleMetadataAndEnablementBranches5"))
	assert.Empty(t, resolver.LookupModeHint("missing"))
	assert.Empty(t, (*modeResolverAdapter)(nil).LookupModeHint("networkModuleMetadataAndEnablementBranches5"))

	cfg.Cron.Enabled = true
	cfg.Background.Enabled = true
	cfg.Workflow.Enabled = false
	rendered := buildAutomationPromptSection(cfg).Render()
	assert.Contains(t, rendered, "Cron Scheduling")
	assert.Contains(t, rendered, "Background Tasks")
	assert.NotContains(t, rendered, "Workflow Pipelines")
	assert.True(t, strings.Contains(rendered, "NEVER use exec"))
}

type networkModuleMetadataAndEnablementBranchesBackgroundDispatcher struct {
	taskID string
	prompt string
}

func (d *networkModuleMetadataAndEnablementBranchesBackgroundDispatcher) Submit(_ context.Context, prompt string, _ background.Origin) (string, error) {
	d.prompt = prompt
	return d.taskID, nil
}

func (d *networkModuleMetadataAndEnablementBranchesBackgroundDispatcher) List() []background.TaskSnapshot {
	return []background.TaskSnapshot{{ID: d.taskID, StatusText: "queued", RetryKey: "retry-networkModuleMetadataAndEnablementBranches5"}}
}
