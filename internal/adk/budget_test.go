package adk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContextBudgetManager_ValidAllocation(t *testing.T) {
	alloc := DefaultAllocation()
	bm, err := NewContextBudgetManager(128000, 4096, 2000, alloc)
	require.NoError(t, err)
	require.NotNil(t, bm)
}

func TestNewContextBudgetManager_InvalidAllocationSum(t *testing.T) {
	tests := []struct {
		give SectionAllocation
		want string
	}{
		{
			give: SectionAllocation{Knowledge: 0.50, RAG: 0.25, Memory: 0.25, RunSummary: 0.10, Headroom: 0.10},
			want: "allocation sum",
		},
		{
			give: SectionAllocation{Knowledge: 0.20, RAG: 0.10, Memory: 0.10, RunSummary: 0.05, Headroom: 0.05},
			want: "allocation sum",
		},
	}

	for _, tt := range tests {
		_, err := NewContextBudgetManager(128000, 4096, 2000, tt.give)
		require.Error(t, err)
		assert.Contains(t, err.Error(), tt.want)
	}
}

func TestSectionBudgets_StandardModels(t *testing.T) {
	alloc := DefaultAllocation()

	tests := []struct {
		give         string
		giveWindow   int
		giveReserve  int
		giveBase     int
		wantPositive bool
	}{
		{give: "8k model", giveWindow: 8192, giveReserve: 1024, giveBase: 1000, wantPositive: true},
		{give: "32k model", giveWindow: 32000, giveReserve: 4096, giveBase: 2000, wantPositive: true},
		{give: "128k model", giveWindow: 128000, giveReserve: 4096, giveBase: 2000, wantPositive: true},
		{give: "200k model", giveWindow: 200000, giveReserve: 4096, giveBase: 2000, wantPositive: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			bm, err := NewContextBudgetManager(tt.giveWindow, tt.giveReserve, tt.giveBase, alloc)
			require.NoError(t, err)

			budgets := bm.SectionBudgets()
			if tt.wantPositive {
				assert.Greater(t, budgets.Knowledge, 0)
				assert.Greater(t, budgets.RAG, 0)
				assert.Greater(t, budgets.Memory, 0)
				assert.Greater(t, budgets.RunSummary, 0)
			}

			// Budget ratios should be proportional to allocation.
			available := tt.giveWindow - bm.responseReserve - tt.giveBase
			if available > 0 {
				assert.InDelta(t, float64(available)*0.30, float64(budgets.Knowledge), 1)
				assert.InDelta(t, float64(available)*0.25, float64(budgets.RAG), 1)
				assert.InDelta(t, float64(available)*0.25, float64(budgets.Memory), 1)
				assert.InDelta(t, float64(available)*0.10, float64(budgets.RunSummary), 1)
			}
		})
	}
}

func TestSectionBudgets_Degradation(t *testing.T) {
	alloc := DefaultAllocation()

	tests := []struct {
		give       string
		giveWindow int
		giveBase   int
	}{
		{give: "negative available", giveWindow: 4096, giveBase: 5000},
		{give: "zero available", giveWindow: 5120, giveBase: 4096}, // reserve clamps to 1024, 5120-1024-4096=0
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			bm, err := NewContextBudgetManager(tt.giveWindow, 0, tt.giveBase, alloc)
			require.NoError(t, err)

			budgets := bm.SectionBudgets()
			assert.Equal(t, 0, budgets.Knowledge, "knowledge should be 0 (unlimited)")
			assert.Equal(t, 0, budgets.RAG, "rag should be 0 (unlimited)")
			assert.Equal(t, 0, budgets.Memory, "memory should be 0 (unlimited)")
			assert.Equal(t, 0, budgets.RunSummary, "runSummary should be 0 (unlimited)")
		})
	}
}

func TestLookupModelWindow(t *testing.T) {
	tests := []struct {
		give string
		want int
	}{
		{give: "gemini-2.0-flash", want: 1000000},
		{give: "gemini-2.0-flash-001", want: 1000000},
		{give: "gemini-1.5-pro", want: 2000000},
		{give: "claude-sonnet-4-5-20250929", want: 200000},
		{give: "claude-opus-4-0", want: 200000},
		{give: "claude-haiku-3-5", want: 200000},
		{give: "gpt-4o-2024-08-06", want: 128000},
		{give: "gpt-4o-mini", want: 128000},
		{give: "gpt-4-turbo-preview", want: 128000},
		{give: "gpt-4", want: 8192},
		{give: "llama3.1", want: 128000},
		{give: "o1-preview", want: 200000},
		{give: "custom-model-v1", want: 128000}, // fallback
		{give: "", want: 128000},                 // empty = fallback
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := LookupModelWindow(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResponseReserveClamping(t *testing.T) {
	alloc := DefaultAllocation()

	t.Run("zero reserve uses default 4096", func(t *testing.T) {
		bm, err := NewContextBudgetManager(128000, 0, 0, alloc)
		require.NoError(t, err)
		assert.Equal(t, 4096, bm.responseReserve)
	})

	t.Run("small reserve clamps to 1024", func(t *testing.T) {
		bm, err := NewContextBudgetManager(128000, 500, 0, alloc)
		require.NoError(t, err)
		assert.Equal(t, 1024, bm.responseReserve)
	})

	t.Run("large reserve clamps to 25% of window", func(t *testing.T) {
		bm, err := NewContextBudgetManager(8192, 8000, 0, alloc)
		require.NoError(t, err)
		assert.Equal(t, 2048, bm.responseReserve) // 25% of 8192
	})
}

func TestDefaultAllocation(t *testing.T) {
	alloc := DefaultAllocation()
	assert.InDelta(t, 1.0, alloc.sum(), 0.001)
	assert.Equal(t, 0.30, alloc.Knowledge)
	assert.Equal(t, 0.25, alloc.RAG)
	assert.Equal(t, 0.25, alloc.Memory)
	assert.Equal(t, 0.10, alloc.RunSummary)
	assert.Equal(t, 0.10, alloc.Headroom)
}

