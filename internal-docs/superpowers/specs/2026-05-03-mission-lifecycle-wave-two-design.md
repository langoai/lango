# Mission Lifecycle Wave 2 Design

## Strategic Goal

Wave 1 made Mission Control visible as a projection over existing runtime facts. Wave 2 turns missions into durable first-class records without collapsing them back into `RunLedger` runs or background tasks.

The core shift is:

- Wave 1: missions are derived UI rows
- Wave 2: missions are durable work units with their own identity and lifecycle

This design keeps execution truth in `RunLedger` while adding a separate mission persistence layer for user-facing work state.

## Wave 2 Decisions

The following design choices are fixed for Wave 2:

- storage strategy: hybrid
- mission creation policy: user-started work plus approved proposed missions
- product state space still includes transient `proposed` suggestions in Mission Control, but durable mission rows start at `prepared`
- durable mission state space: `prepared`, `active`, `waiting_decision`, `blocked`, `done`, `cancelled`
- execution relation: model allows `1 mission : N executions`, while early UX may still behave mostly like `1:1`
- mission identity: separate `mission_id`
- mission may exist before any execution exists
- persistence shape: mission row with latest-state fields plus state-history table
- `TaskEntry` remains a separate lightweight tracker, only loosely linked to missions

## Why Hybrid

`RunLedger` already acts as durable execution truth. It is good at journaling execution, validation, and planner structure. It is not a good top-level home for the human work unit itself.

A separate mission layer is needed because:

- a mission may exist before any run exists
- a mission may survive multiple retries or follow-up runs
- a mission needs latest user-facing state that is not identical to execution state
- a mission must be queryable without replaying execution journals

At the same time, a full mission event journal would duplicate too much of `RunLedger`. Wave 2 therefore stores mission latest state and mission state history in Ent, while linking execution records through lightweight references.

## Core Model

Wave 2 adds three new durable concepts:

1. `Mission`
2. `MissionStateHistory`
3. `MissionExecutionLink`

### Mission

`Mission` is the durable, user-facing work unit.

Suggested fields:

- `id` UUID
- `session_key` string, required and indexed in the first durable slice
- `title` string
- `description` text, optional
- `status` enum:
  `prepared`, `active`, `waiting_decision`, `blocked`, `done`, `cancelled`
- `source_kind` string:
  `user`, `proposed_learning`, `proposed_system`, `runledger`, `background`, `manual`
- `source_ref` string, optional
- `current_blocked_reason` string, optional
- `current_decision_kind` string, optional
- `current_decision_summary` string, optional
- `created_at`
- `updated_at`
- `completed_at`, nullable

The `Mission` row is optimized for fast reads in Mission Control and later operator surfaces.
Wave 2 intentionally does not persist durable `proposed` missions. `proposed` remains a transient suggestion overlay until acceptance creates a durable mission row in `prepared` or `active`.

### MissionStateHistory

`MissionStateHistory` captures transitions and supporting audit detail.

Suggested fields:

- `id` UUID
- `mission_id` UUID
- `seq` int64 per mission
- `from_status` enum, nullable on first event
- `to_status` enum
- `reason` string, optional
- `actor_kind` string:
  `user`, `agent`, `system`
- `actor_ref` string, optional
- `execution_kind` string, optional
- `execution_ref` string, optional
- `decision_kind` string, optional
- `decision_summary` string, optional
- `payload` JSON, optional
- `created_at`

Wave 2 does not need a second “mission journal” abstraction in code. A simple append-only history table is enough.

### MissionExecutionLink

`MissionExecutionLink` links a mission to one or more execution units.

Suggested fields:

- `id` UUID
- `mission_id` UUID
- `execution_kind` string:
  `runledger_run`, `task_os_execution`
- `execution_ref` string
- `link_role` string:
  `primary`, `followup`, `retry`, `research`, `draft`, `handoff`
- `created_at`

Suggested uniqueness:

- unique on `(mission_id, execution_kind, execution_ref)`
- index on `(execution_kind, execution_ref)` for reverse lookup from execution completion back to mission
- index on `(mission_id, link_role)` for fast primary/follow-up fetches

`task_os_execution` is the canonical execution identity for agent-driven background work because `AgentRun.ID` and background task ID are intentionally unified today. Wave 2 should not record the same real execution twice as both `background_task` and `agent_run`.

## State Semantics

Wave 2 mission states are user-facing states, not direct execution mirrors.

- `prepared`
  Accepted or system-created mission with enough structure to act, but no active execution yet

- `active`
  Current mission with active work underway or ready for immediate continuation

- `waiting_decision`
  Progress is paused because user direction is required

- `blocked`
  Progress is paused because a non-decision blocker exists:
  missing dependency, external failure, unavailable input, policy gate, or execution breakage

- `done`
  Mission goal is satisfied

- `cancelled`
  Mission is intentionally abandoned

### State Transition Rules

Allowed early transitions:

- `prepared -> active`
- `prepared -> waiting_decision`
- `prepared -> blocked`
- `active -> waiting_decision`
- `active -> blocked`
- `active -> done`
- `active -> cancelled`
- `waiting_decision -> active`
- `waiting_decision -> blocked`
- `waiting_decision -> cancelled`
- `blocked -> active`
- `blocked -> cancelled`

Wave 2 intentionally does not introduce `failed` as a top-level mission state. Execution may fail, but the mission usually remains `blocked` until the user or runtime decides what to do next.

## Creation Rules

Wave 2 durable mission creation happens only through two channels:

1. direct user-started work
2. explicit acceptance of a proposed mission

Wave 2 does not automatically create durable missions for every background task or run. Transient `proposed` overlays may still exist before a durable mission row is created.

### Direct User Start

Examples:

- user submits a top-level objective in Mission Control
- user explicitly creates a mission through future command or UI action

Creation result:

- new `mission_id`
- initial `Mission.status = prepared` or `active` depending on whether execution starts immediately
- first `MissionStateHistory` row inserted

### Proposed Mission Acceptance

Examples:

- user accepts a learning-based proposed mission
- user accepts a future system suggestion

Creation result:

- new `mission_id`
- `source_kind` reflects proposal source
- accepted proposal may optionally create one linked execution immediately
- accepted proposal is the point where the durable mission row first exists

### Non-Creation Cases

The following do not create a durable mission in Wave 2 by default:

- every background task
- every `RunLedger` run
- every approval request
- every `TaskEntry`

These may still be linked to an existing mission. They also remain eligible to appear as unmatched runtime overlays in Mission Control until linked or dismissed.

## Execution Linking

Wave 2 should allow one mission to outlive one execution.

Initial implementation policy:

- a mission may have zero or one primary execution at creation time
- later follow-up or retry executions can be linked through `MissionExecutionLink`
- Mission Control may still show mostly one execution summary per mission in early UX

### RunLedger

`RunLedger` remains the durable execution truth.

Wave 2 should not embed mission lifecycle into `RunLedger` snapshots directly. Instead:

- mission row stores current user-facing status
- `MissionExecutionLink` links mission to `run_id`
- selected `RunSnapshot` data may be used to refresh mission latest fields

This means mission status may be derived from execution events, but is not stored only inside execution storage.

### Background Task

Background tasks remain a lightweight execution surface.

Wave 2 may attach a background task to a mission by storing:

- `execution_kind = task_os_execution`
- `execution_ref = task_id`

No requirement exists to make background tasks durable themselves in this wave.

### Agent Run

Agent runs remain execution/runtime records. Agent runtime conditions can influence mission latest state, but the canonical execution identity stays singular. Wave 2 should not create a second `MissionExecutionLink` row only because agent metadata exists for the same unified task execution ID.

## Relation to TaskEntry

`TaskEntry` stays separate in Wave 2.

Reason:

- `TaskEntry` is currently a lightweight operational tracker in `agentrt`
- it is not durable in the same sense as mission state
- promoting it into the official mission checklist model would widen scope too much

Wave 2 relation:

- mission may optionally refer to task-tracking items later
- `TaskEntry` may eventually gain an optional `mission_id` field
- mission correctness must not depend on `TaskEntry`

This keeps task tracking loosely linked and avoids rewriting agent control tools during the first mission-lifecycle slice.

## Relation to Ontology

Wave 2 stores mission truth in Ent first, not in ontology or graph storage.

This is deliberate:

- mission lifecycle needs strong latest-state queries and transactional updates
- ontology should not become the primary operational store yet

However, Wave 2 naming must leave room for later promotion:

- `Mission` can map to goal-like ontology entities
- mission history and execution links can map to activity/outcome relations later

The bridge comes after the lifecycle is stable.

## Read and Write Model

### Write Path

Wave 2 should introduce a mission service responsible for:

- create mission
- accept proposed mission
- attach execution to mission
- transition mission status
- append mission state history

This service is the only writer of mission latest state and history.
It is also the only writer of `MissionExecutionLink`.

### Read Path

Mission Control should stop projecting durable mission state only from raw runtime surfaces once Wave 2 is landed.

New read order:

1. durable mission rows
2. latest linked execution summaries
3. live pending approval and live runtime overlays

Mission Control still overlays live approval and activity buffers for freshness, but mission identity and lifecycle come from durable mission storage first. Unmatched runtime overlays must remain visible until linked, cancelled, or explicitly ignored so Wave 2 does not regress Wave 1 visibility.

## Failure Policy

Mission writes should be best-effort only where they are secondary to execution, and authoritative where mission creation itself is the user action.

Examples:

- user explicitly creates mission:
  mission persistence failure is user-visible and should fail the action

- linked run completes and mission refresh tries to mark `done`:
  if mission update fails, keep execution truth and surface degraded mission status rather than corrupt execution

Wave 2 should not invent a distributed transaction between mission store and runledger store.

## Waiting Decision Semantics

Wave 2 allows at most one durable latest decision marker per mission row:

- `current_decision_kind`
- `current_decision_summary`
- `status = waiting_decision`

This is not a durable approval queue. The live pending approval path remains authoritative for active decision content. The mission row only records that user direction is currently required and why.

Rules:

- a mission enters `waiting_decision` only when mission service can deterministically attribute a live decision to that mission
- multiple simultaneous pending approvals for one mission remain out of scope in this wave
- if approval resolves but mission persistence fails, live approval resolution wins and mission row may remain stale until repair; this should surface as degraded mission state rather than block execution truth
- `current_decision_summary` is durable descriptive text, not a foreign key to a durable approval record

## Primary Execution Truth

`MissionExecutionLink` is authoritative for mission-to-execution relationships.

Wave 2 deliberately does not store `primary_execution_kind` / `primary_execution_ref` on the `Mission` row. If a later performance slice needs those values on the row, they should be treated as derived cache only, never as a second source of truth.

## Testing Strategy

Wave 2 will need:

- mission service unit tests for creation and state transitions
- store tests for latest-state + history append behavior
- linking tests for `MissionExecutionLink`
- Mission Control projector tests updated to prefer durable mission rows
- failure tests for missing linked execution, history append ordering, and duplicate execution links

## Non-Goals

Wave 2 does not include:

- automatic mission creation from every execution
- ontology-backed operational truth
- full mission event sourcing
- durable background-task persistence
- conversion of `TaskEntry` into the official mission checklist model
- full proactive mission generation from arbitrary runtime observations

## Deliverable Shape

At the end of Wave 2, the system should support:

- durable `mission_id`
- durable latest mission state
- mission state history
- optional links from mission to one or more executions
- Mission Control reading durable missions first

This is the minimum shape that turns Mission Control from a clever projection into a true mission-native operating surface.
