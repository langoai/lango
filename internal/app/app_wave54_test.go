package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/runledger"
)

func TestWave54InitSecurityDeterministicDisabledAndErrorBranches(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Security.Signer.Provider = ""
	crypto, keys, secrets, err := initSecurity(cfg, &stubSessionStore{}, nil)
	require.NoError(t, err)
	assert.Nil(t, crypto)
	assert.Nil(t, keys)
	assert.Nil(t, secrets)

	cfg.Security.Signer.Provider = "rpc"
	crypto, keys, secrets, err = initSecurity(cfg, &stubSessionStore{}, nil)
	require.Error(t, err)
	assert.ErrorContains(t, err, "rpc security provider requires EntStore")
	assert.Nil(t, crypto)
	assert.Nil(t, keys)
	assert.Nil(t, secrets)

	cfg.Security.Signer.Provider = "wave54-unsupported"
	_, _, _, err = initSecurity(cfg, &stubSessionStore{}, nil)
	require.Error(t, err)
	assert.ErrorContains(t, err, `unsupported security provider "wave54-unsupported"`)
	assert.ErrorContains(t, err, "valid providers are local, rpc")
}

func TestWave54RetrievalWiringHelpersRespectConfigAndNilDependencies(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Retrieval.Enabled = false
	assert.Nil(t, initRetrievalCoordinator(cfg, nil))

	cfg.Retrieval.Enabled = true
	assert.NotNil(t, initRetrievalCoordinator(cfg, nil))

	bus := eventbus.New()
	assert.Zero(t, wave54EventHandlerCount(bus, eventbus.EventContextInjected))

	cfg.Retrieval.Feedback = false
	initFeedbackProcessor(cfg, bus)
	assert.Zero(t, wave54EventHandlerCount(bus, eventbus.EventContextInjected))

	cfg.Retrieval.Feedback = true
	initFeedbackProcessor(cfg, bus)
	assert.Equal(t, 1, wave54EventHandlerCount(bus, eventbus.EventContextInjected))

	cfg.Retrieval.AutoAdjust.Enabled = true
	initRelevanceAdjuster(cfg, nil, nil)
	initRelevanceAdjuster(cfg, nil, bus)
	assert.Equal(t, 1, wave54EventHandlerCount(bus, eventbus.EventContextInjected))
}

func TestWave54RunSummaryProviderAdapterDelegatesMaxJournalSeq(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := runledger.NewMemoryStore()
	adapter := &runSummaryProviderAdapter{store: store}

	require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "wave54-run-a",
		Type:    runledger.EventRunCreated,
		Payload: wave54JSON(t, runledger.RunCreatedPayload{SessionKey: "wave54-session", Goal: "first"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "wave54-run-a",
		Type:    runledger.EventNoteWritten,
		Payload: wave54JSON(t, runledger.NoteWrittenPayload{Key: "note", Value: "first"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "wave54-run-b",
		Type:    runledger.EventRunCreated,
		Payload: wave54JSON(t, runledger.RunCreatedPayload{SessionKey: "other-session", Goal: "other"}),
	}))

	maxSeq, err := adapter.MaxJournalSeqForSession(ctx, "wave54-session")
	require.NoError(t, err)
	assert.Equal(t, int64(2), maxSeq)

	missingSeq, err := adapter.MaxJournalSeqForSession(ctx, "missing-session")
	require.NoError(t, err)
	assert.Zero(t, missingSeq)
}

func TestWave54NetworkModuleInitAllNetworkSystemsDisabledAvoidsRuntimeServices(t *testing.T) {
	t.Parallel()

	cfg := wave34ModuleConfig(t)
	cfg.Payment.Enabled = false
	cfg.P2P.Enabled = false
	cfg.P2P.Workspace.Enabled = true
	cfg.Economy.Enabled = false
	cfg.SmartAccount.Enabled = true

	result, err := (&networkModule{cfg: cfg, bus: eventbus.New()}).Init(
		context.Background(),
		staticResolver{
			appinit.ProvidesSupervisor: &foundationValues{
				Store:        &stubSessionStore{},
				ReceiptStore: receipts.NewStore(),
			},
		},
	)
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
	assert.Contains(t, p2pEntry.Description, "P2P networking (disabled)")
	assert.NotContains(t, p2pEntry.Description, "payment required")
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "workspace").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "economy").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "smartaccount").Enabled)
}

func TestWave54InitP2PEnabledNodeCreationFailureAvoidsComponents(t *testing.T) {
	t.Parallel()

	key, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	keyDirFile := filepath.Join(t.TempDir(), "p2p-keydir-file")
	require.NoError(t, os.WriteFile(keyDirFile, []byte("not a directory"), 0o600))

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.KeyDir = keyDirFile
	cfg.P2P.ListenAddrs = []string{"/ip4/127.0.0.1/tcp/0"}
	wallet := &wiringP2PWallet{publicKey: ethcrypto.CompressPubkey(&key.PublicKey)}

	assert.Nil(t, initP2P(cfg, wallet, nil, nil, nil, nil, nil, nil, ""))
}

func TestWave54MetaToolsUseSettlementRuntimeAsPartialSettlementFallback(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "wave54-meta-partial-fallback")
	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionApprovedForSettlement, receipts.SettlementProgressionReasonCodeApprove, "approved", "settle:0.20-usdc")
	require.NoError(t, err)

	runtime := &fakeSettlementExecutionRuntime{}
	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, runtime, nil, nil, nil, nil), "execute_partial_settlement")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
	})
	require.NoError(t, err)

	payload, ok := got.(executePartialSettlementReceipt)
	require.True(t, ok)
	assert.Equal(t, string(receipts.SettlementProgressionPartiallySettled), payload.SettlementProgressionStatus)
	assert.Equal(t, "0.20", payload.ExecutedAmount)
	assert.Equal(t, "0.30", payload.RemainingAmount)
	assert.Equal(t, "settlement-tx-123", payload.RuntimeReference)
	assert.Equal(t, "0.20", runtime.last.Amount)
	assert.Equal(t, tx.TransactionReceiptID, runtime.last.TransactionReceiptID)
	assert.Equal(t, tx.CurrentSubmissionReceiptID, runtime.last.SubmissionReceiptID)
}

func wave54EventHandlerCount(bus *eventbus.Bus, eventName string) int {
	handlers := reflect.ValueOf(bus).Elem().FieldByName("handlers")
	value := handlers.MapIndex(reflect.ValueOf(eventName))
	if !value.IsValid() {
		return 0
	}
	return value.Len()
}

func wave54JSON(t *testing.T, v interface{}) json.RawMessage {
	t.Helper()

	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
}
