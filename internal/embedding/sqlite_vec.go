package embedding

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
)

func init() {
	sqlite_vec.Auto()
}

// SQLiteVecStore implements VectorStore using sqlite-vec.
type SQLiteVecStore struct {
	mu         sync.Mutex
	db         *sql.DB
	dimensions int
}

// NewSQLiteVecStore creates a new sqlite-vec backed vector store.
// The db connection should be the same SQLite database used by ent.
// Connection pool settings are managed by bootstrap; WAL mode enables
// safe concurrent access for file-based databases.
func NewSQLiteVecStore(db *sql.DB, dimensions int) (*SQLiteVecStore, error) {
	s := &SQLiteVecStore{
		db:         db,
		dimensions: dimensions,
	}
	if err := s.ensureTable(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SQLiteVecStore) ensureTable() error {
	query := fmt.Sprintf(`CREATE VIRTUAL TABLE IF NOT EXISTS vec_embeddings USING vec0(
		collection TEXT NOT NULL,
		source_id  TEXT NOT NULL,
		embedding  float[%d],
		+metadata  TEXT
	)`, s.dimensions)

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("create vec_embeddings table: %w", err)
	}
	return nil
}

// Upsert inserts or replaces vector records.
func (s *SQLiteVecStore) Upsert(ctx context.Context, records []VectorRecord) error {
	if len(records) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete existing records first for upsert semantics
	delStmt, err := tx.PrepareContext(ctx,
		`DELETE FROM vec_embeddings WHERE collection = ? AND source_id = ?`)
	if err != nil {
		return fmt.Errorf("prepare delete: %w", err)
	}
	defer delStmt.Close()

	insStmt, err := tx.PrepareContext(ctx,
		`INSERT INTO vec_embeddings (collection, source_id, embedding, metadata) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer insStmt.Close()

	for _, r := range records {
		if _, err := delStmt.ExecContext(ctx, r.Collection, r.ID); err != nil {
			return fmt.Errorf("delete existing record %s/%s: %w", r.Collection, r.ID, err)
		}

		serialized, err := sqlite_vec.SerializeFloat32(r.Embedding)
		if err != nil {
			return fmt.Errorf("serialize embedding for %s/%s: %w", r.Collection, r.ID, err)
		}

		metaJSON, err := json.Marshal(r.Metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata for %s/%s: %w", r.Collection, r.ID, err)
		}

		if _, err := insStmt.ExecContext(ctx, r.Collection, r.ID, serialized, string(metaJSON)); err != nil {
			return fmt.Errorf("insert record %s/%s: %w", r.Collection, r.ID, err)
		}
	}

	return tx.Commit()
}

// Search finds the most similar vectors in a collection.
// When opts is non-nil and MetadataFilter is set, over-fetches by 3x and post-filters.
func (s *SQLiteVecStore) Search(ctx context.Context, collection string, query []float32, limit int, opts *SearchOptions) ([]SearchResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if limit <= 0 {
		limit = 5
	}

	fetchLimit := limit
	if opts != nil && len(opts.MetadataFilter) > 0 {
		fetchLimit = limit * 3
	}

	serialized, err := sqlite_vec.SerializeFloat32(query)
	if err != nil {
		return nil, fmt.Errorf("serialize query: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT source_id, distance, metadata
		 FROM vec_embeddings
		 WHERE embedding MATCH ? AND collection = ?
		 ORDER BY distance
		 LIMIT ?`,
		serialized, collection, fetchLimit,
	)
	if err != nil {
		return nil, fmt.Errorf("search vec_embeddings: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var metaJSON string
		if err := rows.Scan(&r.ID, &r.Distance, &metaJSON); err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}
		r.Collection = collection

		if metaJSON != "" {
			if err := json.Unmarshal([]byte(metaJSON), &r.Metadata); err != nil {
				return nil, fmt.Errorf("unmarshal metadata: %w", err)
			}
		}

		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Post-filter by metadata if options specified.
	if opts != nil && len(opts.MetadataFilter) > 0 {
		results = filterByMetadata(results, opts.MetadataFilter)
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// filterByMetadata returns only results whose metadata matches all filter key-value pairs.
func filterByMetadata(results []SearchResult, filter map[string]string) []SearchResult {
	filtered := make([]SearchResult, 0, len(results))
	for _, r := range results {
		if matchesMetadata(r.Metadata, filter) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// matchesMetadata returns true if metadata contains all filter key-value pairs.
func matchesMetadata(metadata, filter map[string]string) bool {
	for k, v := range filter {
		if metadata[k] != v {
			return false
		}
	}
	return true
}

// Delete removes vectors by collection and source IDs.
func (s *SQLiteVecStore) Delete(ctx context.Context, collection string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	placeholders := make([]string, len(ids))
	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, collection)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := fmt.Sprintf(
		`DELETE FROM vec_embeddings WHERE collection = ? AND source_id IN (%s)`,
		strings.Join(placeholders, ","),
	)

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete vec_embeddings: %w", err)
	}
	return nil
}

// Close is a no-op; the underlying sql.DB is managed externally.
func (s *SQLiteVecStore) Close() error {
	return nil
}
