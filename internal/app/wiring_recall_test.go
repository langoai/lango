package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/search"
)

func TestRecallProviderAdapter_ReturnsErrorWhenSummaryLookupFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	backend := &stubRecallBackend{
		results: []search.SearchResult{{
			RowID: "sess-prior",
			Rank:  -0.8,
		}},
		summaryErrs: map[string]error{
			"sess-prior": errors.New("summary unavailable"),
		},
	}
	adapter := newRecallProviderAdapter(backend, config.ContextRecallConfig{
		TopN:    1,
		MinRank: 0.2,
	})

	matches, err := adapter.RecallRecent(ctx, "sess-current", "deployment notes")

	require.Error(t, err)
	require.Nil(t, matches)
	require.ErrorContains(t, err, "sess-prior")
	require.ErrorContains(t, err, "summary unavailable")
}

func TestRecallProviderAdapter_FiltersCurrentSessionAndRankBeforeTopN(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	backend := &stubRecallBackend{
		results: []search.SearchResult{
			{RowID: "sess-current", Rank: -0.9},
			{RowID: "sess-low", Rank: -0.1},
			{RowID: "sess-first", Rank: -0.8},
			{RowID: "sess-second", Rank: -0.7},
			{RowID: "sess-third", Rank: -0.6},
		},
		summaries: map[string]string{
			"sess-current": "current summary",
			"sess-low":     "low summary",
			"sess-first":   "first summary",
			"sess-second":  "second summary",
			"sess-third":   "third summary",
		},
	}
	adapter := newRecallProviderAdapter(backend, config.ContextRecallConfig{
		TopN:    2,
		MinRank: 0.2,
	})

	matches, err := adapter.RecallRecent(ctx, "sess-current", "deployment notes")

	require.NoError(t, err)
	require.Len(t, matches, 2)
	require.Equal(t, "sess-first", matches[0].SessionKey)
	require.Equal(t, "first summary", matches[0].Summary)
	require.Equal(t, 0.8, matches[0].Rank)
	require.Equal(t, "sess-second", matches[1].SessionKey)
	require.Equal(t, "second summary", matches[1].Summary)
	require.Equal(t, 0.7, matches[1].Rank)
	require.Equal(t, []string{"sess-first", "sess-second"}, backend.summaryKeys)
	require.Equal(t, 4, backend.searchLimit)
}

type stubRecallBackend struct {
	results     []search.SearchResult
	searchErr   error
	summaries   map[string]string
	summaryErrs map[string]error
	summaryKeys []string
	searchLimit int
}

func (s *stubRecallBackend) Search(ctx context.Context, query string, limit int) ([]search.SearchResult, error) {
	s.searchLimit = limit
	return s.results, s.searchErr
}

func (s *stubRecallBackend) GetSummary(ctx context.Context, key string) (string, error) {
	s.summaryKeys = append(s.summaryKeys, key)
	if err := s.summaryErrs[key]; err != nil {
		return "", err
	}
	return s.summaries[key], nil
}

func (s *stubRecallBackend) ProcessPending(ctx context.Context) error {
	return nil
}

func (s *stubRecallBackend) IndexSession(ctx context.Context, key string) error {
	return nil
}
