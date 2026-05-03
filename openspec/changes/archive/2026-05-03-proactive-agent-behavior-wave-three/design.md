# Design

## Problem

Mission Control can already show transient proposed missions, but the current shape is too thin for meaningful proactive behavior. Raw learning suggestions are visible, yet they do not own an explicit transient lifecycle, they do not carry a prepared brief, and acceptance does not define how prepared context survives into the durable mission path.

At the same time, Wave 3 must avoid overreaching. It should not silently launch generic proposal-owned executions, should not enable vague heuristic producers, and should not blur the boundary between transient proactive work and durable accepted mission truth.

## Goals

- add an explicit transient proposal model for proactive work
- make `LearningSuggestionEvent` the only active producer in the first slice
- generate a deterministic prepared brief from source-native evidence
- render prepared proposals honestly in Mission Control
- preserve prepared context when a proposal is accepted into the durable mission lifecycle

## Non-Goals

- enabling librarian-gap or inquiry producers in this slice
- enabling runtime-failure or blocked-work producers in this slice
- launching generic proposal-owned background or `RunLedger` executions
- introducing durable `proposed` mission rows
- using LLM synthesis for proposal preparation

## Proposal Model

Wave 3 introduces a transient proposal registry distinct from durable mission storage.

Each proposal is session-scoped and source-attributed. The first slice needs, at minimum:

- `proposal_id`
- `session_key`
- `source_kind`
- `source_ref`
- `title`
- `summary`
- `reason`
- `confidence`
- `status`
- `prepared_brief`
- `created_at`
- `updated_at`
- `expires_at`

The proposal lifecycle in this slice is:

- `suggested`
- `preparing`
- `prepared`
- `dismissed`
- `accepted`
- `expired`

`proposed` remains transient. No durable `Mission` row exists before acceptance.

## Active Producer Contract

`LearningSuggestionEvent` is the only active producer in this slice.

Producer contract:

- input: `LearningSuggestionEvent`
- source identity: `source_kind=proposed_learning`, `source_ref=SuggestionID`
- output: one transient proposal record
- preparation allowed: yes, but only through deterministic source-native brief generation

Librarian-gap and runtime-failure producers are explicitly deferred until they have dedicated adapters and stable source identities.

## Deterministic Preparation

Preparation in this slice is not broad agent work. It is a deterministic transformation of already available source evidence into a prepared brief.

The prepared brief should summarize:

- the proposal source
- the rule or recommendation being suggested
- the rationale already attached to the source event
- the expected effect of acceptance
- any stable supporting evidence already present on the source event

Preparation rules:

- no LLM call
- no filesystem mutation
- no external messaging
- no payments
- no dangerous command execution
- no generic proposal-owned execution launch

This keeps preparation explainable, low-risk, and testable.

## Mission Control Behavior

Mission Control should render proposals from the transient proposal registry rather than directly from raw learning-buffer rows.

For each active proposal, the page should surface:

- title
- one-line reason
- preparation state
- prepared brief summary when ready
- source metadata

This slice must support visible `suggested`, `preparing`, and `prepared` states. A prepared proposal is a user-facing artifact, not merely an internal precondition.

## Acceptance and Ownership

Acceptance is the boundary where transient proactive work becomes durable mission truth.

Acceptance flow:

1. the user accepts a transient prepared proposal
2. the application creates a durable mission through `MissionService.AcceptProposal(...)`
3. the durable mission starts at `prepared` or `active`
4. the prepared brief context is preserved on the resulting durable mission path instead of being discarded
5. the transient proposal leaves the active proposal set

This slice does not support generic proposal-owned execution promotion because generic proposal-owned execution is out of scope. The only prepared context that must survive acceptance here is the deterministic prepared brief and its source-native evidence.

## Validation

This change is valid only if:

- `proposed` remains transient
- `LearningSuggestionEvent` is the only active producer in this slice
- preparation is deterministic and source-native
- Mission Control renders prepared proposals from a proposal registry
- acceptance creates a durable mission while preserving prepared brief context
- librarian-gap producers remain out of scope
- runtime-failure producers remain out of scope
- generic proposal-owned executions remain out of scope
