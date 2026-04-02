package search

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// SearchResult holds a single FTS5 search result.
type SearchResult struct {
	RowID string
	Rank  float64
}

// Record holds data for a single FTS5 record to insert.
type Record struct {
	RowID  string
	Values []string
}

// FTS5Index manages an FTS5 virtual table for full-text search.
// It is domain-agnostic — it operates on raw table names, column names,
// rowids, and string values. Domain semantics (e.g., which entries to
// index) are the caller's responsibility.
type FTS5Index struct {
	db        *sql.DB
	tableName string
	columns   []string
}

// NewFTS5Index creates a new FTS5Index. Call EnsureTable() before use.
func NewFTS5Index(db *sql.DB, tableName string, columns []string) *FTS5Index {
	return &FTS5Index{
		db:        db,
		tableName: tableName,
		columns:   columns,
	}
}

// EnsureTable creates the FTS5 virtual table if it does not exist.
// The table includes an UNINDEXED source_id column for row identification
// that is not included in FTS5 text search.
func (idx *FTS5Index) EnsureTable() error {
	cols := strings.Join(idx.columns, ", ")
	query := fmt.Sprintf(
		`CREATE VIRTUAL TABLE IF NOT EXISTS %s USING fts5(%s, source_id UNINDEXED, tokenize='unicode61')`,
		idx.tableName, cols,
	)
	_, err := idx.db.Exec(query)
	if err != nil {
		return fmt.Errorf("create FTS5 table %s: %w", idx.tableName, err)
	}
	return nil
}

// DropTable drops the FTS5 virtual table if it exists.
func (idx *FTS5Index) DropTable() error {
	query := fmt.Sprintf(`DROP TABLE IF EXISTS %s`, idx.tableName)
	_, err := idx.db.Exec(query)
	if err != nil {
		return fmt.Errorf("drop FTS5 table %s: %w", idx.tableName, err)
	}
	return nil
}

// Insert adds a new record to the FTS5 index.
// The rowid is a string identifier (typically the source entity's key).
// values must match the column order from NewFTS5Index.
func (idx *FTS5Index) Insert(ctx context.Context, rowid string, values []string) error {
	placeholders := make([]string, 0, len(idx.columns)+1)
	args := make([]any, 0, len(idx.columns)+1)

	placeholders = append(placeholders, "?")
	args = append(args, rowid)
	for _, v := range values {
		placeholders = append(placeholders, "?")
		args = append(args, v)
	}

	query := fmt.Sprintf(
		`INSERT INTO %s(source_id, %s) VALUES(%s)`,
		idx.tableName,
		strings.Join(idx.columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := idx.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("FTS5 insert into %s: %w", idx.tableName, err)
	}
	return nil
}

// Update replaces an existing record by deleting and re-inserting.
func (idx *FTS5Index) Update(ctx context.Context, rowid string, values []string) error {
	if err := idx.Delete(ctx, rowid); err != nil {
		return err
	}
	return idx.Insert(ctx, rowid, values)
}

// Delete removes a record by rowid.
func (idx *FTS5Index) Delete(ctx context.Context, rowid string) error {
	query := fmt.Sprintf(
		`DELETE FROM %s WHERE source_id = ?`,
		idx.tableName,
	)
	_, err := idx.db.ExecContext(ctx, query, rowid)
	if err != nil {
		return fmt.Errorf("FTS5 delete from %s: %w", idx.tableName, err)
	}
	return nil
}

// BulkInsert inserts multiple records in a single transaction.
func (idx *FTS5Index) BulkInsert(ctx context.Context, records []Record) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := idx.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("FTS5 bulk insert begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	colList := strings.Join(idx.columns, ", ")
	placeholders := make([]string, 0, len(idx.columns)+1)
	for range idx.columns {
		placeholders = append(placeholders, "?")
	}
	placeholders = append([]string{"?"}, placeholders...) // source_id + columns

	query := fmt.Sprintf(
		`INSERT INTO %s(source_id, %s) VALUES(%s)`,
		idx.tableName, colList, strings.Join(placeholders, ", "),
	)

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("FTS5 bulk insert prepare: %w", err)
	}
	defer stmt.Close()

	for _, r := range records {
		args := make([]any, 0, len(r.Values)+1)
		args = append(args, r.RowID)
		for _, v := range r.Values {
			args = append(args, v)
		}
		if _, err := stmt.ExecContext(ctx, args...); err != nil {
			return fmt.Errorf("FTS5 bulk insert row %s: %w", r.RowID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("FTS5 bulk insert commit: %w", err)
	}
	return nil
}

// Search executes an FTS5 MATCH query and returns results ranked by BM25.
// Supports plain keywords, phrase queries (quoted), and prefix queries (trailing *).
// Returns an empty slice for empty queries.
func (idx *FTS5Index) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 10
	}

	matchExpr := sanitizeFTS5Query(query)
	if matchExpr == "" {
		return nil, nil
	}

	sqlQuery := fmt.Sprintf(
		`SELECT source_id, rank FROM %s WHERE %s MATCH ? ORDER BY rank LIMIT ?`,
		idx.tableName, idx.tableName,
	)

	rows, err := idx.db.QueryContext(ctx, sqlQuery, matchExpr, limit)
	if err != nil {
		return nil, fmt.Errorf("FTS5 search %s: %w", idx.tableName, err)
	}
	defer rows.Close()

	results := make([]SearchResult, 0, limit)
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.RowID, &r.Rank); err != nil {
			return nil, fmt.Errorf("FTS5 search scan: %w", err)
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("FTS5 search rows: %w", err)
	}
	return results, nil
}

// sanitizeFTS5Query prepares a query string for FTS5 MATCH.
// It preserves phrases (quoted strings), prefix tokens (trailing *),
// and converts plain keywords into OR-combined terms.
func sanitizeFTS5Query(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	var terms []string
	i := 0
	for i < len(query) {
		if query[i] == '"' {
			// Find closing quote for phrase query.
			end := strings.IndexByte(query[i+1:], '"')
			if end >= 0 {
				phrase := query[i : i+end+2]
				terms = append(terms, phrase)
				i += end + 2
			} else {
				// Unclosed quote — treat rest as plain text.
				word := strings.TrimSpace(query[i+1:])
				if word != "" {
					escaped := escapeFTS5Token(word)
					if escaped != "" {
						terms = append(terms, escaped)
					}
				}
				break
			}
		} else if query[i] == ' ' || query[i] == '\t' {
			i++
		} else {
			// Read a token until space or end.
			end := strings.IndexAny(query[i:], " \t\"")
			var token string
			if end < 0 {
				token = query[i:]
				i = len(query)
			} else {
				token = query[i : i+end]
				i += end
			}
			token = strings.TrimSpace(token)
			if token == "" {
				continue
			}
			// Preserve prefix queries (trailing *).
			if base, ok := strings.CutSuffix(token, "*"); ok {
				escaped := escapeFTS5Token(base)
				if escaped != "" {
					terms = append(terms, escaped+"*")
				}
			} else {
				escaped := escapeFTS5Token(token)
				if escaped != "" {
					terms = append(terms, escaped)
				}
			}
		}
	}

	if len(terms) == 0 {
		return ""
	}
	return strings.Join(terms, " OR ")
}

// escapeFTS5Token removes characters that are special or problematic in FTS5 queries.
func escapeFTS5Token(s string) string {
	replacer := strings.NewReplacer(
		"^", "", "-", "", "+", "",
		"(", "", ")", "",
		"{", "", "}", "",
		":", "", ".", "", "?", "",
		"!", "", "@", "", "#", "",
		"$", "", "%", "", "&", "",
		"=", "", "|", "", "~", "",
		"<", "", ">", "", ";", "",
		",", "", "[", "", "]", "",
		"\\", "", "/", "", "'", "",
	)
	return replacer.Replace(s)
}
