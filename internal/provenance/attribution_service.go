package provenance

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/observability"
)

// TokenUsageReader reads persisted token usage records.
type TokenUsageReader interface {
	QueryBySession(ctx context.Context, sessionKey string) ([]observability.TokenUsage, error)
}

// AttributionView is the raw attribution view returned by CLI show/export logic.
type AttributionView struct {
	Attributions   []Attribution           `json:"attributions"`
	ByAuthorTokens map[string]TokenSummary `json:"by_author_tokens,omitempty"`
	TotalTokens    TokenSummary            `json:"total_tokens"`
	Checkpoints    int                     `json:"checkpoints"`
}

// AttributionService builds provenance views and reports.
type AttributionService struct {
	store       AttributionStore
	checkpoints CheckpointStore
	tokenUsage  TokenUsageReader
}

// GitFileStat describes a per-file git delta used for attribution capture.
type GitFileStat struct {
	FilePath     string
	LinesAdded   int
	LinesRemoved int
}

// NewAttributionService creates a new attribution service.
func NewAttributionService(store AttributionStore, checkpoints CheckpointStore, tokenUsage TokenUsageReader) *AttributionService {
	return &AttributionService{
		store:       store,
		checkpoints: checkpoints,
		tokenUsage:  tokenUsage,
	}
}

// Save records a single attribution row.
func (s *AttributionService) Save(ctx context.Context, attr Attribution) error {
	if attr.ID == "" {
		attr.ID = uuid.New().String()
	}
	if attr.CreatedAt.IsZero() {
		attr.CreatedAt = now()
	}
	return s.store.SaveAttribution(ctx, attr)
}

// RecordWorkspaceOperation persists git-aware attribution evidence.
func (s *AttributionService) RecordWorkspaceOperation(
	ctx context.Context,
	sessionKey, runID, workspaceID string,
	authorType AuthorType,
	authorID, commitHash, stepID string,
	source AttributionSource,
	stats []GitFileStat,
) error {
	if len(stats) == 0 {
		return s.Save(ctx, Attribution{
			SessionKey:  sessionKey,
			RunID:       runID,
			WorkspaceID: workspaceID,
			AuthorType:  authorType,
			AuthorID:    authorID,
			CommitHash:  commitHash,
			StepID:      stepID,
			Source:      source,
		})
	}

	for _, stat := range stats {
		if err := s.Save(ctx, Attribution{
			SessionKey:   sessionKey,
			RunID:        runID,
			WorkspaceID:  workspaceID,
			AuthorType:   authorType,
			AuthorID:     authorID,
			FilePath:     stat.FilePath,
			CommitHash:   commitHash,
			StepID:       stepID,
			Source:       source,
			LinesAdded:   stat.LinesAdded,
			LinesRemoved: stat.LinesRemoved,
		}); err != nil {
			return err
		}
	}
	return nil
}

// View returns the raw attribution rows plus token/checkpoint rollups.
func (s *AttributionService) View(ctx context.Context, sessionKey string, limit int) (*AttributionView, error) {
	rows, err := s.store.ListBySession(ctx, sessionKey, limit)
	if err != nil {
		return nil, fmt.Errorf("list attributions: %w", err)
	}

	view := &AttributionView{
		Attributions:   rows,
		ByAuthorTokens: make(map[string]TokenSummary),
	}

	if s.tokenUsage != nil {
		usages, err := s.tokenUsage.QueryBySession(ctx, sessionKey)
		if err != nil {
			return nil, fmt.Errorf("query token usage: %w", err)
		}
		for _, usage := range usages {
			key := usage.AgentName
			if key == "" {
				key = "unknown"
			}
			sum := view.ByAuthorTokens[key]
			sum.InputTokens += usage.InputTokens
			sum.OutputTokens += usage.OutputTokens
			sum.TotalTokens += usage.TotalTokens
			view.ByAuthorTokens[key] = sum
			view.TotalTokens.InputTokens += usage.InputTokens
			view.TotalTokens.OutputTokens += usage.OutputTokens
			view.TotalTokens.TotalTokens += usage.TotalTokens
		}
	}

	if s.checkpoints != nil {
		count, err := s.checkpoints.CountBySession(ctx, sessionKey)
		if err != nil {
			return nil, fmt.Errorf("count checkpoints: %w", err)
		}
		view.Checkpoints = count
	}

	return view, nil
}

// Report builds an aggregated attribution report for a session.
func (s *AttributionService) Report(ctx context.Context, sessionKey string) (*AttributionReport, error) {
	view, err := s.View(ctx, sessionKey, 0)
	if err != nil {
		return nil, err
	}

	report := &AttributionReport{
		SessionKey:  sessionKey,
		ByAuthor:    make(map[string]AuthorStats),
		ByFile:      make(map[string]FileStats),
		TotalTokens: view.TotalTokens,
		Checkpoints: view.Checkpoints,
		GeneratedAt: now(),
	}

	fileAuthors := make(map[string]map[string]struct{})
	authorFiles := make(map[string]map[string]struct{})

	for _, row := range view.Attributions {
		stats := report.ByAuthor[row.AuthorID]
		stats.AuthorType = row.AuthorType
		stats.LinesAdded += row.LinesAdded
		stats.LinesRemoved += row.LinesRemoved
		report.ByAuthor[row.AuthorID] = stats

		if row.FilePath != "" {
			fs := report.ByFile[row.FilePath]
			fs.LinesAdded += row.LinesAdded
			fs.LinesRemoved += row.LinesRemoved
			report.ByFile[row.FilePath] = fs

			if fileAuthors[row.FilePath] == nil {
				fileAuthors[row.FilePath] = make(map[string]struct{})
			}
			fileAuthors[row.FilePath][row.AuthorID] = struct{}{}

			if authorFiles[row.AuthorID] == nil {
				authorFiles[row.AuthorID] = make(map[string]struct{})
			}
			authorFiles[row.AuthorID][row.FilePath] = struct{}{}
		}
	}

	for authorID, tokens := range view.ByAuthorTokens {
		stats := report.ByAuthor[authorID]
		if stats.AuthorType == "" {
			stats.AuthorType = AuthorAgent
		}
		stats.TokensUsed.InputTokens += tokens.InputTokens
		stats.TokensUsed.OutputTokens += tokens.OutputTokens
		stats.TokensUsed.TotalTokens += tokens.TotalTokens
		report.ByAuthor[authorID] = stats
	}

	for filePath, authors := range fileAuthors {
		fs := report.ByFile[filePath]
		fs.AuthorCount = len(authors)
		report.ByFile[filePath] = fs
	}
	for authorID, files := range authorFiles {
		stats := report.ByAuthor[authorID]
		stats.FileCount = len(files)
		report.ByAuthor[authorID] = stats
	}

	return report, nil
}

var now = func() time.Time { return time.Now() }

// SortedAttributions returns a copy ordered by created_at ascending for canonical bundles.
func SortedAttributions(items []Attribution) []Attribution {
	out := append([]Attribution(nil), items...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

// SortedCheckpoints returns a copy ordered by created_at ascending for canonical bundles.
func SortedCheckpoints(items []Checkpoint) []Checkpoint {
	out := append([]Checkpoint(nil), items...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

// SortedSessionNodes returns a copy ordered by depth, then created_at, then session key.
func SortedSessionNodes(items []SessionNode) []SessionNode {
	out := append([]SessionNode(nil), items...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Depth != out[j].Depth {
			return out[i].Depth < out[j].Depth
		}
		if !out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].CreatedAt.Before(out[j].CreatedAt)
		}
		return out[i].SessionKey < out[j].SessionKey
	})
	return out
}
