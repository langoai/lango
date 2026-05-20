package storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/langoai/lango/internal/agentmemory"
	"github.com/langoai/lango/internal/cron"
	langoent "github.com/langoai/lango/internal/ent"
	entauditlog "github.com/langoai/lango/internal/ent/auditlog"
	"github.com/langoai/lango/internal/ent/enttest"
	entinquiry "github.com/langoai/lango/internal/ent/inquiry"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	entpaymenttx "github.com/langoai/lango/internal/ent/paymenttx"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/turntrace"
	"github.com/langoai/lango/internal/types"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestFacadeUnavailableCapabilitiesReturnErrorsAndNilValues(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	f := NewFacade(nil, nil)

	_, err := f.OpenSessionStore()
	require.ErrorContains(t, err, "session storage unavailable")
	require.Nil(t, f.KeyRegistry())
	require.Nil(t, f.SecretsStore(nil))
	require.Nil(t, f.RunLedger())
	require.Nil(t, f.Cron())
	require.Nil(t, f.TurnTrace())
	require.Nil(t, f.AgentMemory())
	require.Nil(t, f.Provenance())
	require.Nil(t, f.AuditRecorder())
	require.Nil(t, f.TokenStore())
	require.Nil(t, f.OntologyDeps())
	require.Nil(t, f.ReputationStore(nil))
	require.Nil(t, f.WorkflowStateStore(nil))
	require.Nil(t, f.PaymentTxStore())
	require.NoError(t, f.Close())

	_, err = f.SecuritySummary(ctx)
	require.ErrorContains(t, err, "security diagnostics unavailable")
	_, err = f.RecentSandboxDecisions(ctx, "", 1)
	require.ErrorContains(t, err, "audit storage unavailable")
	_, err = f.LearningHistory(ctx, 1)
	require.ErrorContains(t, err, "learning storage unavailable")
	_, err = f.PendingInquiries(ctx, 1)
	require.ErrorContains(t, err, "inquiry storage unavailable")
	_, err = f.Alerts(ctx, time.Time{})
	require.ErrorContains(t, err, "alert storage unavailable")
	_, err = f.ReputationDetails(ctx, "did:lango:test")
	require.ErrorContains(t, err, "reputation storage unavailable")
	_, err = f.PaymentHistory(ctx, 1)
	require.ErrorContains(t, err, "payment history unavailable")
	_, err = f.PaymentUsage(ctx)
	require.ErrorContains(t, err, "payment usage unavailable")
	_, err = f.NewSpendingLimiter("1", "2", "0")
	require.ErrorContains(t, err, "payment limiter unavailable")
}

func TestFacadeSessionDBPathSupportsSessionCRUD(t *testing.T) {
	t.Parallel()

	f := NewFacade(nil, nil, WithSessionDBPath(t.TempDir()+"/sessions.db"))
	store, err := f.OpenSessionStore()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	sess := &session.Session{
		Key:   "session-1",
		Model: "gpt-test",
		History: []session.Message{{
			Role:      types.RoleUser,
			Content:   "hello",
			Timestamp: time.Unix(10, 0),
		}},
	}
	require.NoError(t, store.Create(sess))

	got, err := store.Get("session-1")
	require.NoError(t, err)
	require.Equal(t, "gpt-test", got.Model)
	require.Len(t, got.History, 1)
	require.Equal(t, "hello", got.History[0].Content)

	require.NoError(t, store.AppendMessage("session-1", session.Message{
		Role:      types.RoleAssistant,
		Content:   "world",
		Timestamp: time.Unix(20, 0),
	}))
	got, err = store.Get("session-1")
	require.NoError(t, err)
	require.Len(t, got.History, 2)
	require.Equal(t, "world", got.History[1].Content)

	require.NoError(t, store.Delete("session-1"))
	_, err = store.Get("session-1")
	require.Error(t, err)
}

func TestFacadeOptionsDelegateErrorsAndArguments(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	wantErr := errors.New("forced failure")
	f := NewFacade(nil, nil,
		WithSessionStoreFactory(func(opts ...session.StoreOption) (session.Store, error) {
			require.Len(t, opts, 1)
			return nil, wantErr
		}),
		WithSecurityDiagnostics(func(context.Context) (SecuritySummary, error) {
			return SecuritySummary{EncryptionKeys: 2, StoredSecrets: 3}, nil
		}),
		WithSandboxDecisionReader(func(_ context.Context, prefix string, limit int) ([]SandboxDecisionRecord, error) {
			require.Equal(t, "abc", prefix)
			require.Equal(t, 7, limit)
			return []SandboxDecisionRecord{{SessionKey: "abc-1", Decision: "allow"}}, nil
		}),
	)

	_, err := f.OpenSessionStore(func(*session.EntStore) {})
	require.ErrorIs(t, err, wantErr)

	summary, err := f.SecuritySummary(ctx)
	require.NoError(t, err)
	require.Equal(t, SecuritySummary{EncryptionKeys: 2, StoredSecrets: 3}, summary)

	decisions, err := f.RecentSandboxDecisions(ctx, "abc", 7)
	require.NoError(t, err)
	require.Equal(t, []SandboxDecisionRecord{{SessionKey: "abc-1", Decision: "allow"}}, decisions)
}

type facadeTestCrypto struct{}

func (facadeTestCrypto) Sign(context.Context, string, []byte) ([]byte, error) {
	return nil, nil
}

func (facadeTestCrypto) Encrypt(context.Context, string, []byte) ([]byte, error) {
	return nil, nil
}

func (facadeTestCrypto) Decrypt(context.Context, string, []byte) ([]byte, error) {
	return nil, nil
}

func TestFacadeCapabilityFactoriesAreDelegated(t *testing.T) {
	t.Parallel()

	calls := map[string]int{}
	provenanceStores := &ProvenanceStores{}
	wantKeyRegistry := &security.KeyRegistry{}
	wantSecretsStore := &security.SecretsStore{}
	wantRunLedger := runledger.RunLedgerStore(&struct{ runledger.RunLedgerStore }{})
	wantCron := cron.Store(&struct{ cron.Store }{})
	wantTurnTrace := turntrace.Store(&struct{ turntrace.Store }{})
	wantAgentMemory := agentmemory.Store(&struct{ agentmemory.Store }{})
	f := NewFacade(nil, nil,
		WithKeyRegistryFactory(func() *security.KeyRegistry {
			calls["keyRegistry"]++
			return wantKeyRegistry
		}),
		WithSecretsStoreFactory(func(crypto security.CryptoProvider) *security.SecretsStore {
			calls["secretsStore"]++
			require.NotNil(t, crypto)
			return wantSecretsStore
		}),
		WithRunLedgerFactory(func() runledger.RunLedgerStore {
			calls["runLedger"]++
			return wantRunLedger
		}),
		WithCronFactory(func() cron.Store {
			calls["cron"]++
			return wantCron
		}),
		WithTurnTraceFactory(func() turntrace.Store {
			calls["turnTrace"]++
			return wantTurnTrace
		}),
		WithAgentMemoryFactory(func() agentmemory.Store {
			calls["agentMemory"]++
			return wantAgentMemory
		}),
		WithProvenanceStores(provenanceStores),
	)

	require.Same(t, wantKeyRegistry, f.KeyRegistry())
	require.Same(t, wantSecretsStore, f.SecretsStore(facadeTestCrypto{}))
	require.Same(t, wantRunLedger, f.RunLedger())
	require.Same(t, wantCron, f.Cron())
	require.Same(t, wantTurnTrace, f.TurnTrace())
	require.Same(t, wantAgentMemory, f.AgentMemory())
	require.Same(t, provenanceStores, f.Provenance())
	require.Equal(t, map[string]int{
		"keyRegistry":  1,
		"secretsStore": 1,
		"runLedger":    1,
		"cron":         1,
		"turnTrace":    1,
		"agentMemory":  1,
	}, calls)
}

func TestProvenanceStoresZeroValueGettersReturnNilStores(t *testing.T) {
	t.Parallel()

	stores := &ProvenanceStores{}

	require.Nil(t, stores.Checkpoints())
	require.Nil(t, stores.SessionTree())
	require.Nil(t, stores.Attribution())
	require.Nil(t, stores.TokenUsage())
}

func TestFacadeWithEntClientExposesDerivedCapabilities(t *testing.T) {
	t.Parallel()

	client := enttest.Open(t, "sqlite3", "file:storage-facade-derived-capabilities?mode=memory&_fk=1")
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	f := NewFacade(nil, nil, WithEntClient(client))

	require.NotNil(t, f.KeyRegistry())
	require.NotNil(t, f.SecretsStore(facadeTestCrypto{}))
	require.NotNil(t, f.RunLedger())
	require.NotNil(t, f.Cron())
	require.NotNil(t, f.TurnTrace())
	require.NotNil(t, f.AgentMemory())
	require.NotNil(t, f.Provenance())
	require.NotNil(t, f.Provenance().Checkpoints())
	require.NotNil(t, f.Provenance().SessionTree())
	require.NotNil(t, f.Provenance().Attribution())
	require.NotNil(t, f.Provenance().TokenUsage())
	require.NotNil(t, f.AuditRecorder())
	require.NotNil(t, f.TokenStore())
	require.NotNil(t, f.OntologyDeps())
	require.NotNil(t, f.ReputationStore(nil))
	require.NotNil(t, f.WorkflowStateStore(nil))
	require.NotNil(t, f.PaymentTxStore())

	limiter, err := f.NewSpendingLimiter("10", "20", "1")
	require.NoError(t, err)
	require.NotNil(t, limiter)
}

func TestWithEntClientReadersFilterSortAndDefaultLimits(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := enttest.Open(t, "sqlite3", "file:storage-facade?mode=memory&_fk=1")
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	oldTime := storageStartOfToday().Add(time.Hour).Truncate(time.Second)
	newTime := oldTime.Add(time.Hour)
	require.NoError(t, client.Learning.Create().
		SetTrigger("old").
		SetCategory(entlearning.CategoryTimeout).
		SetDiagnosis("old diagnosis").
		SetFix("old fix").
		SetConfidence(0.4).
		SetCreatedAt(oldTime).
		Exec(ctx))
	require.NoError(t, client.Learning.Create().
		SetTrigger("new").
		SetCategory(entlearning.CategoryGeneral).
		SetDiagnosis("new diagnosis").
		SetFix("new fix").
		SetConfidence(0.9).
		SetCreatedAt(newTime).
		Exec(ctx))

	require.NoError(t, client.Inquiry.Create().
		SetSessionKey("s1").
		SetTopic("pending").
		SetQuestion("what next?").
		SetPriority(entinquiry.PriorityHigh).
		SetStatus(entinquiry.StatusPending).
		SetCreatedAt(oldTime).
		Exec(ctx))
	require.NoError(t, client.Inquiry.Create().
		SetSessionKey("s1").
		SetTopic("resolved").
		SetQuestion("done?").
		SetStatus(entinquiry.StatusResolved).
		Exec(ctx))

	require.NoError(t, client.AuditLog.Create().
		SetAction(entauditlog.ActionAlert).
		SetActor("system").
		SetTarget("policy").
		SetDetails(map[string]interface{}{"severity": "high"}).
		SetTimestamp(newTime).
		Exec(ctx))
	require.NoError(t, client.AuditLog.Create().
		SetAction(entauditlog.ActionSandboxDecision).
		SetActor("sandbox").
		SetSessionKey("abc-123").
		SetTarget("shell").
		SetDetails(map[string]interface{}{
			"decision": "deny",
			"backend":  "local",
			"reason":   "policy",
		}).
		SetTimestamp(newTime).
		Exec(ctx))

	require.NoError(t, client.PaymentTx.Create().
		SetTxHash("tx-new").
		SetFromAddress("alice").
		SetToAddress("bob").
		SetAmount("1.25").
		SetChainID(1).
		SetStatus(entpaymenttx.StatusConfirmed).
		SetPaymentMethod(entpaymenttx.PaymentMethodDirectTransfer).
		SetCreatedAt(newTime).
		Exec(ctx))
	require.NoError(t, client.PaymentTx.Create().
		SetTxHash("tx-failed").
		SetFromAddress("alice").
		SetToAddress("bob").
		SetAmount("9.00").
		SetChainID(1).
		SetStatus(entpaymenttx.StatusFailed).
		SetCreatedAt(newTime.Add(time.Second)).
		Exec(ctx))

	f := NewFacade(nil, nil, WithEntClient(client))

	learning, err := f.LearningHistory(ctx, 1)
	require.NoError(t, err)
	require.Len(t, learning, 1)
	require.Equal(t, "new", learning[0].Trigger)
	require.Equal(t, "general", learning[0].Category)

	inquiries, err := f.PendingInquiries(ctx, 0)
	require.NoError(t, err)
	require.Len(t, inquiries, 1)
	require.Equal(t, "pending", inquiries[0].Topic)
	require.Equal(t, "high", inquiries[0].Priority)

	alerts, err := f.Alerts(ctx, oldTime)
	require.NoError(t, err)
	require.Len(t, alerts, 1)
	require.Equal(t, "policy", alerts[0].Type)
	require.Equal(t, "high", alerts[0].Details["severity"])

	decisions, err := f.RecentSandboxDecisions(ctx, "abc", 0)
	require.NoError(t, err)
	require.Len(t, decisions, 1)
	require.True(t, decisions[0].Timestamp.Equal(newTime))
	require.Equal(t, "abc-123", decisions[0].SessionKey)
	require.Equal(t, "shell", decisions[0].Target)
	require.Equal(t, "deny", decisions[0].Decision)
	require.Equal(t, "local", decisions[0].Backend)
	require.Equal(t, "policy", decisions[0].Reason)

	history, err := f.PaymentHistory(ctx, 1)
	require.NoError(t, err)
	require.Len(t, history, 1)
	require.Equal(t, "tx-failed", history[0].TxHash)

	usage, err := f.PaymentUsage(ctx)
	require.NoError(t, err)
	require.Equal(t, "1.25", usage.DailySpent)

	resolved, ok := ResolveEntBacked(f, func(*langoent.Client) string {
		return "ent-backed"
	})
	require.True(t, ok)
	require.Equal(t, "ent-backed", resolved)

	missing, ok := ResolveEntBacked[string](nil, nil)
	require.False(t, ok)
	require.Empty(t, missing)
}

func TestWithRawDBUsesDatabaseCloseWhenNoCloserExists(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	f := NewFacade(nil, nil, WithRawDB(db))
	require.NoError(t, f.Close())
	require.Error(t, db.Ping())
}
