package budget

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTaskBudget_Remaining(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give         string
		giveTotal    int64
		giveSpent    int64
		giveReserved int64
		want         int64
	}{
		{
			give:         "full budget remaining",
			giveTotal:    1000000,
			giveSpent:    0,
			giveReserved: 0,
			want:         1000000,
		},
		{
			give:         "partial spend",
			giveTotal:    1000000,
			giveSpent:    300000,
			giveReserved: 0,
			want:         700000,
		},
		{
			give:         "partial reserve",
			giveTotal:    1000000,
			giveSpent:    0,
			giveReserved: 200000,
			want:         800000,
		},
		{
			give:         "spend and reserve",
			giveTotal:    1000000,
			giveSpent:    300000,
			giveReserved: 200000,
			want:         500000,
		},
		{
			give:         "fully spent",
			giveTotal:    1000000,
			giveSpent:    1000000,
			giveReserved: 0,
			want:         0,
		},
		{
			give:         "overspent returns negative",
			giveTotal:    1000000,
			giveSpent:    1100000,
			giveReserved: 0,
			want:         -100000,
		},
		{
			give:         "zero budget",
			giveTotal:    0,
			giveSpent:    0,
			giveReserved: 0,
			want:         0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			tb := &TaskBudget{
				TotalBudget: big.NewInt(tt.giveTotal),
				Spent:       big.NewInt(tt.giveSpent),
				Reserved:    big.NewInt(tt.giveReserved),
			}
			got := tb.Remaining()
			want := big.NewInt(tt.want)
			assert.Equal(t, 0, got.Cmp(want), "Remaining() = %s, want %s", got, want)
		})
	}
}

func TestTaskBudget_Remaining_DoesNotMutateFields(t *testing.T) {
	t.Parallel()

	tb := &TaskBudget{
		TotalBudget: big.NewInt(1000000),
		Spent:       big.NewInt(300000),
		Reserved:    big.NewInt(200000),
	}

	_ = tb.Remaining()

	assert.Equal(t, 0, tb.TotalBudget.Cmp(big.NewInt(1000000)))
	assert.Equal(t, 0, tb.Spent.Cmp(big.NewInt(300000)))
	assert.Equal(t, 0, tb.Reserved.Cmp(big.NewInt(200000)))
}
