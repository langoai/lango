# Grouped Summary Families Workstream Design

## Purpose / Scope

This workstream adds the first grouped summary axis to the dead-letter operator summaries so operators are not forced to interpret only raw latest dead-letter reason strings.

The first slice adds reason-family summary parity across:

- `lango status dead-letter-summary`
- the cockpit `dead-letters` page top summary strip

This workstream directly includes:

- a shared reason-family heuristic classifier
- CLI summary `by_reason_family`
- cockpit summary `reason families: ...`
- preservation of the existing raw top latest dead-letter reasons

This workstream does not directly include:

- actor families
- dispatch families
- configurable taxonomy
- backend summary services
- persisted reason normalization

The goal is reason-family grouped summary parity, not a full analytics or taxonomy subsystem.

## Reason Family Model

Reason families use a small built-in heuristic taxonomy in this first slice.

Initial families:

- `retry-exhausted`
- `policy-blocked`
- `receipt-invalid`
- `background-failed`
- `unknown`

Classification basis:

- the current latest dead-letter reason string for each backlog row
- case-insensitive keyword matching
- fallback to `unknown` when no family matches

This keeps the classifier simple, local, and easy to revise after seeing more real reason strings. The classifier is an operator summary helper, not a canonical persisted reason model.

## CLI Summary Scope

The CLI keeps the existing `lango status dead-letter-summary` command and extends its output additively.

Existing fields remain:

- `total_dead_letters`
- `retryable_count`
- `by_adjudication`
- `by_latest_family`
- `top_latest_dead_letter_reasons`
- `top_latest_manual_replay_actors`
- `top_latest_dispatch_references`

Added field:

- `by_reason_family`

Output behavior:

- table output gains a `By reason family` section
- JSON output gains a `by_reason_family` bucket array

Raw top latest dead-letter reasons remain in place. The grouped family view complements the raw reasons instead of replacing them.

## Cockpit Summary Scope

The cockpit keeps the existing `dead-letters` page top summary strip and extends it additively.

Existing strip lines remain:

1. global overview
2. `reasons: ...`
3. `actors: ...`
4. `dispatch: ...`

Added line:

- `reason families: ...`

Example:

- `reason families: policy-blocked(4), retry-exhausted(3), unknown(1)`

The cockpit strip uses the same current backlog rows and the same reason-family classifier as the CLI summary.

## Shared Semantics

CLI and cockpit summaries must agree on reason-family semantics.

Shared rules:

- aggregation basis is latest dead-letter reason
- classifier taxonomy is identical
- matching is case-insensitive
- fallback family is `unknown`
- raw top latest dead-letter reasons stay visible
- reason-family summary is additive

The implementation should put the classifier in a shared internal helper rather than duplicating keyword rules in CLI and cockpit packages.

## Execution / Parallelization Model

This workstream is handled as a larger batch.

Execution model:

- one spec
- one implementation plan
- three workers in parallel

### Worker A

Owns:

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

The shared classifier helper is owned by Worker A. Worker B reuses that helper to avoid divergent classification rules.

## Implementation Shape

Recommended implementation shape:

- add a shared reason-family classifier
  - classify latest dead-letter reason strings into the initial family taxonomy
  - include focused tests for keyword matching and unknown fallback

- extend `internal/cli/status`
  - aggregate `by_reason_family`
  - render the new table section
  - include `by_reason_family` in JSON output
  - update status CLI tests

- extend `internal/cli/cockpit/pages/deadletters.go`
  - aggregate reason-family buckets from the current backlog rows
  - render a compact `reason families:` strip line
  - update cockpit page tests

- update docs / OpenSpec
  - `docs/cli/status.md`
  - `docs/cli/index.md` if summary copy changes
  - `README.md` if the command summary changes
  - `docs/architecture/dead-letter-browsing-status-observation.md`
  - `docs/architecture/p2p-knowledge-exchange-track.md`
  - `openspec/specs/docs-only/spec.md`
  - one archive change for the workstream

This workstream is additive. It does not change existing retry behavior, list filters, backend read models, or raw top reason summaries.

## Follow-On Inputs

Natural follow-on work after this workstream:

1. actor family summary
- operator / system / service grouping

2. dispatch family summary
- dispatch prefix or source grouping

3. richer taxonomy evolution
- configurable family map
- persisted reason normalization
- trend/time-window summaries
