# Proposal: Add Typed Event Bus Package

## Problem

There are 13+ `SetXxxCallback()` calls scattered through wiring code (`internal/app/wiring.go`).
These callbacks are:
1. Synchronously invoked
2. Optional (nil-checked before invocation)
3. Set via setter methods after construction

This creates tight coupling between producers and consumers, making it hard to
add new subscribers or change wiring without touching multiple files.

## Solution

Create a new `internal/eventbus/` package that provides a synchronous, typed
event bus. The bus acts as a central dispatcher that decouples event producers
from consumers.

**Phase 4 scope:** Create the package only. No existing code is modified.
Future phases will incrementally migrate callbacks to publish/subscribe.

## Benefits

- Decouples producers from consumers (no more SetXxxCallback)
- Type-safe subscriptions via generics
- Thread-safe by design (RWMutex)
- Zero external dependencies
- Dependency-free mirror types avoid import cycles
