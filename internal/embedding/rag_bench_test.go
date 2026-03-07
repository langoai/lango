package embedding

import (
	"fmt"
	"math/rand"
	"testing"
)

func makeRAGResults(n int) []RAGResult {
	results := make([]RAGResult, n)
	for i := range results {
		results[i] = RAGResult{
			Collection: "knowledge",
			SourceID:   fmt.Sprintf("doc_%d", i),
			Content:    fmt.Sprintf("Content for document %d with some text.", i),
			Distance:   rand.Float32(), //nolint:gosec
		}
	}
	return results
}

func BenchmarkSortByDistance(b *testing.B) {
	tests := []struct {
		name string
		size int
	}{
		{"Results_10", 10},
		{"Results_50", 50},
		{"Results_200", 200},
	}

	for _, tt := range tests {
		original := makeRAGResults(tt.size)

		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			buf := make([]RAGResult, tt.size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				copy(buf, original)
				sortByDistance(buf)
			}
		})
	}
}

func BenchmarkFilterByMaxDistance(b *testing.B) {
	tests := []struct {
		name    string
		size    int
		maxDist float32
	}{
		{"Results_50_Dist_0.5", 50, 0.5},
		{"Results_50_Dist_0.1", 50, 0.1},
		{"Results_200_Dist_0.5", 200, 0.5},
	}

	for _, tt := range tests {
		results := makeRAGResults(tt.size)

		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				filterByMaxDistance(results, tt.maxDist)
			}
		})
	}
}

func BenchmarkEmbeddingCacheGet(b *testing.B) {
	cache := newEmbeddingCache(0, 1000) // TTL=0 means entries never expire for bench
	vec := make([]float32, 768)
	for i := range vec {
		vec[i] = rand.Float32() //nolint:gosec
	}

	// Pre-populate cache.
	for i := 0; i < 100; i++ {
		cache.set(fmt.Sprintf("query_%d", i), vec)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.get(fmt.Sprintf("query_%d", i%100))
	}
}

func BenchmarkEmbeddingCacheSet(b *testing.B) {
	vec := make([]float32, 768)
	for i := range vec {
		vec[i] = rand.Float32() //nolint:gosec
	}

	b.Run("UnderCapacity", func(b *testing.B) {
		cache := newEmbeddingCache(0, b.N+100)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.set(fmt.Sprintf("query_%d", i), vec)
		}
	})

	b.Run("AtCapacity", func(b *testing.B) {
		cache := newEmbeddingCache(0, 50)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.set(fmt.Sprintf("query_%d", i), vec)
		}
	})
}
