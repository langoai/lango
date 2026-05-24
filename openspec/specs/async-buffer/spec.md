# Spec: Generic Async Buffer

## Overview
Generic async buffer package (`internal/asyncbuf/`) providing two reusable buffer types that replace 5 duplicate implementations across the codebase.

## Purpose

Capability spec for async-buffer. See requirements below for scope and behavior contracts.

## Requirements

### R1: BatchBuffer[T] — Batch-Oriented Async Processing
The system SHALL provide a generic `BatchBuffer[T]` that:
- Accepts items via non-blocking `Enqueue(T)`
- Collects items into batches up to a configurable `BatchSize`
- Flushes batches on a configurable `BatchTimeout` timer
- Processes batches via a user-provided `ProcessBatchFunc[T]`
- Tracks dropped items when the queue is full (`DroppedCount()`)
- Drains remaining items on `Stop()` before returning
- Follows `Start(wg *sync.WaitGroup)` / `Stop()` lifecycle

#### Scenario: Normal batch flush
- **WHEN** items accumulate until `BatchSize` is reached
- **THEN** the buffer SHALL flush the batch immediately

#### Scenario: Timeout flush
- **WHEN** a partial batch sits idle until `BatchTimeout` expires
- **THEN** the buffer SHALL flush the partial batch

#### Scenario: Queue full
- **WHEN** `Enqueue` is called while the queue is full
- **THEN** the item SHALL be dropped
- **AND** the dropped counter SHALL increment

#### Scenario: Graceful shutdown
- **WHEN** `Stop()` is called with queued items remaining
- **THEN** the buffer SHALL process the remaining queued items before returning

### R2: TriggerBuffer[T] — Per-Item Async Processing
The system SHALL provide a generic `TriggerBuffer[T]` that:
- Accepts items via non-blocking `Enqueue(T)`
- Processes each item individually via `ProcessFunc[T]`
- Drains remaining items on `Stop()` before returning
- Follows `Start(wg *sync.WaitGroup)` / `Stop()` lifecycle

#### Scenario: Normal processing
- **WHEN** items are enqueued into the trigger buffer
- **THEN** each item SHALL be processed one at a time

#### Scenario: Queue full
- **WHEN** `Enqueue` is called while the queue is full
- **THEN** the item SHALL be dropped without blocking

#### Scenario: Graceful shutdown
- **WHEN** `Stop()` is called with queued items remaining
- **THEN** the trigger buffer SHALL process the remaining queued items before returning

### R3: Backward-Compatible Migration
All 5 existing buffers SHALL be migrated to thin wrappers around asyncbuf types with zero public API changes:
- `embedding.EmbeddingBuffer` wraps `BatchBuffer[EmbedRequest]`
- `graph.GraphBuffer` wraps `BatchBuffer[GraphRequest]`
- `memory.Buffer` wraps `TriggerBuffer[string]`
- `learning.AnalysisBuffer` wraps `TriggerBuffer[AnalysisRequest]`
- `librarian.ProactiveBuffer` wraps `TriggerBuffer[string]`

#### Scenario: Legacy wrappers preserve public APIs
- **WHEN** callers use the existing embedding, graph, memory, learning, or librarian buffer types
- **THEN** those wrappers SHALL preserve the previous public API shape while delegating to asyncbuf implementations

## Dependencies
- `sync`, `sync/atomic`, `time` (stdlib)
- `go.uber.org/zap` (logging)
- No imports from application packages (leaf dependency)
