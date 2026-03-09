package policy

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/contract"
	"github.com/langoai/lango/internal/smartaccount/bindings"
)

// mockContractCaller stubs out the ContractCaller interface for testing.
type mockContractCaller struct {
	readFn  func(ctx context.Context, req contract.ContractCallRequest) (*contract.ContractCallResult, error)
	writeFn func(ctx context.Context, req contract.ContractCallRequest) (*contract.ContractCallResult, error)
}

func (m *mockContractCaller) Read(ctx context.Context, req contract.ContractCallRequest) (*contract.ContractCallResult, error) {
	if m.readFn != nil {
		return m.readFn(ctx, req)
	}
	return &contract.ContractCallResult{}, nil
}

func (m *mockContractCaller) Write(ctx context.Context, req contract.ContractCallRequest) (*contract.ContractCallResult, error) {
	if m.writeFn != nil {
		return m.writeFn(ctx, req)
	}
	return &contract.ContractCallResult{TxHash: "0xmocktx"}, nil
}

// newTestSyncer creates a Syncer wired with mocked dependencies.
func newTestSyncer(caller *mockContractCaller) (*Syncer, *Engine) {
	engine := New()
	hookAddr := common.HexToAddress("0xHook")
	hook := bindings.NewSpendingHookClient(caller, hookAddr, 1)
	syncer := NewSyncer(engine, hook)
	return syncer, engine
}

func TestPushToChain(t *testing.T) {
	t.Parallel()

	account := common.HexToAddress("0xABCD")

	tests := []struct {
		give       string
		policy     *HarnessPolicy
		writeFn    func(ctx context.Context, req contract.ContractCallRequest) (*contract.ContractCallResult, error)
		wantTxHash string
		wantErr    string
	}{
		{
			give: "all limits set",
			policy: &HarnessPolicy{
				MaxTxAmount:  big.NewInt(1000),
				DailyLimit:   big.NewInt(5000),
				MonthlyLimit: big.NewInt(50000),
			},
			writeFn: func(_ context.Context, req contract.ContractCallRequest) (*contract.ContractCallResult, error) {
				return &contract.ContractCallResult{TxHash: "0xaaa"}, nil
			},
			wantTxHash: "0xaaa",
		},
		{
			give: "nil limits default to zero",
			policy: &HarnessPolicy{
				MaxTxAmount:  nil,
				DailyLimit:   nil,
				MonthlyLimit: nil,
			},
			writeFn: func(_ context.Context, req contract.ContractCallRequest) (*contract.ContractCallResult, error) {
				args := req.Args
				require.Len(t, args, 3)
				for i, arg := range args {
					v, ok := arg.(*big.Int)
					require.True(t, ok, "arg[%d] should be *big.Int", i)
					assert.Equal(t, 0, v.Sign(), "nil limit should become zero")
				}
				return &contract.ContractCallResult{TxHash: "0xbbb"}, nil
			},
			wantTxHash: "0xbbb",
		},
		{
			give: "partial nil limits",
			policy: &HarnessPolicy{
				MaxTxAmount:  big.NewInt(100),
				DailyLimit:   nil,
				MonthlyLimit: big.NewInt(9999),
			},
			writeFn: func(_ context.Context, req contract.ContractCallRequest) (*contract.ContractCallResult, error) {
				args := req.Args
				require.Len(t, args, 3)
				perTx := args[0].(*big.Int)
				daily := args[1].(*big.Int)
				cumul := args[2].(*big.Int)
				assert.Equal(t, int64(100), perTx.Int64())
				assert.Equal(t, int64(0), daily.Int64())
				assert.Equal(t, int64(9999), cumul.Int64())
				return &contract.ContractCallResult{TxHash: "0xccc"}, nil
			},
			wantTxHash: "0xccc",
		},
		{
			give: "write error propagated",
			policy: &HarnessPolicy{
				MaxTxAmount: big.NewInt(1),
			},
			writeFn: func(_ context.Context, _ contract.ContractCallRequest) (*contract.ContractCallResult, error) {
				return nil, errors.New("rpc down")
			},
			wantErr: "set limits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			caller := &mockContractCaller{writeFn: tt.writeFn}
			syncer, engine := newTestSyncer(caller)
			engine.SetPolicy(account, tt.policy)

			txHash, err := syncer.PushToChain(context.Background(), account)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantTxHash, txHash)
		})
	}
}

func TestPushToChain_NoPolicy(t *testing.T) {
	t.Parallel()

	caller := &mockContractCaller{}
	syncer, _ := newTestSyncer(caller)
	missingAccount := common.HexToAddress("0xDEAD")

	_, err := syncer.PushToChain(context.Background(), missingAccount)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no policy for account")
}

func TestPullFromChain(t *testing.T) {
	t.Parallel()

	account := common.HexToAddress("0xABCD")

	tests := []struct {
		give        string
		prePolicy   *HarnessPolicy
		onChainCfg  *bindings.SpendingConfig
		wantPerTx   *big.Int
		wantDaily   *big.Int
		wantMonthly *big.Int
	}{
		{
			give:      "updates existing policy with on-chain values",
			prePolicy: &HarnessPolicy{MaxTxAmount: big.NewInt(100)},
			onChainCfg: &bindings.SpendingConfig{
				PerTxLimit:      big.NewInt(500),
				DailyLimit:      big.NewInt(2000),
				CumulativeLimit: big.NewInt(20000),
			},
			wantPerTx:   big.NewInt(500),
			wantDaily:   big.NewInt(2000),
			wantMonthly: big.NewInt(20000),
		},
		{
			give:      "creates policy when none exists",
			prePolicy: nil,
			onChainCfg: &bindings.SpendingConfig{
				PerTxLimit:      big.NewInt(300),
				DailyLimit:      big.NewInt(1500),
				CumulativeLimit: big.NewInt(15000),
			},
			wantPerTx:   big.NewInt(300),
			wantDaily:   big.NewInt(1500),
			wantMonthly: big.NewInt(15000),
		},
		{
			give:      "zero on-chain values do not override existing",
			prePolicy: &HarnessPolicy{MaxTxAmount: big.NewInt(100), DailyLimit: big.NewInt(999)},
			onChainCfg: &bindings.SpendingConfig{
				PerTxLimit:      big.NewInt(0),
				DailyLimit:      big.NewInt(0),
				CumulativeLimit: big.NewInt(0),
			},
			wantPerTx:   big.NewInt(100),
			wantDaily:   big.NewInt(999),
			wantMonthly: nil,
		},
		{
			give:      "nil on-chain values do not override existing",
			prePolicy: &HarnessPolicy{MaxTxAmount: big.NewInt(42)},
			onChainCfg: &bindings.SpendingConfig{
				PerTxLimit:      nil,
				DailyLimit:      nil,
				CumulativeLimit: nil,
			},
			wantPerTx:   big.NewInt(42),
			wantDaily:   nil,
			wantMonthly: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			cfg := tt.onChainCfg
			caller := &mockContractCaller{
				readFn: func(_ context.Context, _ contract.ContractCallRequest) (*contract.ContractCallResult, error) {
					return &contract.ContractCallResult{
						Data: []interface{}{cfg.PerTxLimit, cfg.DailyLimit, cfg.CumulativeLimit},
					}, nil
				},
			}
			syncer, engine := newTestSyncer(caller)
			if tt.prePolicy != nil {
				engine.SetPolicy(account, tt.prePolicy)
			}

			gotCfg, err := syncer.PullFromChain(context.Background(), account)
			require.NoError(t, err)
			require.NotNil(t, gotCfg)

			// Verify returned config matches on-chain.
			assert.Equal(t, cfg.PerTxLimit, gotCfg.PerTxLimit)
			assert.Equal(t, cfg.DailyLimit, gotCfg.DailyLimit)
			assert.Equal(t, cfg.CumulativeLimit, gotCfg.CumulativeLimit)

			// Verify Go-side policy was updated.
			policy, ok := engine.GetPolicy(account)
			require.True(t, ok)

			assertBigIntEqual(t, tt.wantPerTx, policy.MaxTxAmount, "MaxTxAmount")
			assertBigIntEqual(t, tt.wantDaily, policy.DailyLimit, "DailyLimit")
			assertBigIntEqual(t, tt.wantMonthly, policy.MonthlyLimit, "MonthlyLimit")
		})
	}
}

func TestPullFromChain_ReadError(t *testing.T) {
	t.Parallel()

	caller := &mockContractCaller{
		readFn: func(_ context.Context, _ contract.ContractCallRequest) (*contract.ContractCallResult, error) {
			return nil, errors.New("network timeout")
		},
	}
	syncer, _ := newTestSyncer(caller)

	_, err := syncer.PullFromChain(context.Background(), common.HexToAddress("0x1"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "get on-chain config")
}

func TestDetectDrift(t *testing.T) {
	t.Parallel()

	account := common.HexToAddress("0xABCD")

	tests := []struct {
		give           string
		goPolicy       *HarnessPolicy
		onChainCfg     *bindings.SpendingConfig
		wantDrift      bool
		wantDiffCount  int
		wantDiffSubstr []string
	}{
		{
			give: "no drift when values match",
			goPolicy: &HarnessPolicy{
				MaxTxAmount:  big.NewInt(1000),
				DailyLimit:   big.NewInt(5000),
				MonthlyLimit: big.NewInt(50000),
			},
			onChainCfg: &bindings.SpendingConfig{
				PerTxLimit:      big.NewInt(1000),
				DailyLimit:      big.NewInt(5000),
				CumulativeLimit: big.NewInt(50000),
			},
			wantDrift:     false,
			wantDiffCount: 0,
		},
		{
			give: "no drift when both nil",
			goPolicy: &HarnessPolicy{
				MaxTxAmount:  nil,
				DailyLimit:   nil,
				MonthlyLimit: nil,
			},
			onChainCfg: &bindings.SpendingConfig{
				PerTxLimit:      nil,
				DailyLimit:      nil,
				CumulativeLimit: nil,
			},
			wantDrift:     false,
			wantDiffCount: 0,
		},
		{
			give: "no drift nil vs zero",
			goPolicy: &HarnessPolicy{
				MaxTxAmount: nil,
			},
			onChainCfg: &bindings.SpendingConfig{
				PerTxLimit:      big.NewInt(0),
				DailyLimit:      big.NewInt(0),
				CumulativeLimit: big.NewInt(0),
			},
			wantDrift:     false,
			wantDiffCount: 0,
		},
		{
			give: "drift on perTxLimit mismatch",
			goPolicy: &HarnessPolicy{
				MaxTxAmount:  big.NewInt(100),
				DailyLimit:   big.NewInt(500),
				MonthlyLimit: big.NewInt(5000),
			},
			onChainCfg: &bindings.SpendingConfig{
				PerTxLimit:      big.NewInt(200),
				DailyLimit:      big.NewInt(500),
				CumulativeLimit: big.NewInt(5000),
			},
			wantDrift:      true,
			wantDiffCount:  1,
			wantDiffSubstr: []string{"perTxLimit"},
		},
		{
			give: "drift on dailyLimit mismatch",
			goPolicy: &HarnessPolicy{
				MaxTxAmount:  big.NewInt(100),
				DailyLimit:   big.NewInt(500),
				MonthlyLimit: big.NewInt(5000),
			},
			onChainCfg: &bindings.SpendingConfig{
				PerTxLimit:      big.NewInt(100),
				DailyLimit:      big.NewInt(999),
				CumulativeLimit: big.NewInt(5000),
			},
			wantDrift:      true,
			wantDiffCount:  1,
			wantDiffSubstr: []string{"dailyLimit"},
		},
		{
			give: "drift on cumulativeLimit mismatch",
			goPolicy: &HarnessPolicy{
				MaxTxAmount:  big.NewInt(100),
				DailyLimit:   big.NewInt(500),
				MonthlyLimit: big.NewInt(5000),
			},
			onChainCfg: &bindings.SpendingConfig{
				PerTxLimit:      big.NewInt(100),
				DailyLimit:      big.NewInt(500),
				CumulativeLimit: big.NewInt(9999),
			},
			wantDrift:      true,
			wantDiffCount:  1,
			wantDiffSubstr: []string{"cumulativeLimit"},
		},
		{
			give: "all three fields differ",
			goPolicy: &HarnessPolicy{
				MaxTxAmount:  big.NewInt(1),
				DailyLimit:   big.NewInt(2),
				MonthlyLimit: big.NewInt(3),
			},
			onChainCfg: &bindings.SpendingConfig{
				PerTxLimit:      big.NewInt(10),
				DailyLimit:      big.NewInt(20),
				CumulativeLimit: big.NewInt(30),
			},
			wantDrift:      true,
			wantDiffCount:  3,
			wantDiffSubstr: []string{"perTxLimit", "dailyLimit", "cumulativeLimit"},
		},
		{
			give: "drift when go-side nil but on-chain non-zero",
			goPolicy: &HarnessPolicy{
				MaxTxAmount: nil,
			},
			onChainCfg: &bindings.SpendingConfig{
				PerTxLimit:      big.NewInt(999),
				DailyLimit:      nil,
				CumulativeLimit: nil,
			},
			wantDrift:      true,
			wantDiffCount:  1,
			wantDiffSubstr: []string{"perTxLimit"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			cfg := tt.onChainCfg
			caller := &mockContractCaller{
				readFn: func(_ context.Context, _ contract.ContractCallRequest) (*contract.ContractCallResult, error) {
					return &contract.ContractCallResult{
						Data: []interface{}{cfg.PerTxLimit, cfg.DailyLimit, cfg.CumulativeLimit},
					}, nil
				},
			}
			syncer, engine := newTestSyncer(caller)
			engine.SetPolicy(account, tt.goPolicy)

			report, err := syncer.DetectDrift(context.Background(), account)
			require.NoError(t, err)
			require.NotNil(t, report)
			assert.Equal(t, account, report.Account)
			assert.Equal(t, tt.wantDrift, report.HasDrift)
			assert.Len(t, report.Differences, tt.wantDiffCount)

			for _, substr := range tt.wantDiffSubstr {
				found := false
				for _, diff := range report.Differences {
					if assert.ObjectsAreEqual(true, contains(diff, substr)) {
						found = true
						break
					}
				}
				assert.True(t, found, "expected difference containing %q", substr)
			}
		})
	}
}

func TestDetectDrift_NoGoPolicy(t *testing.T) {
	t.Parallel()

	caller := &mockContractCaller{}
	syncer, _ := newTestSyncer(caller)
	missingAccount := common.HexToAddress("0xDEAD")

	_, err := syncer.DetectDrift(context.Background(), missingAccount)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no Go-side policy")
}

func TestDetectDrift_OnChainError(t *testing.T) {
	t.Parallel()

	account := common.HexToAddress("0xABCD")
	caller := &mockContractCaller{
		readFn: func(_ context.Context, _ contract.ContractCallRequest) (*contract.ContractCallResult, error) {
			return nil, errors.New("contract reverted")
		},
	}
	syncer, engine := newTestSyncer(caller)
	engine.SetPolicy(account, &HarnessPolicy{MaxTxAmount: big.NewInt(100)})

	_, err := syncer.DetectDrift(context.Background(), account)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "get on-chain config")
}

func TestBigIntEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		a    *big.Int
		b    *big.Int
		want bool
	}{
		{
			give: "both nil",
			a:    nil,
			b:    nil,
			want: true,
		},
		{
			give: "a nil b zero",
			a:    nil,
			b:    big.NewInt(0),
			want: true,
		},
		{
			give: "a zero b nil",
			a:    big.NewInt(0),
			b:    nil,
			want: true,
		},
		{
			give: "both zero",
			a:    big.NewInt(0),
			b:    big.NewInt(0),
			want: true,
		},
		{
			give: "equal positive",
			a:    big.NewInt(42),
			b:    big.NewInt(42),
			want: true,
		},
		{
			give: "equal negative",
			a:    big.NewInt(-7),
			b:    big.NewInt(-7),
			want: true,
		},
		{
			give: "different values",
			a:    big.NewInt(100),
			b:    big.NewInt(200),
			want: false,
		},
		{
			give: "a nil b non-zero",
			a:    nil,
			b:    big.NewInt(999),
			want: false,
		},
		{
			give: "a non-zero b nil",
			a:    big.NewInt(999),
			b:    nil,
			want: false,
		},
		{
			give: "large equal values",
			a:    new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
			b:    new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
			want: true,
		},
		{
			give: "large different values",
			a:    new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
			b:    new(big.Int).Exp(big.NewInt(10), big.NewInt(19), nil),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := bigIntEqual(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewSyncer(t *testing.T) {
	t.Parallel()

	engine := New()
	hookAddr := common.HexToAddress("0xHook")
	caller := &mockContractCaller{}
	hook := bindings.NewSpendingHookClient(caller, hookAddr, 1)

	syncer := NewSyncer(engine, hook)

	require.NotNil(t, syncer)
}

// assertBigIntEqual is a test helper that compares two *big.Int with a label.
func assertBigIntEqual(t *testing.T, want, got *big.Int, label string) {
	t.Helper()
	if want == nil && got == nil {
		return
	}
	if want == nil {
		assert.Nil(t, got, "%s: expected nil", label)
		return
	}
	require.NotNil(t, got, "%s: expected non-nil", label)
	assert.Equal(t, 0, want.Cmp(got), "%s: want=%v got=%v", label, want, got)
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
