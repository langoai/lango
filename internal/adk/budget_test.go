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

func TestReallocateBudgets(t *testing.T) {
	alloc := DefaultAllocation()
	bm, err := NewContextBudgetManager(128000, 4096, 2000, alloc)
	require.NoError(t, err)
	base := bm.SectionBudgets()

	tests := []struct {
		give         string
		measured     SectionTokens
		wantChanged  bool
		wantKnGt     int  // Knowledge budget > this value
		wantRAG      int  // RAG budget exact
		wantDegraded bool
	}{
		{
			give:        "all sections present — no reallocation",
			measured:    SectionTokens{Knowledge: 500, RAG: 300, Memory: 1000, RunSummary: 100},
			wantChanged: false,
		},
		{
			give:        "one section empty — surplus redistributed",
			measured:    SectionTokens{Knowledge: 500, RAG: 0, Memory: 1000, RunSummary: 100},
			wantChanged: true,
			wantKnGt:    base.Knowledge, // Knowledge should be > initial
			wantRAG:     0,              // RAG donated
		},
		{
			give:        "two sections empty — both donate",
			measured:    SectionTokens{Knowledge: 500, RAG: 0, Memory: 1000, RunSummary: 0},
			wantChanged: true,
			wantKnGt:    base.Knowledge,
			wantRAG:     0,
		},
		{
			give:        "all sections empty — all-zero budgets",
			measured:    SectionTokens{Knowledge: 0, RAG: 0, Memory: 0, RunSummary: 0},
			wantChanged: true,
			wantRAG:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result := bm.ReallocateBudgets(tt.measured)

			if !tt.wantChanged {
				assert.Equal(t, base.Knowledge, result.Knowledge)
				assert.Equal(t, base.RAG, result.RAG)
				assert.Equal(t, base.Memory, result.Memory)
				assert.Equal(t, base.RunSummary, result.RunSummary)
				return
			}

			assert.Equal(t, tt.wantRAG, result.RAG)
			if tt.wantKnGt > 0 {
				assert.Greater(t, result.Knowledge, tt.wantKnGt,
					"knowledge should receive surplus")
			}
		})
	}
}

func TestReallocateBudgets_Degraded(t *testing.T) {
	alloc := DefaultAllocation()
	bm, err := NewContextBudgetManager(4096, 0, 5000, alloc) // negative available
	require.NoError(t, err)

	result := bm.ReallocateBudgets(SectionTokens{Knowledge: 100})
	assert.True(t, result.Degraded, "degraded should pass through")
}

func TestReallocateBudgets_ProportionalDistribution(t *testing.T) {
	alloc := DefaultAllocation()
	bm, err := NewContextBudgetManager(128000, 4096, 2000, alloc)
	require.NoError(t, err)
	base := bm.SectionBudgets()

	// Only RAG empty → 25% surplus to K/M/RS
	result := bm.ReallocateBudgets(SectionTokens{Knowledge: 500, RAG: 0, Memory: 1000, RunSummary: 100})

	surplus := base.RAG
	presentRatioSum := 0.30 + 0.25 + 0.10 // Knowledge + Memory + RunSummary

	wantKnowledge := base.Knowledge + int(float64(surplus)*0.30/presentRatioSum)
	wantMemory := base.Memory + int(float64(surplus)*0.25/presentRatioSum)
	wantRunSummary := base.RunSummary + int(float64(surplus)*0.10/presentRatioSum)

	assert.Equal(t, wantKnowledge, result.Knowledge, "knowledge share")
	assert.Equal(t, wantMemory, result.Memory, "memory share")
	assert.Equal(t, wantRunSummary, result.RunSummary, "runSummary share")
	assert.Equal(t, 0, result.RAG, "RAG donated")
}

func TestReallocateBudgets_AllEmpty(t *testing.T) {
	alloc := DefaultAllocation()
	bm, err := NewContextBudgetManager(128000, 4096, 2000, alloc)
	require.NoError(t, err)

	result := bm.ReallocateBudgets(SectionTokens{})
	assert.Equal(t, 0, result.Knowledge)
	assert.Equal(t, 0, result.RAG)
	assert.Equal(t, 0, result.Memory)
	assert.Equal(t, 0, result.RunSummary)
	assert.False(t, result.Degraded, "all-empty is not degradation")
}

