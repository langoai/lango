package team

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
)

func TestTeamForm_RequiresGoalAndMemberCountParameters(t *testing.T) {
	t.Parallel()

	coord, _ := setupCoordinator(t)
	tools := BuildTools(coord)
	tool := findToolByName(tools, "team_form")
	require.NotNil(t, tool)

	cases := []struct {
		name      string
		params    map[string]interface{}
		wantError string
	}{
		{
			name: "missing goal",
			params: map[string]interface{}{
				"name":        "search-team",
				"capability":  "search",
				"memberCount": float64(2),
				"leaderDid":   "did:leader",
			},
			wantError: "missing goal parameter",
		},
		{
			name: "missing memberCount",
			params: map[string]interface{}{
				"name":       "search-team",
				"goal":       "find information",
				"capability": "search",
				"leaderDid":  "did:leader",
			},
			wantError: "missing memberCount parameter",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tool.Handler(context.Background(), tt.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantError)
		})
	}
}

func TestTeamFormWithBudget_RequiresCanonicalInputs(t *testing.T) {
	t.Parallel()

	coord, _ := setupCoordinator(t)
	tools := BuildEscrowTools(coord, nil, nil)
	tool := findToolByName(tools, "team_form_with_budget")
	require.NotNil(t, tool)

	cases := []struct {
		name      string
		params    map[string]interface{}
		wantError string
	}{
		{
			name: "missing goal",
			params: map[string]interface{}{
				"name":        "budget-team",
				"capability":  "search",
				"memberCount": float64(2),
				"leaderDid":   "did:leader",
				"budget":      float64(10),
			},
			wantError: "missing goal parameter",
		},
		{
			name: "missing memberCount",
			params: map[string]interface{}{
				"name":       "budget-team",
				"goal":       "fund a team",
				"capability": "search",
				"leaderDid":  "did:leader",
				"budget":     float64(10),
			},
			wantError: "missing memberCount parameter",
		},
		{
			name: "missing budget",
			params: map[string]interface{}{
				"name":        "budget-team",
				"goal":        "fund a team",
				"capability":  "search",
				"memberCount": float64(2),
				"leaderDid":   "did:leader",
			},
			wantError: "missing budget parameter",
		},
		{
			name: "non-positive budget",
			params: map[string]interface{}{
				"name":        "budget-team",
				"goal":        "fund a team",
				"capability":  "search",
				"memberCount": float64(2),
				"leaderDid":   "did:leader",
				"budget":      float64(0),
			},
			wantError: "budget must be greater than zero",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tool.Handler(context.Background(), tt.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantError)
		})
	}
}

func TestTeamCompleteMilestone_RequiresEscrowAndMilestoneIDs(t *testing.T) {
	t.Parallel()

	coord, _ := setupCoordinator(t)
	tools := BuildEscrowTools(coord, nil, nil)
	tool := findToolByName(tools, "team_complete_milestone")
	require.NotNil(t, tool)

	cases := []struct {
		name      string
		params    map[string]interface{}
		wantError string
	}{
		{
			name: "missing escrowId",
			params: map[string]interface{}{
				"milestoneId": "milestone-1",
			},
			wantError: "missing escrowId parameter",
		},
		{
			name: "missing milestoneId",
			params: map[string]interface{}{
				"escrowId": "escrow-1",
			},
			wantError: "missing milestoneId parameter",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tool.Handler(context.Background(), tt.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantError)
		})
	}
}

func findToolByName(tools []*agent.Tool, name string) *agent.Tool {
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	return nil
}
