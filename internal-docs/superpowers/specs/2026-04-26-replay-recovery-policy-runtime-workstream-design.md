# Replay / Recovery Policy Runtime Workstream Design

## Purpose / Scope

This workstream consolidates the replay and recovery runtime layer beneath the already-landed operator surfaces so the project can move into dispute-runtime completion and deeper runtime integration without carrying policy and retry-substrate drift.

The target runtime surfaces are:

- post-adjudication execution mode defaults
- background retry / dead-letter policy behavior
- replay policy and recovery policy alignment

This workstream directly includes:

- policy-driven defaults for follow-up execution and replay behavior
- generic async retry policy normalization
- replay / recovery substrate normalization

This workstream does not directly include:

- keep-hold or re-escalation states
- dispute engine completion
- settlement progression redesign
- richer policy editing surfaces
- reputation model changes

The goal is to stabilize the replay / recovery runtime substrate before the next larger dispute-runtime workstream.

## Current Baseline

The landed runtime already includes:

- optional `auto_execute=true` on adjudication
- optional `background_execute=true` on adjudication
- bounded retry / dead-letter handling for post-adjudication background execution
- operator replay / manual retry
- first policy-driven replay controls
- rich dead-letter operator surfaces in CLI and cockpit

The remaining runtime gaps are now concentrated:

- defaults for when inline execution, background execution, or replay should be policy-selected rather than surface-selected
- retry scheduling / dead-letter policy behavior that is still too post-adjudication-specific
- replay substrate semantics that still sit too separately from the background retry substrate

This makes a dedicated runtime workstream feasible without mixing in dispute state-machine changes.

## Workstream Scope

This workstream is intentionally runtime-focused and excludes broader dispute branching.

### 1. Policy-Driven Defaults

Clarify and normalize how the runtime chooses among:

- inline execution
- background execution
- manual replay / operator-triggered recovery

The first implementation should:

- preserve current explicit surface flags and controls
- define runtime defaults when those controls are absent
- keep the adjudication path as the canonical write layer

This workstream should not redesign operator UX first. It should make the runtime default behavior more coherent underneath the existing surfaces.

### 2. Generic Async Retry Policy

Normalize the current retry / dead-letter behavior into a clearer runtime policy unit.

The first implementation should:

- preserve the currently landed retry scheduling and dead-letter semantics
- make retry limits, retry scheduling, and dead-letter transition rules easier to reason about as runtime policy
- reduce the degree to which those rules are implicitly tied to only one code path

This does not require building a fully generic background-manager-wide retry subsystem in one step, but it should move the current implementation closer to that shape.

### 3. Replay / Recovery Substrate Normalization

Clarify and normalize how replay interacts with the retry / dead-letter substrate.

The first implementation should:

- preserve the canonical adjudication + dead-letter evidence gate
- preserve replay policy enforcement
- align replay-request semantics with the runtime recovery substrate more explicitly
- reduce duplication between replay-specific and retry-specific policy handling where practical

This remains substrate work, not a new operator feature set.

## Architectural Shape

This workstream should preserve the current separation between:

- canonical adjudication and receipts
- background dispatch / retry handling
- replay policy enforcement
- operator surfaces

### Runtime Boundary

Primary ownership is expected to fall around:

- `internal/postadjudicationreplay/*`
- `internal/background/*`
- `internal/app/tools_meta*.go`

The runtime boundary should own:

- execution-mode default policy
- retry / dead-letter policy normalization
- replay / recovery substrate alignment

### Receipts Boundary

Potential secondary ownership may touch:

- `internal/receipts/*`

Only where needed to:

- preserve canonical receipt evidence contracts
- normalize state or evidence behavior relied on by runtime policy

This workstream should not move broad dispute branching into receipts changes.

### Operator Boundary

Operator surfaces may need limited alignment work in:

- `internal/cli/status/*`
- `internal/cli/cockpit/*`
- docs / OpenSpec

But these are downstream alignment concerns. The primary change belongs in the runtime layer.

## Execution / Parallelization Model

This workstream should be treated as a larger batch with narrower parallelism than operator-surface work.

Execution model:

- one spec
- one implementation plan
- three workers in parallel
- implementation remains semi-serial around the runtime core

### Worker A

Owns:

- runtime policy / retry core
- replay / recovery substrate logic
- primary tests for runtime behavior

### Worker B

Owns:

- downstream CLI / cockpit / meta-tool impact
- runtime-facing integration tests

### Worker C

Owns:

- `docs/architecture/*`
- `docs/cli/*`
- `README.md` when needed
- `openspec/*`

Worker A establishes the core contract first. Worker B should align with that contract rather than invent a second runtime shape.

## Implementation Strategy

This workstream should be implemented as one plan with grouped task bands rather than more micro-workstreams.

Recommended task bands:

1. map and normalize the current retry / replay runtime helpers and contracts
2. land policy-driven defaults in the canonical runtime path
3. normalize async retry / dead-letter policy behavior
4. align replay / recovery substrate behavior
5. truth-align downstream docs / OpenSpec / operator wording
6. final integrated verification

The sequence is important because the core runtime contract must settle before downstream alignment work.

## Completion Criteria

The workstream is complete when:

- policy-driven defaults for follow-up execution and replay behavior are coherent and landed in code
- retry / dead-letter behavior is more clearly normalized as runtime policy
- replay and recovery substrate semantics are more explicitly aligned
- downstream docs and OpenSpec reflect the landed runtime behavior
- `go build ./...`
- `go test ./...`
- `.venv/bin/zensical build`
- docs-only OpenSpec validation
all pass

## Follow-On Inputs

Natural follow-on work after this workstream:

1. Dispute Runtime Completion Workstream
- keep-hold / re-escalation
- broader dispute engine completion
- richer settlement progression
- escrow lifecycle completion

2. Reputation V2 + Runtime Integration Workstream
- reputation model v2
- trust-entry contract strengthening
- deeper runtime integration
