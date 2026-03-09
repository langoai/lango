package hub

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnChainDealStatus_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give OnChainDealStatus
		want string
	}{
		{give: DealStatusCreated, want: "created"},
		{give: DealStatusDeposited, want: "deposited"},
		{give: DealStatusWorkSubmitted, want: "work_submitted"},
		{give: DealStatusReleased, want: "released"},
		{give: DealStatusRefunded, want: "refunded"},
		{give: DealStatusDisputed, want: "disputed"},
		{give: DealStatusResolved, want: "resolved"},
		{give: OnChainDealStatus(99), want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.give.String())
		})
	}
}

func TestOnChainDealStatus_Values(t *testing.T) {
	t.Parallel()

	assert.Equal(t, OnChainDealStatus(0), DealStatusCreated)
	assert.Equal(t, OnChainDealStatus(1), DealStatusDeposited)
	assert.Equal(t, OnChainDealStatus(2), DealStatusWorkSubmitted)
	assert.Equal(t, OnChainDealStatus(3), DealStatusReleased)
	assert.Equal(t, OnChainDealStatus(4), DealStatusRefunded)
	assert.Equal(t, OnChainDealStatus(5), DealStatusDisputed)
	assert.Equal(t, OnChainDealStatus(6), DealStatusResolved)
}
