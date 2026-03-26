package search

import "database/sql"

// ProbeFTS5 tests whether the SQLite connection supports FTS5.
// It creates a temporary FTS5 table and immediately drops it.
// Returns true if FTS5 is available, false otherwise.
func ProbeFTS5(db *sql.DB) bool {
	const probe = `CREATE VIRTUAL TABLE IF NOT EXISTS _fts5_probe USING fts5(x)`
	_, err := db.Exec(probe)
	if err != nil {
		return false
	}
	_, _ = db.Exec(`DROP TABLE IF EXISTS _fts5_probe`)
	return true
}
