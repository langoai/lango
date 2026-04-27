# Reputation V2 + Runtime Integration Workstream Design

## Purpose / Scope

This workstream upgrades the trust / reputation model from the currently stabilized first slices into a stronger runtime contract that later systems can consume without ambiguity.

The target surfaces are:

- owner-root trust versus earned agent/domain reputation
- trust-entry semantics for new and returning peers
- runtime consumption of reputation and trust signals

This workstream directly includes:

- reputation model v2
- stronger trust-entry contracts
- deeper runtime integration

This workstream does not directly include:

- dispute-core redesign
- operator analytics or operator UI redesign
- broader dead-letter or recovery UX work
- unrelated product-surface expansion

The goal is to move reputation from an audited and partially consumed concept into a clearer canonical model that runtime systems use more consistently.

## Current Baseline

The codebase already includes:

- P2P identity continuity
- root trust and peer admission signals
- reputation storage and runtime readers
- trust-sensitive pricing and payment friction
- team and coordination bridges that react to reputation changes
- architecture audits that stabilized the first model

The remaining gaps are now concentrated:

- bootstrap trust versus earned reputation is still not explicit enough
- operational safety signals and durable negative reputation are still too easy to conflate
- adjudicated dispute outcomes do not yet feed into a stronger reputation contract
- runtime systems consume reputation, but not yet through one clearly finished V2 model

This makes a dedicated final workstream around reputation semantics and runtime integration feasible and timely.

## Workstream Scope

This workstream is intentionally centered on meaning first, then consumption.

### 1. Reputation V2 Model

Clarify the canonical distinction between:

- owner-root trust
- earned agent/domain reputation
- transaction-local trust observations
- durable negative reputation
- temporary operational safety signals

The first implementation should:

- preserve the already-landed owner-root trust baseline
- keep agent/domain reputation earned from actual history
- prevent temporary runtime safety events from automatically becoming permanent durable reputation damage
- make adjudicated negative outcomes eligible for stronger durable impact than unaudited operational failures

### 2. Stronger Trust-Entry Contract

Clarify how the system treats:

- first-time peers
- returning peers
- low-trust but not-yet-banned peers
- temporarily unsafe peers

The first implementation should:

- make bootstrap trust and earned trust separate but connected
- align admission, approval, payment friction, and runtime collaboration entry around the same trust-entry meaning
- avoid collapsing all trust-entry decisions into a single scalar score

### 3. Deeper Runtime Integration

Make runtime systems consume the stronger trust/reputation contract more directly.

Likely integration points include:

- pricing / risk
- payment-gate trust checks
- firewall / admission checks
- team and coordination bridges
- reputation-sensitive runtime automation or selection

The first implementation should:

- preserve current runtime behavior where it is already consistent
- tighten the contract where multiple subsystems currently read reputation differently
- avoid turning this workstream into a broad operator-surface project

## Architectural Shape

This workstream should preserve the current split between canonical trust meaning and consuming runtimes.

### Reputation Boundary

Primary ownership is expected to fall around:

- `internal/p2p/reputation/*`
- related trust / reputation readers and bridge helpers

This boundary should own:

- the stronger V2 meaning of reputation
- durable versus temporary negative signal handling
- canonical trust-entry-related read helpers

### Runtime Consumer Boundary

Secondary ownership is expected to fall around:

- `internal/app/wiring_p2p.go`
- `internal/app/wiring_economy.go`
- `internal/p2p/firewall/*`
- `internal/p2p/team/*`
- relevant pricing / paygate / selection consumers

This boundary should consume the reputation contract, not redefine it.

### Docs / OpenSpec Boundary

Primary ownership is expected to fall around:

- `docs/architecture/*`
- `docs/features/*` where reputation is user-visible
- `docs/cli/*` only where behavior changes are actually surfaced
- `openspec/*`

Docs should describe only runtime behavior and semantics that are actually landed.

## Execution / Parallelization Model

This workstream should be treated as a larger batch with moderate parallelism.

Execution model:

- one spec
- one implementation plan
- three workers in parallel
- implementation remains semi-serial around the canonical reputation contract

### Worker A

Owns:

- reputation / trust contract core
- canonical store / bridge semantics
- primary trust-model tests

### Worker B

Owns:

- runtime consumer integration
- pricing / paygate / admission / team bridge alignment
- focused integration tests

### Worker C

Owns:

- `docs/architecture/*`
- `docs/features/*`
- `docs/cli/*` only when needed
- `README.md` when needed
- `openspec/*`

Worker A establishes the canonical V2 trust / reputation contract first. Worker B should align consumers with that contract rather than invent local interpretations.

## Implementation Strategy

This workstream should be implemented as one plan with grouped task bands rather than more micro-workstreams.

Recommended task bands:

1. map the current reputation / trust-entry semantics and runtime consumers
2. land the stronger canonical V2 reputation contract
3. align runtime trust-entry behavior across key consuming subsystems
4. integrate adjudicated dispute outcomes and operational signals under the stronger semantics
5. truth-align docs / OpenSpec
6. final integrated verification

The sequence matters because the canonical meaning must settle before downstream consumers and docs are aligned.

## Completion Criteria

The workstream is complete when:

- the V2 reputation contract is explicit in code
- bootstrap trust, earned trust, and durable negative impact are more clearly separated
- runtime consumers read a more consistent trust / reputation contract
- docs and OpenSpec reflect the landed semantics
- `go build ./...`
- `go test ./...`
- `.venv/bin/zensical build`
- docs-only OpenSpec validation
all pass

## Follow-On Inputs

Natural follow-on work after this workstream:

1. broader runtime integration polish
- remaining long-tail consumer alignment

2. future product-surface expansion
- operator-facing trust / reputation controls only after the core semantics are stable
