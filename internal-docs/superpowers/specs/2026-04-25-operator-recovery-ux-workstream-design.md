# Operator Recovery UX Workstream Design

## Purpose / Scope

This workstream upgrades the landed dead-letter operator recovery surface so that retry outcomes are easier to interpret after invocation.

This workstream bundles:

- CLI retry success/failure messaging refinement
- CLI structured result refinement
- cockpit retry success/failure messaging refinement
- cockpit retry state-transition messaging refinement
- CLI / cockpit recovery semantics alignment

The target operator surfaces are:

- `lango status dead-letter retry <transaction-receipt-id>`
- cockpit dead-letter detail pane retry flow

This workstream directly includes:

- clearer retry success wording
- clearer retry failure wording
- better separation of precheck failure vs invocation failure
- more explicit operator-facing result payloads

This workstream does not directly include:

- polling
- action history
- bulk recovery
- new recovery actions
- generic async retry policy changes
- broader replay / recovery substrate redesign

## CLI Recovery UX Scope

The CLI retry action is already landed, so this workstream keeps the existing retry path and refines only the operator-facing experience.

Covered behavior:

- make `can_retry=false` precheck failures clearer and more operator-facing
- distinguish precheck failure from mutation failure
- refine default success output so it reads as retry acceptance/request rather than completion
- refine default failure output so the operator can tell where the failure happened
- refine `json` output shape so scripts receive a clearer success/failure result contract

The CLI flow remains:

1. detail read
2. `can_retry` precheck
3. confirm prompt or `--yes`
4. retry invocation

This workstream upgrades the feedback quality around that flow, not the flow itself.

## Cockpit Recovery UX Scope

The cockpit retry action is also already landed. The existing flow remains:

- inline confirm
- running state
- duplicate retry guard
- success refresh
- failure returns to idle

This workstream refines the operator-facing presentation around that flow:

- clarify success wording in the detail pane
- clarify failure wording in the detail pane / status message
- make `confirm -> running -> success/failure` transitions more legible
- ensure retry feedback remains understandable after refresh and selection-preservation behavior

The cockpit scope is additive UX polish. It does not add a new recovery capability.

## Shared Recovery Semantics

This workstream keeps CLI and cockpit on the same recovery contract.

Shared rules:

- retryability is determined from the existing detail status surface via `can_retry`
- the final mutation gate remains the existing `retry_post_adjudication_execution` path
- success means the retry request was accepted, not that the full background execution has completed
- precheck failure and invocation failure must be distinguishable in operator messaging
- policy gates and evidence gates remain enforced by the control plane, not by the operator surface

The goal is for CLI and cockpit to differ in presentation, not in the meaning of retry outcomes.

## Execution / Parallelization Model

This workstream is intentionally handled as a larger batch.

Execution model:

- one spec
- one implementation plan
- three workers in parallel

### Worker A

Owns:

- `internal/cli/status/*`
- retry output / messaging behavior
- CLI tests

### Worker B

Owns:

- `internal/cli/cockpit/*`
- cockpit retry messaging / rendering behavior
- cockpit tests

### Worker C

Owns:

- `docs/cli/*`
- `docs/architecture/*`
- `README.md`
- `openspec/*`

This keeps the superpowers process intact while reducing overhead from repeated micro-slice cycles.

## Implementation Shape

Recommended implementation shape:

- extend `internal/cli/status`
  - refine retry success output
  - refine retry failure output
  - clarify precheck failure wording
  - refine `json` result shape
  - strengthen retry tests

- extend `internal/cli/cockpit`
  - refine retry state text
  - refine success/failure message wording
  - clarify message priority during state transitions
  - strengthen cockpit retry tests

- update documentation set
  - `docs/cli/status.md`
  - `docs/cli/index.md`
  - `README.md`
  - `docs/architecture/dead-letter-browsing-status-observation.md`
  - `docs/architecture/p2p-knowledge-exchange-track.md`

- sync OpenSpec
  - main docs-only spec
  - archive one workstream change

This workstream is additive. It does not redesign the current dead-letter read model or retry substrate.

## Follow-On Inputs

Natural follow-on work after this workstream:

1. Operator Summary Workstream
- grouped dead-letter summaries
- broader operator entrypoints

2. CLI Recovery UX v2
- polling
- richer result payloads
- follow-up inspection helpers

3. Replay / Recovery Policy Runtime Workstream
- policy-driven defaults
- generic async retry policy
- broader recovery substrate normalization
