package search

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupBenchDB(b *testing.B, count int) (*sql.DB, *FTS5Index) {
	b.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = db.Close() })

	if !ProbeFTS5(db) {
		b.Skip("FTS5 not available")
	}

	idx := NewFTS5Index(db, "bench_fts", []string{"key", "content"})
	if err := idx.EnsureTable(); err != nil {
		b.Fatal(err)
	}

	// Also create a regular table for LIKE comparison.
	if _, err := db.Exec(`CREATE TABLE bench_like (key TEXT, content TEXT)`); err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	var ftsRecords []Record
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("key-%d", i)
		content := fmt.Sprintf("entry %d about various topics including deployment configuration server setup database management", i)
		ftsRecords = append(ftsRecords, Record{RowID: key, Values: []string{key, content}})
	}

	if err := idx.BulkInsert(ctx, ftsRecords); err != nil {
		b.Fatal(err)
	}

	// Populate LIKE table.
	tx, err := db.Begin()
	if err != nil {
		b.Fatal(err)
	}
	stmt, err := tx.Prepare(`INSERT INTO bench_like(key, content) VALUES(?, ?)`)
	if err != nil {
		b.Fatal(err)
	}
	for _, r := range ftsRecords {
		if _, err := stmt.Exec(r.Values[0], r.Values[1]); err != nil {
			b.Fatal(err)
		}
	}
	stmt.Close()
	if err := tx.Commit(); err != nil {
		b.Fatal(err)
	}

	return db, idx
}

func BenchmarkFTS5Search_1k(b *testing.B) {
	_, idx := setupBenchDB(b, 1000)
	ctx := context.Background()
	b.ResetTimer()
	for b.Loop() {
		_, _ = idx.Search(ctx, "deployment configuration", 10)
	}
}

func BenchmarkLIKESearch_1k(b *testing.B) {
	db, _ := setupBenchDB(b, 1000)
	b.ResetTimer()
	for b.Loop() {
		rows, _ := db.Query(`SELECT key FROM bench_like WHERE content LIKE '%deployment%' OR content LIKE '%configuration%' ORDER BY key LIMIT 10`)
		if rows != nil {
			rows.Close()
		}
	}
}

func BenchmarkFTS5Search_10k(b *testing.B) {
	_, idx := setupBenchDB(b, 10000)
	ctx := context.Background()
	b.ResetTimer()
	for b.Loop() {
		_, _ = idx.Search(ctx, "deployment configuration", 10)
	}
}

func BenchmarkLIKESearch_10k(b *testing.B) {
	db, _ := setupBenchDB(b, 10000)
	b.ResetTimer()
	for b.Loop() {
		rows, _ := db.Query(`SELECT key FROM bench_like WHERE content LIKE '%deployment%' OR content LIKE '%configuration%' ORDER BY key LIMIT 10`)
		if rows != nil {
			rows.Close()
		}
	}
}
