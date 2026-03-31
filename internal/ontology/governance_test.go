package ontology

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFSM_ValidTransitions(t *testing.T) {
	g := NewGovernanceEngine(GovernancePolicy{})

	tests := []struct {
		from SchemaStatus
		to   SchemaStatus
	}{
		{SchemaProposed, SchemaShadow},
		{SchemaProposed, SchemaQuarantined},
		{SchemaShadow, SchemaActive},
		{SchemaShadow, SchemaQuarantined},
		{SchemaQuarantined, SchemaProposed},
		{SchemaActive, SchemaDeprecated},
	}
	for _, tt := range tests {
		t.Run(string(tt.from)+"_to_"+string(tt.to), func(t *testing.T) {
			assert.NoError(t, g.ValidateTransition(tt.from, tt.to))
		})
	}
}

func TestFSM_InvalidTransitions(t *testing.T) {
	g := NewGovernanceEngine(GovernancePolicy{})

	tests := []struct {
		from SchemaStatus
		to   SchemaStatus
	}{
		{SchemaProposed, SchemaActive},    // must go through shadow
		{SchemaActive, SchemaProposed},    // can't un-activate
		{SchemaDeprecated, SchemaActive},  // deprecated is terminal
		{SchemaDeprecated, SchemaProposed},
		{SchemaShadow, SchemaProposed},    // must go through quarantined
		{SchemaQuarantined, SchemaActive}, // must re-propose first
	}
	for _, tt := range tests {
		t.Run(string(tt.from)+"_to_"+string(tt.to), func(t *testing.T) {
			assert.Error(t, g.ValidateTransition(tt.from, tt.to))
		})
	}
}

func TestRateLimit_Within(t *testing.T) {
	g := NewGovernanceEngine(GovernancePolicy{MaxNewPerDay: 5})

	for i := 0; i < 5; i++ {
		require.NoError(t, g.CheckRateLimit(context.Background()))
	}
}

func TestRateLimit_Exceeded(t *testing.T) {
	g := NewGovernanceEngine(GovernancePolicy{MaxNewPerDay: 2})

	require.NoError(t, g.CheckRateLimit(context.Background()))
	require.NoError(t, g.CheckRateLimit(context.Background()))
	assert.Error(t, g.CheckRateLimit(context.Background()))
}

func TestRateLimit_NoLimit(t *testing.T) {
	g := NewGovernanceEngine(GovernancePolicy{MaxNewPerDay: 0})

	for i := 0; i < 100; i++ {
		require.NoError(t, g.CheckRateLimit(context.Background()))
	}
}
