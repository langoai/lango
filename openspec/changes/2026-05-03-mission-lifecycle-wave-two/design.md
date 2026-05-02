# Design

## Problem

Mission Control currently shows mission-shaped work, but the rows are still assembled from background tasks, learning suggestions, approvals, and optional runtime readers. That makes the UI useful, but not durable. A mission can disappear once the runtime projection changes, and there is no stable `mission_id` that survives retries, follow-up executions, or return visits.

At the same time, `RunLedger` already serves as durable execution truth. Wave 2 must avoid duplicating execution storage or pretending that every execution row is itself the user-facing mission record.

## Goals

- add durable first-class mission records with separate `mission_id`
- keep `RunLedger` as execution truth while adding a mission persistence layer for user-facing work state
- make Mission Control read durable mission rows first
- preserve unmatched runtime overlays until the work is linked or dismissed
- support real write paths for direct mission start and proposal acceptance
- represent approval pauses as coarse durable `waiting_decision` state
- keep task tracking lightweight instead of turning it into durable mission truth

## Non-Goals

- auto-creating durable missions for every runtime execution
- a full second journal that duplicates `RunLedger`
- durable storage for transient `proposed` rows before acceptance
- a durable approval queue with per-request rendering history
- making `TaskEntry` the official durable mission checklist system

## Storage Strategy

Wave 2 uses hybrid storage.

- `RunLedger` remains the durable execution truth
- a new mission persistence layer stores the latest durable mission row
- a mission history table stores append-only state transitions
- a mission-execution link table stores the durable relationship from mission to execution identity

This split is required because a mission may exist before any execution exists, may survive multiple executions, and needs fast durable reads without replaying execution journals.

## Durable Mission Model

### Mission

The durable `Mission` row is the primary user-facing record. It has its own `mission_id`, session association, title/description, current lifecycle status, source metadata, current decision/blocking summary, and timestamps.

Durable mission rows begin at `prepared`, not `proposed`. `proposed` remains a transient Mission Control overlay that only becomes durable once the user accepts it.

Durable statuses in Wave 2 are:

- `prepared`
- `active`
- `waiting_decision`
- `blocked`
- `done`
- `cancelled`

### MissionStateHistory

`MissionStateHistory` stores append-only transitions for auditability and replay of the coarse mission lifecycle. Each row records `from_status`, `to_status`, actor metadata, optional execution references, optional decision summary, optional payload, and a per-mission sequence.

Wave 2 intentionally stops at latest-row plus history. It does not introduce a second general-purpose event journal abstraction.

### MissionExecutionLink

`MissionExecutionLink` is the durable truth for mission-to-execution relationships.

- one mission may have zero or many linked executions
- the link identifies execution kind, execution reference, and link role
- reverse lookup from execution to mission must work without inferring from task tracking tables

Execution truth stays in the execution system. Relationship truth lives here.

## Read Model

Mission Control becomes durable-first.

1. read durable mission rows for the session
2. enrich from linked execution truth where needed
3. overlay unmatched runtime work that has not been linked to a durable mission
4. retain live activity and pending-decision surfaces needed for honest current-state rendering

This preserves Wave 1 honesty. Runtime work does not disappear merely because it has not yet been linked into the durable mission graph.

## Write Paths

Wave 2 introduces two real mission creation paths:

### Direct Mission Start

When the user starts a mission from Mission Control, the app creates a durable mission row immediately. The initial status is `prepared` if execution has not begun yet, or `active` if execution starts as part of the same action.

### Proposal Acceptance

When the user accepts a proposed mission, acceptance creates the first durable mission row. The transient proposal overlay is not itself durable; the durable record begins at acceptance.

## Decision and Blocking Semantics

Wave 2 stores `waiting_decision` as a coarse durable mission state. It records that the mission is paused on user direction, plus a latest decision kind/summary if available.

Wave 2 does not store a durable approval queue. Live approval rendering can still use runtime/session-owned structures while the durable mission row keeps only coarse mission state.

Non-decision blockers remain `blocked`.

## Task Tracking Boundary

`TaskEntry` remains lightweight in Wave 2.

- it may continue serving local operational tracking
- it is not promoted into the durable mission checklist model
- mission truth does not depend on retrofitting all task tracking into mission persistence

This keeps scope bounded while leaving room for optional future linkage.

## Execution Link Attachment

Mission-aware execution linkage must attach at execution creation sites.

- when mission-bound work creates a background execution, the app should write the `MissionExecutionLink` then
- when mission-bound work creates a `RunLedger`-tracked execution, the app should write the link then
- Wave 2 should not rely on later inference from unrelated task-tracking records to reconstruct mission ownership

## Validation

This change is valid only if:

- the design keeps hybrid storage instead of replacing `RunLedger`
- durable mission rows begin at `prepared`, not `proposed`
- `mission_id` is separate from execution identity
- `MissionExecutionLink` is the durable execution linkage truth
- Mission Control reads durable mission rows first but still keeps unmatched runtime overlay visibility
- direct mission start and proposal acceptance are both real durable write paths
- `waiting_decision` is coarse durable state, not a durable approval queue
- `TaskEntry` remains lightweight and outside the durable mission checklist model
