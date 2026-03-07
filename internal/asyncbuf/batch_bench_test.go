package asyncbuf

import (
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

func BenchmarkBatchEnqueue(b *testing.B) {
	logger := zap.NewNop().Sugar()
	buf := NewBatchBuffer[int](BatchConfig{
		QueueSize:    b.N + 1024,
		BatchSize:    64,
		BatchTimeout: time.Hour, // never fire by timeout during bench
	}, func(batch []int) {}, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Enqueue(i)
	}
	b.StopTimer()

	buf.Stop()
	wg.Wait()
}

func BenchmarkBatchEnqueueParallel(b *testing.B) {
	logger := zap.NewNop().Sugar()
	buf := NewBatchBuffer[int](BatchConfig{
		QueueSize:    1024 * 1024,
		BatchSize:    64,
		BatchTimeout: time.Hour,
	}, func(batch []int) {}, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			buf.Enqueue(i)
			i++
		}
	})
	b.StopTimer()

	buf.Stop()
	wg.Wait()
}

func BenchmarkBatchProcess(b *testing.B) {
	tests := []struct {
		name      string
		batchSize int
	}{
		{"BatchSize_8", 8},
		{"BatchSize_32", 32},
		{"BatchSize_128", 128},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			logger := zap.NewNop().Sugar()

			var processed int
			buf := NewBatchBuffer[int](BatchConfig{
				QueueSize:    b.N + 1024,
				BatchSize:    tt.batchSize,
				BatchTimeout: time.Hour,
			}, func(batch []int) {
				processed += len(batch)
			}, logger)

			var wg sync.WaitGroup
			buf.Start(&wg)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf.Enqueue(i)
			}
			b.StopTimer()

			buf.Stop()
			wg.Wait()
		})
	}
}

func BenchmarkTriggerEnqueue(b *testing.B) {
	logger := zap.NewNop().Sugar()
	buf := NewTriggerBuffer[int](TriggerConfig{
		QueueSize: b.N + 1024,
	}, func(item int) {}, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Enqueue(i)
	}
	b.StopTimer()

	buf.Stop()
	wg.Wait()
}
