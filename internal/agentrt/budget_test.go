package agentrt

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/config"
)

func TestBudgetPolicy_RecordTurn(t *testing.T) {
	bp := NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   10,
		DelegationLimit: 5,
		AlertThreshold:  0.8,
	})

	for i := 0; i < 5; i++ {
		bp.RecordTurn()
	}
	assert.Equal(t, 5, bp.TurnCount())
	assert.Equal(t, 0, bp.DelegationCount())
}

func TestBudgetPolicy_RecordDelegation(t *testing.T) {
	bp := NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   50,
		DelegationLimit: 15,
		AlertThreshold:  0.8,
	})

	bp.RecordDelegation("operator")
	bp.RecordDelegation("navigator")
	bp.RecordDelegation("operator") // duplicate agent

	assert.Equal(t, 3, bp.DelegationCount())
	assert.Equal(t, 2, bp.UniqueAgentCount())
}

func TestBudgetPolicy_RecordDelegation_RootExcludedFromUniqueAgents(t *testing.T) {
	bp := NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   50,
		DelegationLimit: 15,
		AlertThreshold:  0.8,
	})

	bp.RecordDelegation("operator")
	bp.RecordDelegation("lango-orchestrator")

	assert.Equal(t, 2, bp.DelegationCount())
	assert.Equal(t, 1, bp.UniqueAgentCount())
}

func TestBudgetPolicy_AlertThreshold(t *testing.T) {
	var alerts []BudgetAlert
	bp := NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   10,
		DelegationLimit: 5,
		AlertThreshold:  0.8,
	})
	bp.SetAlertHandler(func(alert BudgetAlert) {
		alerts = append(alerts, alert)
	})

	// 7 turns = 70% — no alert
	for i := 0; i < 7; i++ {
		bp.RecordTurn()
	}
	assert.Empty(t, alerts)

	// 8 turns = 80% — alert
	bp.RecordTurn()
	assert.Len(t, alerts, 1)
	assert.Equal(t, "turns", alerts[0].Resource)
	assert.Equal(t, 8, alerts[0].Used)
	assert.Equal(t, 10, alerts[0].Limit)

	// Further turns don't re-alert
	bp.RecordTurn()
	assert.Len(t, alerts, 1)
}

func TestBudgetPolicy_DelegationAlert(t *testing.T) {
	var alerts []BudgetAlert
	bp := NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   50,
		DelegationLimit: 5,
		AlertThreshold:  0.8,
	})
	bp.SetAlertHandler(func(alert BudgetAlert) {
		alerts = append(alerts, alert)
	})

	for i := 0; i < 4; i++ {
		bp.RecordDelegation("agent")
	}
	assert.Len(t, alerts, 1)
	assert.Equal(t, "delegations", alerts[0].Resource)
}

func TestBudgetPolicy_Reset(t *testing.T) {
	bp := NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   10,
		DelegationLimit: 5,
		AlertThreshold:  0.8,
	})

	bp.RecordTurn()
	bp.RecordDelegation("op")
	bp.Reset()

	assert.Equal(t, 0, bp.TurnCount())
	assert.Equal(t, 0, bp.DelegationCount())
	assert.Equal(t, 0, bp.UniqueAgentCount())
}

func TestBudgetPolicy_Serialize(t *testing.T) {
	bp := NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   50,
		DelegationLimit: 15,
		AlertThreshold:  0.8,
	})

	for i := 0; i < 5; i++ {
		bp.RecordTurn()
	}
	for i := 0; i < 3; i++ {
		bp.RecordDelegation("agent")
	}

	state := bp.Serialize()
	assert.Equal(t, "5", state["usage:budget_turns"])
	assert.Equal(t, "3", state["usage:budget_delegations"])
	assert.Len(t, state, 2)
}

func TestBudgetPolicy_Restore(t *testing.T) {
	tests := []struct {
		give     map[string]string
		wantTurn int
		wantDel  int
	}{
		{
			give:     map[string]string{"usage:budget_turns": "10", "usage:budget_delegations": "4"},
			wantTurn: 10,
			wantDel:  4,
		},
		{
			give:     map[string]string{},
			wantTurn: 0,
			wantDel:  0,
		},
		{
			give:     map[string]string{"usage:budget_turns": "abc"},
			wantTurn: 0,
			wantDel:  0,
		},
		{
			give:     map[string]string{"usage:budget_turns": "7"},
			wantTurn: 7,
			wantDel:  0,
		},
	}

	for _, tt := range tests {
		bp := NewBudgetPolicy(config.BudgetCfg{
			ToolCallLimit:   50,
			DelegationLimit: 15,
			AlertThreshold:  0.8,
		})
		bp.Restore(tt.give)
		assert.Equal(t, tt.wantTurn, bp.TurnCount())
		assert.Equal(t, tt.wantDel, bp.DelegationCount())
	}
}

func TestBudgetPolicy_SerializeRestoreRoundTrip(t *testing.T) {
	bp := NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   50,
		DelegationLimit: 15,
		AlertThreshold:  0.8,
	})
	for i := 0; i < 12; i++ {
		bp.RecordTurn()
	}
	for i := 0; i < 6; i++ {
		bp.RecordDelegation("agent")
	}

	state := bp.Serialize()

	bp2 := NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   50,
		DelegationLimit: 15,
		AlertThreshold:  0.8,
	})
	bp2.Restore(state)

	assert.Equal(t, 12, bp2.TurnCount())
	assert.Equal(t, 6, bp2.DelegationCount())
}

func TestBudgetPolicy_CloneIsolatesMutableState(t *testing.T) {
	base := NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   10,
		DelegationLimit: 5,
		AlertThreshold:  0.8,
	})
	base.RecordTurn()

	cloneA := base.Clone()
	cloneB := base.Clone()

	cloneA.RecordTurn()
	cloneA.RecordDelegation("operator")

	assert.Equal(t, 1, cloneA.TurnCount())
	assert.Equal(t, 1, cloneA.DelegationCount())
	assert.Equal(t, 0, cloneB.TurnCount())
	assert.Equal(t, 0, cloneB.DelegationCount())
	assert.Equal(t, 1, base.TurnCount())
}
