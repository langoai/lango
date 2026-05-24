# Execution Hardening Workstream Design

## Purpose / Scope

This workstream hardens the recently expanded dead-letter, recovery, dispute, and runtime surfaces by closing a small set of high-risk execution and wiring gaps.

The target surfaces are:

- dead-letter cockpit shell wiring
- replay invocation safety
- background task execution safety
- reputation update safety
- dispatch-family classifier consistency

This workstream directly includes:

- per-transaction or per-identity synchronization where missing
- replay principal injection hardening
- background panic recovery hardening
- partial-commit failure coverage
- replay dispatch dedup policy or at least explicit dedup verification
- small classifier / wiring consistency fixes

This workstream does not directly include:

- new product-surface expansion
- new dispute state-machine semantics
- new trust / reputation product features
- broader runtime redesign

The goal is to remove the highest-risk operational holes without reopening the broader roadmap.

## Current Baseline

Large domain and runtime workstreams are already landed:

- operator surface consolidation
- replay / recovery runtime normalization
- dispute runtime completion
- reputation v2 + runtime integration

What remains is mostly systemic hardening:

- cockpit shell wiring gaps can silently distort dead-letter filters and summaries
- replay invocation still depends on correct principal propagation
- background manager panic handling needed explicit hardening
- reputation persistence still needed serialization hardening
- family classifiers and adapters can drift across surfaces

This makes a narrow stabilization workstream appropriate.

## Workstream Scope

This workstream is intentionally narrow and tactical.

### 1. Cockpit / CLI Wiring Safety

Close wiring gaps where data or principal context is silently lost.

The first implementation should:

- ensure dead-letter filter fields are forwarded through the cockpit shell adapter
- ensure retry invocations do not reach replay policy with an empty principal when a stable local fallback is appropriate

### 2. Background Execution Hardening

Clarify and harden how the background manager behaves under runner failures and panics.

The first implementation should:

- recover from panics in background execution
- ensure tasks do not become orphaned in running state
- preserve explicit task failure visibility

### 3. Reputation Persistence Hardening

Reduce the chance of lost updates and invalid score propagation.

The first implementation should:

- serialize per-peer reputation updates
- prevent `NaN` score propagation through trust-entry logic

### 4. Consistency and Coverage

Close small but important drift points.

The first implementation should:

- unify dispatch-family classification between CLI and cockpit
- add targeted tests for partial-commit failure semantics and replay/recovery safety
- document any intentionally chosen panic or retry policy behavior when it is not obvious from the code

## Architectural Shape

This workstream should preserve the current boundaries.

### Shell / Operator Wiring Boundary

Primary ownership is expected around:

- `cmd/lango/main.go`
- `internal/cli/status/*`
- `internal/cli/cockpit/*`

This boundary should own:

- filter forwarding correctness
- principal injection correctness
- UI / CLI contract consistency

### Background Runtime Boundary

Primary ownership is expected around:

- `internal/background/*`

This boundary should own:

- panic recovery semantics
- task lifecycle integrity

### Reputation Boundary

Primary ownership is expected around:

- `internal/p2p/reputation/*`

This boundary should own:

- per-peer serialization
- score clamping safety

## Execution / Parallelization Model

This workstream should be handled as a smaller but still structured batch.

Execution model:

- one spec
- one implementation plan
- three workers in parallel

### Worker A

Owns:

- background execution hardening
- reputation persistence hardening
- focused core tests

### Worker B

Owns:

- cockpit / CLI wiring and retry principal hardening
- classifier consistency
- focused surface tests

### Worker C

Owns:

- docs / OpenSpec / README

This split keeps concurrency and operator wiring concerns separate while still letting them land together.

## Implementation Strategy

Recommended task bands:

1. identify and test the wiring / execution gaps directly
2. land background and reputation hardening
3. land cockpit / CLI wiring fixes
4. add or extend targeted failure-path tests
5. truth-align docs / OpenSpec
6. final integrated verification

The workstream should stay surgical. Any refactor that is not needed to close a listed gap is out of scope.

## Completion Criteria

The workstream is complete when:

- cockpit dead-letter filters are forwarded correctly
- replay invocation no longer depends on an absent principal where a stable local fallback is appropriate
- background task panics do not orphan task state
- reputation updates are safer under concurrency
- dispatch-family classification is consistent across CLI and cockpit
- targeted failure-path tests exist for the stabilized behavior
- `go build ./...`
- `go test ./...`
- `.venv/bin/zensical build`
- docs-only OpenSpec validation
all pass

## Follow-On Inputs

Natural follow-on work after this workstream:

1. longer-tail hardening and cleanup
- remaining minor consistency fixes

2. future feature work
- only after the newly landed larger workstreams are considered operationally stable
