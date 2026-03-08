package app

import (
	"context"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/contract"
	"github.com/langoai/lango/internal/economy/budget"
	"github.com/langoai/lango/internal/economy/escrow/sentinel"
	"github.com/langoai/lango/internal/eventbus"
	sa "github.com/langoai/lango/internal/smartaccount"
	"github.com/langoai/lango/internal/smartaccount/bindings"
	"github.com/langoai/lango/internal/smartaccount/policy"
	sasession "github.com/langoai/lango/internal/smartaccount/session"
)

// ---------------------------------------------------------------------------
// WU-E3 Test 1: OnChainTracker budget callback sync
// ---------------------------------------------------------------------------

func TestBudgetTrackerSync(t *testing.T) {
	t.Parallel()

	tracker := budget.NewOnChainTracker()

	type callbackRecord struct {
		sessionID string
		spent     *big.Int
	}
	ch := make(chan callbackRecord, 10)

	tracker.SetCallback(func(sessionID string, spent *big.Int) {
		ch <- callbackRecord{sessionID: sessionID, spent: new(big.Int).Set(spent)}
	})

	// First spend.
	tracker.Record("session-A", big.NewInt(500))

	select {
	case rec := <-ch:
		assert.Equal(t, "session-A", rec.sessionID)
		assert.Equal(t, 0, rec.spent.Cmp(big.NewInt(500)),
			"first callback should report 500")
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for first callback")
	}

	// Second spend — cumulative.
	tracker.Record("session-A", big.NewInt(300))

	select {
	case rec := <-ch:
		assert.Equal(t, "session-A", rec.sessionID)
		assert.Equal(t, 0, rec.spent.Cmp(big.NewInt(800)),
			"second callback should report cumulative 800")
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for second callback")
	}

	// Verify GetSpent returns cumulative.
	assert.Equal(t, 0, tracker.GetSpent("session-A").Cmp(big.NewInt(800)))
}

func TestBudgetTrackerSync_MultipleSessions(t *testing.T) {
	t.Parallel()

	tracker := budget.NewOnChainTracker()

	var mu sync.Mutex
	calls := make(map[string]*big.Int)
	tracker.SetCallback(func(sessionID string, spent *big.Int) {
		mu.Lock()
		defer mu.Unlock()
		calls[sessionID] = new(big.Int).Set(spent)
	})

	tracker.Record("session-X", big.NewInt(100))
	tracker.Record("session-Y", big.NewInt(200))
	tracker.Record("session-X", big.NewInt(50))

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 0, calls["session-X"].Cmp(big.NewInt(150)))
	assert.Equal(t, 0, calls["session-Y"].Cmp(big.NewInt(200)))
}

// ---------------------------------------------------------------------------
// WU-E3 Test 2: SessionGuard revocation via sentinel alerts
// ---------------------------------------------------------------------------

func TestSessionGuardRevocation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a session manager with memory store.
	store := sasession.NewMemoryStore()
	mgr := sasession.NewManager(store, sasession.WithMaxKeys(10))

	// Create some session keys.
	now := time.Now()
	p := sa.SessionPolicy{
		AllowedTargets:   []common.Address{common.HexToAddress("0xaaaa")},
		AllowedFunctions: []string{"0x12345678"},
		SpendLimit:       big.NewInt(1000),
		ValidAfter:       now,
		ValidUntil:       now.Add(1 * time.Hour),
	}

	sk1, err := mgr.Create(ctx, p, "")
	require.NoError(t, err)
	sk2, err := mgr.Create(ctx, p, "")
	require.NoError(t, err)

	// Pre-check: both keys are active.
	active, err := store.ListActive(ctx)
	require.NoError(t, err)
	assert.Len(t, active, 2)

	// Create session guard wired to the manager (same pattern as wiring_smartaccount.go:201-204).
	bus := eventbus.New()
	guard := sentinel.NewSessionGuard(bus)
	guard.SetRevokeFunc(func() error {
		return mgr.RevokeAll(context.Background())
	})
	guard.Start()

	// Trigger a critical alert.
	bus.Publish(sentinel.SentinelAlertEvent{
		Alert: sentinel.Alert{
			Severity: sentinel.SeverityCritical,
			Type:     "anomalous_spend",
			Message:  "spending anomaly detected",
		},
	})

	// Verify all sessions are revoked.
	active, err = store.ListActive(ctx)
	require.NoError(t, err)
	assert.Empty(t, active, "all sessions should be revoked after critical alert")

	// Verify each key is individually marked as revoked.
	got1, err := mgr.Get(ctx, sk1.ID)
	require.NoError(t, err)
	assert.True(t, got1.Revoked)

	got2, err := mgr.Get(ctx, sk2.ID)
	require.NoError(t, err)
	assert.True(t, got2.Revoked)
}

func TestSessionGuardRevocation_HighSeverity(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := sasession.NewMemoryStore()
	mgr := sasession.NewManager(store)

	now := time.Now()
	p := sa.SessionPolicy{
		AllowedTargets: []common.Address{common.HexToAddress("0xbbbb")},
		SpendLimit:     big.NewInt(500),
		ValidAfter:     now,
		ValidUntil:     now.Add(1 * time.Hour),
	}
	_, err := mgr.Create(ctx, p, "")
	require.NoError(t, err)

	bus := eventbus.New()
	guard := sentinel.NewSessionGuard(bus)
	guard.SetRevokeFunc(func() error {
		return mgr.RevokeAll(context.Background())
	})
	guard.Start()

	// High severity should also trigger revocation.
	bus.Publish(sentinel.SentinelAlertEvent{
		Alert: sentinel.Alert{
			Severity: sentinel.SeverityHigh,
			Type:     "threat_detected",
			Message:  "high threat",
		},
	})

	active, err := store.ListActive(ctx)
	require.NoError(t, err)
	assert.Empty(t, active, "high severity should also revoke all sessions")
}

func TestSessionGuardRevocation_MediumNoRevoke(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := sasession.NewMemoryStore()
	mgr := sasession.NewManager(store)

	now := time.Now()
	p := sa.SessionPolicy{
		AllowedTargets: []common.Address{common.HexToAddress("0xcccc")},
		SpendLimit:     big.NewInt(500),
		ValidAfter:     now,
		ValidUntil:     now.Add(1 * time.Hour),
	}
	_, err := mgr.Create(ctx, p, "")
	require.NoError(t, err)

	bus := eventbus.New()
	guard := sentinel.NewSessionGuard(bus)

	revokedCalled := false
	guard.SetRevokeFunc(func() error {
		revokedCalled = true
		return mgr.RevokeAll(context.Background())
	})
	guard.SetRestrictFunc(func(factor float64) error {
		return nil
	})
	guard.Start()

	// Medium severity should NOT trigger revocation.
	bus.Publish(sentinel.SentinelAlertEvent{
		Alert: sentinel.Alert{
			Severity: sentinel.SeverityMedium,
			Type:     "suspicious_pattern",
			Message:  "medium threat",
		},
	})

	active, err := store.ListActive(ctx)
	require.NoError(t, err)
	assert.Len(t, active, 1, "medium alert should not revoke sessions")
	assert.False(t, revokedCalled, "revoke function should not be called on medium alert")
}

// ---------------------------------------------------------------------------
// WU-E3 Test 3: PolicySyncer drift detection
// ---------------------------------------------------------------------------

// mockContractCaller is a simple in-memory mock for contract.ContractCaller.
type mockContractCaller struct {
	mu       sync.Mutex
	readData map[string][]interface{} // method -> return data
}

// Compile-time check.
var _ contract.ContractCaller = (*mockContractCaller)(nil)

func newMockContractCaller() *mockContractCaller {
	return &mockContractCaller{
		readData: make(map[string][]interface{}),
	}
}

func (m *mockContractCaller) SetReadResponse(method string, data []interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readData[method] = data
}

func (m *mockContractCaller) Read(_ context.Context, req contract.ContractCallRequest) (*contract.ContractCallResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.readData[req.Method]
	if !ok {
		return &contract.ContractCallResult{Data: []interface{}{}}, nil
	}
	return &contract.ContractCallResult{Data: data}, nil
}

func (m *mockContractCaller) Write(_ context.Context, _ contract.ContractCallRequest) (*contract.ContractCallResult, error) {
	return &contract.ContractCallResult{TxHash: "0xmocktxhash"}, nil
}

func TestPolicySyncerDriftDetection_NoDrift(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	account := common.HexToAddress("0x1111")

	// Set up policy engine with a policy.
	engine := policy.New()
	engine.SetPolicy(account, &policy.HarnessPolicy{
		MaxTxAmount:  big.NewInt(1000),
		DailyLimit:   big.NewInt(5000),
		MonthlyLimit: big.NewInt(50000),
	})

	// Mock on-chain: same values.
	caller := newMockContractCaller()
	caller.SetReadResponse("getConfig", []interface{}{
		big.NewInt(1000),
		big.NewInt(5000),
		big.NewInt(50000),
	})

	hookAddr := common.HexToAddress("0x2222")
	hookClient := bindings.NewSpendingHookClient(caller, hookAddr, 1)

	syncer := policy.NewSyncer(engine, hookClient)

	report, err := syncer.DetectDrift(ctx, account)
	require.NoError(t, err)
	assert.False(t, report.HasDrift, "identical policies should not have drift")
	assert.Empty(t, report.Differences)
}

func TestPolicySyncerDriftDetection_DriftDetected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	account := common.HexToAddress("0x3333")

	// Go-side policy.
	engine := policy.New()
	engine.SetPolicy(account, &policy.HarnessPolicy{
		MaxTxAmount:  big.NewInt(1000),
		DailyLimit:   big.NewInt(5000),
		MonthlyLimit: big.NewInt(50000),
	})

	// On-chain: different values.
	caller := newMockContractCaller()
	caller.SetReadResponse("getConfig", []interface{}{
		big.NewInt(2000),  // differs from 1000
		big.NewInt(5000),  // same
		big.NewInt(30000), // differs from 50000
	})

	hookAddr := common.HexToAddress("0x4444")
	hookClient := bindings.NewSpendingHookClient(caller, hookAddr, 1)

	syncer := policy.NewSyncer(engine, hookClient)

	report, err := syncer.DetectDrift(ctx, account)
	require.NoError(t, err)
	assert.True(t, report.HasDrift, "differing limits should be detected as drift")
	assert.Len(t, report.Differences, 2, "should have 2 differences (perTx and cumulative)")
}

func TestPolicySyncerDriftDetection_NoPolicyError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	account := common.HexToAddress("0x5555")

	engine := policy.New()
	// No policy set for this account.

	caller := newMockContractCaller()
	hookClient := bindings.NewSpendingHookClient(caller, common.HexToAddress("0x6666"), 1)

	syncer := policy.NewSyncer(engine, hookClient)

	_, err := syncer.DetectDrift(ctx, account)
	require.Error(t, err, "should error when no Go-side policy exists")
	assert.Contains(t, err.Error(), "no Go-side policy")
}

func TestPolicySyncerDriftDetection_ZeroValues(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	account := common.HexToAddress("0x7777")

	// Go-side: nil limits (treated as zero).
	engine := policy.New()
	engine.SetPolicy(account, &policy.HarnessPolicy{})

	// On-chain: zero values.
	caller := newMockContractCaller()
	caller.SetReadResponse("getConfig", []interface{}{
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
	})

	hookClient := bindings.NewSpendingHookClient(caller, common.HexToAddress("0x8888"), 1)
	syncer := policy.NewSyncer(engine, hookClient)

	report, err := syncer.DetectDrift(ctx, account)
	require.NoError(t, err)
	assert.False(t, report.HasDrift, "nil and zero should be treated as equal (no drift)")
}

func TestPolicySyncerPullFromChain(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	account := common.HexToAddress("0x9999")

	engine := policy.New()
	engine.SetPolicy(account, &policy.HarnessPolicy{})

	caller := newMockContractCaller()
	caller.SetReadResponse("getConfig", []interface{}{
		big.NewInt(777),
		big.NewInt(8888),
		big.NewInt(99999),
	})

	hookClient := bindings.NewSpendingHookClient(caller, common.HexToAddress("0xAAAA"), 1)
	syncer := policy.NewSyncer(engine, hookClient)

	cfg, err := syncer.PullFromChain(ctx, account)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify on-chain values were pulled correctly.
	assert.Equal(t, 0, cfg.PerTxLimit.Cmp(big.NewInt(777)))
	assert.Equal(t, 0, cfg.DailyLimit.Cmp(big.NewInt(8888)))
	assert.Equal(t, 0, cfg.CumulativeLimit.Cmp(big.NewInt(99999)))

	// Verify Go-side policy was updated.
	goPolicy, ok := engine.GetPolicy(account)
	require.True(t, ok)
	assert.Equal(t, 0, goPolicy.MaxTxAmount.Cmp(big.NewInt(777)))
	assert.Equal(t, 0, goPolicy.DailyLimit.Cmp(big.NewInt(8888)))
	assert.Equal(t, 0, goPolicy.MonthlyLimit.Cmp(big.NewInt(99999)))
}
