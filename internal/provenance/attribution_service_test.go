package provenance

import (
	"context"
	"testing"
	"time"

	"github.com/langoai/lango/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubTokenReader struct {
	items []observability.TokenUsage
}

func (s *stubTokenReader) QueryBySession(_ context.Context, _ string) ([]observability.TokenUsage, error) {
	return s.items, nil
}

func TestAttributionService_Report(t *testing.T) {
	store := NewMemoryAttributionStore()
	cpStore := NewMemoryStore()
	tokenReader := &stubTokenReader{
		items: []observability.TokenUsage{
			{SessionKey: "sess-1", AgentName: "operator", InputTokens: 10, OutputTokens: 5, TotalTokens: 15},
		},
	}
	svc := NewAttributionService(store, cpStore, tokenReader)

	ctx := context.Background()
	require.NoError(t, cpStore.SaveCheckpoint(ctx, Checkpoint{
		ID:         "cp-1",
		SessionKey: "sess-1",
		Label:      "cp",
		Trigger:    TriggerManual,
		CreatedAt:  time.Now(),
	}))
	require.NoError(t, svc.RecordWorkspaceOperation(ctx, "sess-1", "", "ws-1", AuthorAgent, "operator", "abc", "", AttributionSourceWorkspaceMerge, []GitFileStat{
		{FilePath: "main.go", LinesAdded: 10, LinesRemoved: 2},
	}))

	report, err := svc.Report(ctx, "sess-1")
	require.NoError(t, err)
	assert.Equal(t, 1, report.Checkpoints)
	assert.Equal(t, int64(15), report.TotalTokens.TotalTokens)
	assert.Equal(t, 10, report.ByAuthor["operator"].LinesAdded)
	assert.Equal(t, 1, report.ByAuthor["operator"].FileCount)
	assert.Equal(t, 1, report.ByFile["main.go"].AuthorCount)
}
