# Spec: Generic Async Buffer

## Overview
Generic async buffer package (`internal/asyncbuf/`) providing two reusable buffer types that replace 5 duplicate implementations across the codebase.

## Requirements

### R1: BatchBuffer[T] — Batch-Oriented Async Processing
The system must provide a generic `BatchBuffer[T]` that:
- Accepts items via non-blocking `Enqueue(T)`
- Collects items into batches up to a configurable `BatchSize`
- Flushes batches on a configurable `BatchTimeout` timer
- Processes batches via a user-provided `ProcessBatchFunc[T]`
- Tracks dropped items when the queue is full (`DroppedCount()`)
- Drains remaining items on `Stop()` before returning
- Follows `Start(wg *sync.WaitGroup)` / `Stop()` lifecycle

#### Scenarios
- **Normal batch flush**: Items accumulate until `BatchSize` is reached, then flush.
- **Timeout flush**: Partial batch flushes after `BatchTimeout` with no new items.
- **Queue full**: `Enqueue` drops silently and increments drop counter.
- **Graceful shutdown**: `Stop()` processes remaining queued items before returning.

### R2: TriggerBuffer[T] — Per-Item Async Processing
The system must provide a generic `TriggerBuffer[T]` that:
- Accepts items via non-blocking `Enqueue(T)`
- Processes each item individually via `ProcessFunc[T]`
- Drains remaining items on `Stop()` before returning
- Follows `Start(wg *sync.WaitGroup)` / `Stop()` lifecycle

#### Scenarios
- **Normal processing**: Each enqueued item processed one-at-a-time.
- **Queue full**: `Enqueue` drops silently (non-blocking).
- **Graceful shutdown**: `Stop()` processes remaining queued items before returning.

### R3: Backward-Compatible Migration
All 5 existing buffers must be migrated to thin wrappers around asyncbuf types with zero public API changes:
- `embedding.EmbeddingBuffer` wraps `BatchBuffer[EmbedRequest]`
- `graph.GraphBuffer` wraps `BatchBuffer[GraphRequest]`
- `memory.Buffer` wraps `TriggerBuffer[string]`
- `learning.AnalysisBuffer` wraps `TriggerBuffer[AnalysisRequest]`
- `librarian.ProactiveBuffer` wraps `TriggerBuffer[string]`

## Dependencies
- `sync`, `sync/atomic`, `time` (stdlib)
- `go.uber.org/zap` (logging)
- No imports from application packages (leaf dependency)
