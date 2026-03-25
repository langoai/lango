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
