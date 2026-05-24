# Actor Family Summary Workstream Design

## Purpose / Scope

This workstream adds the first grouped actor-summary axis to the landed dead-letter operator summaries so operators can quickly understand what kind of actor is most associated with recent manual replay activity.

The first slice adds actor-family summary parity across:

- `lango status dead-letter-summary`
- the cockpit `dead-letters` page top summary strip

This workstream directly includes:

- a shared actor-family heuristic classifier
- CLI summary `by_actor_family`
- cockpit summary `actor families: ...`
- preservation of the existing raw top latest manual replay actors

This workstream does not directly include:

- dispatch families
- configurable actor taxonomy
- backend summary services
- persisted actor normalization
- history-wide actor accumulation

The goal is actor-family grouped summary parity, not a full identity or operator analytics subsystem.

## Actor Family Model

Actor families use a small built-in heuristic taxonomy in this first slice.

Initial families:

- `operator`
- `system`
- `service`
- `unknown`

Classification basis:

- the current latest manual replay actor string for each backlog row
- case-insensitive keyword or prefix matching
- fallback to `unknown` when no family matches

Recommended first-slice heuristics:

- values like `operator:alice` map to `operator`
- runtime or automated replay identities map to `system`
- service, bridge, or integration identities map to `service`
- anything else falls back to `unknown`

This keeps the taxonomy local, readable, and easy to revise after more real actor strings are observed.

## CLI Summary Scope

The CLI keeps the existing `lango status dead-letter-summary` command and extends its output additively.

Existing fields remain:

- `total_dead_letters`
- `retryable_count`
- `by_adjudication`
- `by_latest_family`
- `by_reason_family`
- `top_latest_dead_letter_reasons`
- `top_latest_manual_replay_actors`
- `top_latest_dispatch_references`

Added field:

- `by_actor_family`

Output behavior:

- table output gains a `By actor family` section
- JSON output gains a `by_actor_family` bucket array

Raw top latest manual replay actors remain in place. The grouped family view complements the raw actors instead of replacing them.

## Cockpit Summary Scope

The cockpit keeps the existing `dead-letters` page top summary strip and extends it additively.

Existing strip lines remain:

1. global overview
2. `reasons: ...`
3. `reason families: ...`
4. `actors: ...`
5. `dispatch: ...`

Added line:

- `actor families: ...`

Recommended placement:

- after `actors: ...`
- before `dispatch: ...`

Example:

- `actor families: operator(5), system(2), unknown(1)`

The cockpit strip uses the same current backlog rows and the same actor-family classifier as the CLI summary.

## Shared Semantics

CLI and cockpit summaries must agree on actor-family semantics.

Shared rules:

- aggregation basis is latest manual replay actor
- classifier taxonomy is identical
- matching is case-insensitive
- fallback family is `unknown`
- raw top latest manual replay actors stay visible
- actor-family summary is additive

The implementation should put the classifier in a shared internal helper rather than duplicating string rules in CLI and cockpit packages.

## Execution / Parallelization Model

This workstream is handled as a larger batch.

Execution model:

- one spec
- one implementation plan
- three workers in parallel

### Worker A

Owns:

- shared actor-family helper
- `internal/cli/status/*`
- CLI summary result changes
- CLI table / JSON rendering
- CLI tests

### Worker B

Owns:

- `internal/cli/cockpit/*`
- cockpit summary aggregation
- strip rendering
- cockpit tests

### Worker C

Owns:

- `docs/cli/*`
- `docs/architecture/*`
- `README.md` when necessary
- `openspec/*`

The shared classifier helper is owned by Worker A. Worker B reuses that helper to avoid divergent actor-family rules.

## Implementation Shape

Recommended implementation shape:

- add a shared actor-family classifier
  - classify latest manual replay actor strings into the initial taxonomy
  - include focused tests for operator, system, service, and unknown fallback

- extend `internal/cli/status`
  - aggregate `by_actor_family`
  - render the new table section
  - include `by_actor_family` in JSON output
  - update status CLI tests

- extend `internal/cli/cockpit/pages/deadletters.go`
  - aggregate actor-family buckets from the current backlog rows
  - render a compact `actor families:` strip line
  - update cockpit page tests

- update docs / OpenSpec
  - `docs/cli/status.md`
  - `docs/cli/index.md` if summary copy changes
  - `README.md` if the command summary changes
  - `docs/architecture/dead-letter-browsing-status-observation.md`
  - `docs/architecture/p2p-knowledge-exchange-track.md`
  - `openspec/specs/docs-only/spec.md`
  - one archive change for the workstream

This workstream is additive. It does not change existing retry behavior, filter behavior, backend read models, or the raw top-actor summary.

## Follow-On Inputs

Natural follow-on work after this workstream:

1. dispatch family summary
- dispatch prefix or source grouping

2. summary evolution
- configurable family maps
- richer top-N
- trend/time-window summaries

3. broader operator surface consolidation
- dead-letter CLI `any_match_family`
- polling / follow-up recovery UX
- richer structured retry results
