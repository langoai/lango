## Why

The `App` struct has 22+ fields (God Object), `New()` is a 440-line sequential initializer (God Function), 13 `SetXxxCallback()` calls create ad-hoc wiring, tool wrapping has order-dependent 3-layer nesting, and bootstrap has no rollback on failure. These structural problems make adding new components error-prone and increase maintenance cost.

## What Changes

- Introduce `internal/lifecycle/` — Component interface with Registry for ordered startup, reverse shutdown, and rollback on failure
- Introduce `internal/toolchain/` — HTTP-style middleware chain for tool wrapping (learning, approval, browser recovery)
- Introduce `internal/bootstrap/pipeline.go` — Phase+Pipeline pattern with cleanup stack for bootstrap sequence
- Introduce `internal/appinit/` — Module interface with Builder and topological sort for declarative initialization
- Introduce `internal/eventbus/` — Typed synchronous event bus to replace 13+ SetXxxCallback() calls
- Refactor `App.Start()`/`Stop()` to delegate to lifecycle registry
- Refactor `bootstrap.Run()` from 230-line function to 3-line pipeline invocation
- Refactor tool wrapping from 3 nested loops to `ChainAll()` one-liner

## Capabilities

### New Capabilities
- `lifecycle-registry`: Component lifecycle management with priority-ordered start, reverse stop, and failure rollback
- `tool-middleware`: Composable middleware chain for cross-cutting tool concerns (learning, approval, browser recovery)
- `bootstrap-pipeline`: Phase-based bootstrap with sequential execution and reverse-order cleanup on failure
- `appinit-modules`: Module interface with topological sort for declarative app initialization
- `event-bus`: Typed synchronous publish/subscribe bus for decoupled component communication

### Modified Capabilities

## Impact

- `internal/app/app.go` — `New()`, `Start()`, `Stop()` refactored to use lifecycle registry and toolchain
- `internal/app/tools.go` — Wrapping functions delegate to toolchain package
- `internal/app/types.go` — Added `registry *lifecycle.Registry` field
- `internal/bootstrap/bootstrap.go` — `Run()` replaced with pipeline invocation
- No breaking changes to external APIs or CLI
- No changes to existing tests (all pass)
