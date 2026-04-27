# Dispute Runtime Completion Workstream Design

## Purpose / Scope

This workstream completes the next major layer of dispute / settlement / escrow runtime behavior so the project can move toward reputation and deeper runtime integration without carrying large state-machine gaps in the dispute core.

The target runtime surfaces are:

- dispute progression after hold and adjudication
- settlement progression after disagreement and partial settlement
- escrow lifecycle behavior under dispute-linked outcomes

This workstream directly includes:

- keep-hold / re-escalation semantics
- broader dispute engine completion
- richer settlement progression
- escrow lifecycle completion

This workstream does not directly include:

- replay / recovery policy redesign
- operator analytics or operator UI expansion
- reputation model changes
- broader runtime integration outside dispute / settlement / escrow

The goal is to complete the canonical dispute / settlement / escrow state machine far enough that later reputation and runtime-integration work can assume a much more stable domain core.

## Current Baseline

The landed domain core already includes:

- dispute hold
- release-vs-refund adjudication
- adjudication-aware release/refund execution gating
- direct settlement execution
- partial settlement execution first slice
- escrow release first slice
- escrow refund first slice
- post-adjudication execution and recovery first slices

The remaining gaps are now concentrated:

- no canonical keep-hold or re-escalation path after adjudication pressure or incomplete resolution
- settlement progression still lacks deeper disagreement and multi-round completion semantics
- escrow lifecycle still lacks broader dispute-linked completion behavior
- dispute engine behavior remains a collection of first slices rather than a more coherent runtime

This makes a dedicated dispute-runtime completion workstream feasible and timely.

## Workstream Scope

This workstream is intentionally core-domain-focused.

### 1. Keep-Hold / Re-Escalation

Clarify whether the runtime can:

- keep an escrow or dispute in held state beyond the first adjudication branching pass
- re-escalate after a failed or incomplete downstream resolution

The first implementation should:

- define explicit canonical states or transitions for continued hold and re-escalation
- preserve existing adjudication evidence and current canonical receipts
- avoid hiding disagreement continuation behind ad hoc failure behavior

### 2. Broader Dispute Engine Completion

Make dispute handling more coherent as a domain engine rather than a chain of isolated tool behaviors.

The first implementation should:

- align dispute hold, adjudication, execution gating, and downstream resolution semantics
- make the canonical disagreement path easier to reason about in receipts and services
- preserve already-landed release/refund branching where it remains correct

### 3. Richer Settlement Progression

Extend the current settlement progression model beyond the landed first path.

The first implementation should:

- support deeper disagreement aftermath semantics
- clarify how partial settlement, continued review, or renewed dispute pressure fit into progression
- keep transaction-level progression canonical

### 4. Escrow Lifecycle Completion

Complete the next important escrow behaviors tied to dispute outcomes.

The first implementation should:

- make release/refund safety rules more explicit
- cover remaining dispute-linked escrow transitions that are still implicit or incomplete
- move closer to a stable escrow lifecycle under dispute pressure

This should still avoid a full milestone-engine redesign unless the current codebase already makes that a natural narrow extension.

## Architectural Shape

This workstream should preserve the current layered ownership.

### Receipts Boundary

Primary ownership is expected to fall around:

- `internal/receipts/*`

This boundary should continue to own:

- canonical transaction / submission state
- settlement progression state
- dispute and escrow evidence trails

This workstream will likely rely on receipts as the canonical dispute-core authority.

### Domain Service Boundary

Primary ownership is expected to fall around:

- `internal/disputehold/*`
- `internal/escrowadjudication/*`
- `internal/escrowrelease/*`
- `internal/escrowrefund/*`
- `internal/settlementprogression/*`
- `internal/settlementexecution/*`
- `internal/partialsettlementexecution/*`

This boundary should own:

- domain transition rules
- service-level validation and invariants
- execution result mapping back into receipts

### Meta-Tool Boundary

Secondary ownership may touch:

- `internal/app/tools_meta*.go`

Only where needed to keep the public tool contracts aligned with the landed canonical dispute runtime.

### Operator Boundary

Operator surfaces may require only minimal downstream wording alignment if new canonical states become visible.

This workstream should not turn into an operator-surface redesign.

## Execution / Parallelization Model

This workstream should be treated as a larger batch with constrained parallelism.

Execution model:

- one spec
- one implementation plan
- three workers in parallel
- implementation remains semi-serial around receipts and domain-core contracts

### Worker A

Owns:

- receipts and dispute-core canonical state
- progression and evidence invariants
- primary domain tests

### Worker B

Owns:

- settlement / escrow service alignment
- meta-tool and downstream integration tests

### Worker C

Owns:

- `docs/architecture/*`
- `docs/cli/*` only if runtime-facing behavior changes
- `README.md` when needed
- `openspec/*`

Worker A establishes the canonical dispute-core contract first. Worker B should align settlement / escrow service behavior with that contract rather than competing with it.

## Implementation Strategy

This workstream should be implemented in one plan with grouped task bands rather than more micro-workstreams.

Recommended task bands:

1. map current dispute / settlement / escrow invariants and missing transitions
2. land canonical keep-hold / re-escalation semantics in receipts and core services
3. extend settlement progression for richer disagreement outcomes
4. complete dispute-linked escrow lifecycle behavior
5. align meta tools and downstream contract wording
6. truth-align docs / OpenSpec
7. final integrated verification

The sequence matters because the canonical state contract must settle before execution paths and tools are aligned to it.

## Completion Criteria

The workstream is complete when:

- keep-hold / re-escalation behavior is explicitly modeled and landed
- settlement progression handles richer disagreement outcomes more coherently
- dispute-linked escrow lifecycle behavior is more complete and less implicit
- downstream tools and docs reflect the landed canonical dispute runtime
- `go build ./...`
- `go test ./...`
- `.venv/bin/zensical build`
- docs-only OpenSpec validation
all pass

## Follow-On Inputs

Natural follow-on work after this workstream:

1. Reputation V2 + Runtime Integration Workstream
- reputation model v2
- trust-entry contract strengthening
- deeper runtime integration
