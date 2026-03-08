package policy

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sa "github.com/langoai/lango/internal/smartaccount"
)

func TestValidator_Check(t *testing.T) {
	t.Parallel()

	addrA := common.HexToAddress("0xaaaa")
	addrB := common.HexToAddress("0xbbbb")
	now := time.Now()

	tests := []struct {
		give       string
		policy     *HarnessPolicy
		tracker    *SpendTracker
		call       *sa.ContractCall
		wantErr    error
		wantNoErr  bool
	}{
		{
			give: "allowed call passes all checks",
			policy: &HarnessPolicy{
				MaxTxAmount:      big.NewInt(1000),
				DailyLimit:       big.NewInt(5000),
				MonthlyLimit:     big.NewInt(50000),
				AllowedTargets:   []common.Address{addrA},
				AllowedFunctions: []string{"0x12345678"},
			},
			tracker: &SpendTracker{
				DailySpent:       big.NewInt(0),
				MonthlySpent:     big.NewInt(0),
				LastDailyReset:   now,
				LastMonthlyReset: now,
			},
			call: &sa.ContractCall{
				Target:      addrA,
				Value:       big.NewInt(500),
				FunctionSig: "0x12345678",
			},
			wantNoErr: true,
		},
		{
			give: "exceeds max transaction amount",
			policy: &HarnessPolicy{
				MaxTxAmount: big.NewInt(100),
			},
			tracker: nil,
			call: &sa.ContractCall{
				Target: addrA,
				Value:  big.NewInt(200),
			},
			wantErr: sa.ErrSpendLimitExceeded,
		},
		{
			give: "target not allowed",
			policy: &HarnessPolicy{
				AllowedTargets: []common.Address{addrA},
			},
			tracker: nil,
			call: &sa.ContractCall{
				Target: addrB,
				Value:  big.NewInt(0),
			},
			wantErr: sa.ErrTargetNotAllowed,
		},
		{
			give: "function not allowed",
			policy: &HarnessPolicy{
				AllowedFunctions: []string{"0x12345678"},
			},
			tracker: nil,
			call: &sa.ContractCall{
				Target:      addrA,
				Value:       big.NewInt(0),
				FunctionSig: "0xdeadbeef",
			},
			wantErr: sa.ErrFunctionNotAllowed,
		},
		{
			give: "exceeds daily spend limit",
			policy: &HarnessPolicy{
				DailyLimit: big.NewInt(1000),
			},
			tracker: &SpendTracker{
				DailySpent:       big.NewInt(800),
				MonthlySpent:     big.NewInt(800),
				LastDailyReset:   now,
				LastMonthlyReset: now,
			},
			call: &sa.ContractCall{
				Target: addrA,
				Value:  big.NewInt(300),
			},
			wantErr: sa.ErrSpendLimitExceeded,
		},
		{
			give: "exceeds monthly spend limit",
			policy: &HarnessPolicy{
				DailyLimit:   big.NewInt(10000),
				MonthlyLimit: big.NewInt(2000),
			},
			tracker: &SpendTracker{
				DailySpent:       big.NewInt(500),
				MonthlySpent:     big.NewInt(1800),
				LastDailyReset:   now,
				LastMonthlyReset: now,
			},
			call: &sa.ContractCall{
				Target: addrA,
				Value:  big.NewInt(300),
			},
			wantErr: sa.ErrSpendLimitExceeded,
		},
		{
			give: "empty function sig skips function check",
			policy: &HarnessPolicy{
				AllowedFunctions: []string{"0x12345678"},
			},
			tracker: nil,
			call: &sa.ContractCall{
				Target:      addrA,
				Value:       big.NewInt(0),
				FunctionSig: "",
			},
			wantNoErr: true,
		},
		{
			give: "empty targets allows any target",
			policy: &HarnessPolicy{
				AllowedTargets: nil,
			},
			tracker: nil,
			call: &sa.ContractCall{
				Target: addrB,
				Value:  big.NewInt(0),
			},
			wantNoErr: true,
		},
		{
			give:    "nil tracker skips spend checks",
			policy:  &HarnessPolicy{DailyLimit: big.NewInt(100)},
			tracker: nil,
			call: &sa.ContractCall{
				Target: addrA,
				Value:  big.NewInt(200),
			},
			wantNoErr: true,
		},
	}

	v := NewValidator()

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			err := v.Check(tt.policy, tt.tracker, tt.call)
			if tt.wantNoErr {
				require.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func TestValidator_Check_DailyResetAllowsSpend(t *testing.T) {
	t.Parallel()

	v := NewValidator()
	p := &HarnessPolicy{DailyLimit: big.NewInt(1000)}
	tracker := &SpendTracker{
		DailySpent:       big.NewInt(900),
		MonthlySpent:     big.NewInt(900),
		LastDailyReset:   time.Now().Add(-25 * time.Hour), // expired
		LastMonthlyReset: time.Now(),
	}
	call := &sa.ContractCall{Target: common.Address{}, Value: big.NewInt(500)}

	err := v.Check(p, tracker, call)
	require.NoError(t, err)
	// After reset, daily spent should be zero.
	assert.Equal(t, 0, tracker.DailySpent.Sign())
}
