package graph

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

func newBenchStore(b *testing.B) *BoltStore {
	b.Helper()
	dbPath := filepath.Join(b.TempDir(), "bench.db")
	store, err := NewBoltStore(dbPath)
	if err != nil {
		b.Fatalf("open bolt store: %v", err)
	}
	b.Cleanup(func() { store.Close() })
	return store
}

func seedTriples(b *testing.B, store *BoltStore, count int) {
	b.Helper()
	ctx := context.Background()
	triples := make([]Triple, count)
	for i := range triples {
		triples[i] = Triple{
			Subject:   fmt.Sprintf("entity_%d", i%100),
			Predicate: RelatedTo,
			Object:    fmt.Sprintf("entity_%d", (i+1)%100),
			Metadata:  map[string]string{"source": "bench"},
		}
	}
	if err := store.AddTriples(ctx, triples); err != nil {
		b.Fatalf("seed triples: %v", err)
	}
}

func BenchmarkAddTriple(b *testing.B) {
	store := newBenchStore(b)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.AddTriple(ctx, Triple{
			Subject:   fmt.Sprintf("sub_%d", i),
			Predicate: RelatedTo,
			Object:    fmt.Sprintf("obj_%d", i),
			Metadata:  map[string]string{"i": fmt.Sprintf("%d", i)},
		})
	}
}

func BenchmarkAddTriples(b *testing.B) {
	tests := []struct {
		name      string
		batchSize int
	}{
		{"Batch_10", 10},
		{"Batch_50", 50},
		{"Batch_100", 100},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			store := newBenchStore(b)
			ctx := context.Background()

			batch := make([]Triple, tt.batchSize)
			for i := range batch {
				batch[i] = Triple{
					Subject:   fmt.Sprintf("sub_%d", i),
					Predicate: RelatedTo,
					Object:    fmt.Sprintf("obj_%d", i),
				}
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = store.AddTriples(ctx, batch)
			}
		})
	}
}

func BenchmarkQueryBySubject(b *testing.B) {
	tests := []struct {
		name      string
		seedCount int
	}{
		{"Store_100", 100},
		{"Store_1000", 1000},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			store := newBenchStore(b)
			seedTriples(b, store, tt.seedCount)
			ctx := context.Background()

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = store.QueryBySubject(ctx, "entity_0")
			}
		})
	}
}

func BenchmarkQueryByObject(b *testing.B) {
	store := newBenchStore(b)
	seedTriples(b, store, 500)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.QueryByObject(ctx, "entity_0")
	}
}

func BenchmarkTraverse(b *testing.B) {
	tests := []struct {
		name     string
		maxDepth int
	}{
		{"Depth_1", 1},
		{"Depth_2", 2},
		{"Depth_3", 3},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			store := newBenchStore(b)
			seedTriples(b, store, 500)
			ctx := context.Background()

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = store.Traverse(ctx, "entity_0", tt.maxDepth, nil)
			}
		})
	}
}

func BenchmarkTraverseWithPredicateFilter(b *testing.B) {
	store := newBenchStore(b)
	seedTriples(b, store, 500)
	ctx := context.Background()
	predicates := []string{RelatedTo}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.Traverse(ctx, "entity_0", 2, predicates)
	}
}
