package ontology_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ontology"
)

func TestResolveConflictToolValidatesRequiredParameters(t *testing.T) {
	ctx := context.Background()
	conflictID := uuid.MustParse("11111111-2222-3333-4444-555555555555")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError string
	}{
		{
			name:      "missing conflict id",
			params:    map[string]interface{}{"winner_object": "fact:1", "reason": "approved"},
			wantError: "missing conflict_id parameter",
		},
		{
			name:      "invalid conflict id",
			params:    map[string]interface{}{"conflict_id": "not-a-uuid", "winner_object": "fact:1", "reason": "approved"},
			wantError: "invalid conflict_id",
		},
		{
			name:      "missing winner object",
			params:    map[string]interface{}{"conflict_id": conflictID.String(), "reason": "approved"},
			wantError: "missing winner_object parameter",
		},
		{
			name:      "missing reason",
			params:    map[string]interface{}{"conflict_id": conflictID.String(), "winner_object": "fact:1"},
			wantError: "missing reason parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &resolveConflictToolService{}
			tool := requireResolveConflictTool(t, ontology.BuildTools(svc, nil))

			result, err := tool.Handler(ctx, tt.params)

			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantError)
			assert.Nil(t, result)
			assert.Zero(t, svc.resolveConflictCalls, "invalid parameters must not call the service")
		})
	}
}

func TestResolveConflictToolCallsServiceAndReturnsResolvedStatus(t *testing.T) {
	ctx := context.Background()
	conflictID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	svc := &resolveConflictToolService{}
	tool := requireResolveConflictTool(t, ontology.BuildTools(svc, nil))

	result, err := tool.Handler(ctx, map[string]interface{}{
		"conflict_id":   conflictID.String(),
		"winner_object": "user:alice",
		"reason":        "manual adjudication",
	})

	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"status": "resolved"}, result)
	assert.Equal(t, 1, svc.resolveConflictCalls)
	assert.Equal(t, conflictID, svc.lastConflictID)
	assert.Equal(t, "user:alice", svc.lastWinnerObject)
	assert.Equal(t, "manual adjudication", svc.lastReason)
}

func TestResolveConflictToolWrapsServiceError(t *testing.T) {
	ctx := context.Background()
	conflictID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	wantErr := errors.New("truth store unavailable")
	svc := &resolveConflictToolService{resolveConflictErr: wantErr}
	tool := requireResolveConflictTool(t, ontology.BuildTools(svc, nil))

	result, err := tool.Handler(ctx, map[string]interface{}{
		"conflict_id":   conflictID.String(),
		"winner_object": "user:alice",
		"reason":        "manual adjudication",
	})

	require.ErrorIs(t, err, wantErr)
	assert.ErrorContains(t, err, "resolve conflict")
	assert.Nil(t, result)
	assert.Equal(t, 1, svc.resolveConflictCalls)
}

func requireResolveConflictTool(t *testing.T, tools []*agent.Tool) *agent.Tool {
	t.Helper()
	for _, tool := range tools {
		if tool.Name == "ontology_resolve_conflict" {
			return tool
		}
	}
	require.FailNow(t, "missing ontology_resolve_conflict tool")
	return nil
}

type resolveConflictToolService struct {
	ontology.OntologyService

	resolveConflictErr   error
	resolveConflictCalls int
	lastConflictID       uuid.UUID
	lastWinnerObject     string
	lastReason           string
}

func (s *resolveConflictToolService) ResolveConflict(_ context.Context, conflictID uuid.UUID, winnerObject, reason string) error {
	s.resolveConflictCalls++
	s.lastConflictID = conflictID
	s.lastWinnerObject = winnerObject
	s.lastReason = reason
	return s.resolveConflictErr
}
