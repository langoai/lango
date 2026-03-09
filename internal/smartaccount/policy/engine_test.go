package policy

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sa "github.com/langoai/lango/internal/smartaccount"
)

func TestEngine_SetPolicy_GetPolicy(t *testing.T) {
	t.Parallel()

	e := New()
	addr := common.HexToAddress("0x1234")

	p := &HarnessPolicy{
		MaxTxAmount: big.NewInt(1000),
		DailyLimit:  big.NewInt(5000),
	}
	e.SetPolicy(addr, p)

	got, ok := e.GetPolicy(addr)
	require.True(t, ok)
	assert.Equal(t, 0, got.MaxTxAmount.Cmp(big.NewInt(1000)))
}

func TestEngine_GetPolicy_NotFound(t *testing.T) {
	t.Parallel()

	e := New()
	_, ok := e.GetPolicy(common.HexToAddress("0x9999"))
	assert.False(t, ok)
}

func TestEngine_Validate_Pass(t *testing.T) {
	t.Parallel()

	e := New()
	addr := common.HexToAddress("0x1234")
	target := common.HexToAddress("0xaaaa")

	e.SetPolicy(addr, &HarnessPolicy{
		MaxTxAmount:    big.NewInt(1000),
		AllowedTargets: []common.Address{target},
	})

	err := e.Validate(addr, &sa.ContractCall{
		Target: target,
		Value:  big.NewInt(500),
	})
	require.NoError(t, err)
}

func TestEngine_Validate_NoPolicySet(t *testing.T) {
	t.Parallel()

	e := New()
	err := e.Validate(common.HexToAddress("0x1234"), &sa.ContractCall{
		Target: common.HexToAddress("0xaaaa"),
		Value:  big.NewInt(100),
	})
	assert.ErrorIs(t, err, sa.ErrPolicyViolation)
}

func TestEngine_Validate_TargetNotAllowed(t *testing.T) {
	t.Parallel()

	e := New()
	addr := common.HexToAddress("0x1234")

	e.SetPolicy(addr, &HarnessPolicy{
		AllowedTargets: []common.Address{common.HexToAddress("0xaaaa")},
	})

	err := e.Validate(addr, &sa.ContractCall{
		Target: common.HexToAddress("0xbbbb"),
		Value:  big.NewInt(0),
	})
	assert.ErrorIs(t, err, sa.ErrTargetNotAllowed)
}

func TestEngine_RecordSpend(t *testing.T) {
	t.Parallel()

	e := New()
	addr := common.HexToAddress("0x1234")

	e.SetPolicy(addr, &HarnessPolicy{
		DailyLimit:   big.NewInt(1000),
		MonthlyLimit: big.NewInt(10000),
	})

	e.RecordSpend(addr, big.NewInt(200))
	e.RecordSpend(addr, big.NewInt(300))

	// Validate that cumulative spend is tracked.
	err := e.Validate(addr, &sa.ContractCall{
		Target: common.Address{},
		Value:  big.NewInt(600),
	})
	assert.ErrorIs(t, err, sa.ErrSpendLimitExceeded)
}

func TestEngine_RecordSpend_NoPolicy(t *testing.T) {
	t.Parallel()

	e := New()
	addr := common.HexToAddress("0x1234")

	// Recording spend without a policy should not panic.
	e.RecordSpend(addr, big.NewInt(100))
}

func TestMergePolicies(t *testing.T) {
	t.Parallel()

	addrA := common.HexToAddress("0xaaaa")
	addrB := common.HexToAddress("0xbbbb")

	tests := []struct {
		give      string
		master    *HarnessPolicy
		task      *HarnessPolicy
		wantCheck func(*testing.T, *HarnessPolicy)
	}{
		{
			give: "smaller limits from master",
			master: &HarnessPolicy{
				MaxTxAmount:  big.NewInt(100),
				DailyLimit:   big.NewInt(500),
				MonthlyLimit: big.NewInt(5000),
			},
			task: &HarnessPolicy{
				MaxTxAmount:  big.NewInt(200),
				DailyLimit:   big.NewInt(1000),
				MonthlyLimit: big.NewInt(10000),
			},
			wantCheck: func(t *testing.T, p *HarnessPolicy) {
				assert.Equal(t, 0, p.MaxTxAmount.Cmp(big.NewInt(100)))
				assert.Equal(t, 0, p.DailyLimit.Cmp(big.NewInt(500)))
				assert.Equal(t, 0, p.MonthlyLimit.Cmp(big.NewInt(5000)))
			},
		},
		{
			give: "smaller limits from task",
			master: &HarnessPolicy{
				MaxTxAmount: big.NewInt(200),
			},
			task: &HarnessPolicy{
				MaxTxAmount: big.NewInt(100),
			},
			wantCheck: func(t *testing.T, p *HarnessPolicy) {
				assert.Equal(t, 0, p.MaxTxAmount.Cmp(big.NewInt(100)))
			},
		},
		{
			give: "target intersection",
			master: &HarnessPolicy{
				AllowedTargets: []common.Address{addrA, addrB},
			},
			task: &HarnessPolicy{
				AllowedTargets: []common.Address{addrA},
			},
			wantCheck: func(t *testing.T, p *HarnessPolicy) {
				require.Len(t, p.AllowedTargets, 1)
				assert.Equal(t, addrA, p.AllowedTargets[0])
			},
		},
		{
			give: "function intersection",
			master: &HarnessPolicy{
				AllowedFunctions: []string{"0x11111111", "0x22222222"},
			},
			task: &HarnessPolicy{
				AllowedFunctions: []string{"0x22222222", "0x33333333"},
			},
			wantCheck: func(t *testing.T, p *HarnessPolicy) {
				require.Len(t, p.AllowedFunctions, 1)
				assert.Equal(t, "0x22222222", p.AllowedFunctions[0])
			},
		},
		{
			give: "higher risk score wins",
			master: &HarnessPolicy{
				RequiredRiskScore: 0.5,
			},
			task: &HarnessPolicy{
				RequiredRiskScore: 0.8,
			},
			wantCheck: func(t *testing.T, p *HarnessPolicy) {
				assert.Equal(t, 0.8, p.RequiredRiskScore)
			},
		},
		{
			give: "nil limits propagated from master",
			master: &HarnessPolicy{
				MaxTxAmount: big.NewInt(100),
			},
			task: &HarnessPolicy{
				MaxTxAmount: nil,
			},
			wantCheck: func(t *testing.T, p *HarnessPolicy) {
				assert.Equal(t, 0, p.MaxTxAmount.Cmp(big.NewInt(100)))
			},
		},
		{
			give:   "both nil limits stay nil",
			master: &HarnessPolicy{},
			task:   &HarnessPolicy{},
			wantCheck: func(t *testing.T, p *HarnessPolicy) {
				assert.Nil(t, p.MaxTxAmount)
				assert.Nil(t, p.DailyLimit)
				assert.Nil(t, p.MonthlyLimit)
			},
		},
		{
			give: "master targets only inherits to result",
			master: &HarnessPolicy{
				AllowedTargets: []common.Address{addrA},
			},
			task: &HarnessPolicy{},
			wantCheck: func(t *testing.T, p *HarnessPolicy) {
				require.Len(t, p.AllowedTargets, 1)
				assert.Equal(t, addrA, p.AllowedTargets[0])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			result := MergePolicies(tt.master, tt.task)
			tt.wantCheck(t, result)
		})
	}
}
