package session

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/langoai/lango/internal/search"
	"github.com/langoai/lango/internal/types"
)

// RecallTableName is the dedicated FTS5 table used by session recall. Kept
// separate from the knowledge FTS5 table per the fts5-search-index spec.
const RecallTableName = "fts_session_recall"

var (
	recallEmailPattern = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
	recallLongDigits   = regexp.MustCompile(`\b\d{6,}\b`)
	recallLongSecret   = regexp.MustCompile(`\b(?:[A-Fa-f0-9]{32,}|[A-Za-z0-9_\-]{32,})\b`)
)

// RecallIndex wraps search.FTS5Index for session recall. Each row indexes
// exactly one ended session by its key.
type RecallIndex struct {
	db    *sql.DB
	store *EntStore
	fts   *search.FTS5Index

	mu    sync.Mutex
	ready bool
}

// NewRecallIndex constructs a recall index on top of the given store's
// database. Call EnsureReady() (or IndexSession — which calls it internally)
// before first use.
func NewRecallIndex(store *EntStore) *RecallIndex {
	columns := []string{"summary", "role_mix", "ended_at"}
	return &RecallIndex{
		db:    store.DB(),
		store: store,
		fts:   search.NewFTS5Index(store.DB(), RecallTableName, columns),
	}
}

// EnsureReady ensures the FTS5 table exists. Idempotent.
func (r *RecallIndex) EnsureReady() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ready {
		return nil
	}
	if err := r.fts.EnsureTable(); err != nil {
		return err
	}
	r.ready = true
	return nil
}

// IndexSession produces a one-shot summary of the session identified by key
// and upserts it into the recall FTS table (delete-then-insert). Returns
// without error if the session has no messages.
func (r *RecallIndex) IndexSession(ctx context.Context, key string) error {
	if err := r.EnsureReady(); err != nil {
		return err
	}
	sess, err := r.store.Get(key)
	if err != nil {
		return fmt.Errorf("recall get %q: %w", key, err)
	}
	if sess == nil || len(sess.History) == 0 {
		return nil
	}
	summary := summarizeHistory(sess.History, 2000)
	roleMix := countRoles(sess.History)
	endedAt := time.Now().UTC().Format(time.RFC3339)

	// Delete-then-insert keeps a single row per session_key.
	if err := r.fts.Update(ctx, key, []string{summary, roleMix, endedAt}); err != nil {
		return fmt.Errorf("recall index %q: %w", key, err)
	}
	return nil
}

// ProcessPending runs IndexSession for every session currently marked with
// MetadataKeyEndPending=true. Successful processing clears the flag.
// Failures are logged and leave the flag set for the next sweep.
func (r *RecallIndex) ProcessPending(ctx context.Context) error {
	keys, err := r.store.ListEndPending()
	if err != nil {
		return fmt.Errorf("list end-pending: %w", err)
	}
	for _, k := range keys {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := r.IndexSession(ctx, k); err != nil {
			slog.Warn("recall index session failed", "key", k, "error", err)
			continue
		}
		if err := r.store.ClearEndPending(k); err != nil {
			slog.Warn("recall clear end-pending failed", "key", k, "error", err)
		}
	}
	return nil
}

// Search delegates to the underlying FTS5 index.
func (r *RecallIndex) Search(ctx context.Context, query string, limit int) ([]search.SearchResult, error) {
	if err := r.EnsureReady(); err != nil {
		return nil, err
	}
	return r.fts.Search(ctx, query, limit)
}

// GetSummary returns the stored summary text for a session key.
// Returns an empty string with nil error if no row exists for the key.
func (r *RecallIndex) GetSummary(ctx context.Context, key string) (string, error) {
	if err := r.EnsureReady(); err != nil {
		return "", err
	}
	row := r.db.QueryRowContext(ctx,
		"SELECT summary FROM "+RecallTableName+" WHERE source_id = ? LIMIT 1",
		key,
	)
	var summary string
	if err := row.Scan(&summary); err != nil {
		if err.Error() == "sql: no rows in result set" {
			return "", nil
		}
		return "", fmt.Errorf("recall get summary %q: %w", key, err)
	}
	return summary, nil
}

// summarizeHistory produces a token-bounded plain-text summary of a session's
// messages. Uses the newest messages first, truncating at maxTokens.
func summarizeHistory(history []Message, maxTokens int) string {
	var b strings.Builder
	totalTokens := 0
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		piece := fmt.Sprintf("[%s] %s\n", msg.Role, redactRecallProjection(strings.TrimSpace(msg.Content)))
		pieceTokens := types.EstimateTokens(piece)
		if totalTokens+pieceTokens > maxTokens {
			break
		}
		b.WriteString(piece)
		totalTokens += pieceTokens
	}
	return strings.TrimSpace(b.String())
}

func redactRecallProjection(content string) string {
	content = recallEmailPattern.ReplaceAllString(content, "[email]")
	content = recallLongDigits.ReplaceAllString(content, "[number]")
	content = recallLongSecret.ReplaceAllString(content, "[secret]")
	content = strings.Join(strings.Fields(content), " ")
	if len(content) > 512 {
		content = content[:512]
	}
	return content
}

// countRoles returns a compact role-mix string like "user:8 assistant:8 tool:2".
func countRoles(history []Message) string {
	counts := map[types.MessageRole]int{}
	for _, m := range history {
		counts[m.Role]++
	}
	parts := make([]string, 0, len(counts))
	for role, n := range counts {
		parts = append(parts, fmt.Sprintf("%s:%d", role, n))
	}
	return strings.Join(parts, " ")
}
