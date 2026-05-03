# Multi-Agent Coworking Wave 5 Design

## Strategic Goal

By the end of Wave 4, Lango can show:

- durable missions
- transient proposals
- live decisions
- operating loops and agenda

Wave 5 should make collaboration inside those work units visible. The user should be able to tell not only what work exists, but also how multiple agents are cooperating, blocking each other, handing off, reviewing, or consuming budget inside one mission.

## Product Intent

Wave 5 should answer:

> "Which agents are working on this mission, what role is each one playing, where is the handoff, and what is stuck?"

This is not just more log lines. It is a higher-level collaboration surface.

## Fixed Decisions

The following decisions are fixed for Wave 5:

- the first slice focuses on local built-in teammate collaboration signals
- external P2P team surfaces remain secondary until local coworking is clear
- collaboration state remains attached to missions; it does not become a separate durable work model
- the first slice should prioritize observability of collaboration over editing/delegation controls
- budget/trust/conflict surfaces must be grounded in real existing runtime signals

## Why Coworking Needs Its Own Wave

Current Mission Control already shows:

- who owns a mission at a point in time
- recent activity and runtime hints
- live decisions

But it still does not clearly answer:

- who delegated to whom
- whether a teammate is blocked on approval
- whether the mission is in a review or handoff phase
- whether budget pressure is building inside a multi-agent effort

That is the gap Wave 5 addresses.

## First-Slice Collaboration Sources

Wave 5 should start only from sources that already exist in code today.

### Source 1: Local Delegation Events

The runtime already emits delegation and budget-warning events into the TUI/runtime surfaces. These are the best first collaboration signals because they already describe local multi-agent work.

Initial uses:

- handoff edges
- current collaborating agents
- delegation count and pressure

First-slice attribution rule:

- local delegation signals may only be attached to a mission when they can be attributed through a mission-linked local execution
- if multiple unrelated missions share a session and no clear execution-linked attribution exists, the signal stays session-level and must not be projected onto a mission row

### Source 2: AgentRun / Teammate Runtime State

AgentRun already exposes:

- requested agent
- runtime condition
- blocked reason
- approval wait state

These are direct coworking-state signals for local teammate execution.

### Source 3: RunLedger / Mission Execution Links

Wave 2 already links mission to execution identities. Wave 5 should use those links to tie delegation and teammate runtime back into the mission surface.

### Source 4: Budget And Recovery Signals

Budget warnings and recovery decisions are already emitted. They should become collaboration-state indicators inside the mission, not just transcript noise.

First-slice attribution rule:

- budget and recovery signals may only appear on a mission when they can be attributed through that mission's linked active local execution
- otherwise they remain session-level runtime context rather than mission collaboration state

### Source 5: External Team Runtime (Deferred In First Slice)

The codebase contains P2P team coordination, but the first Wave 5 slice should not try to solve full external team UX at the same time.

That means:

- local built-in teammate coworking is enabled first
- external P2P team state is explicitly deferred or shown only in limited operator detail later

## Non-Sources For The First Slice

The first slice should not claim:

- fully unified external + internal team timelines
- trust-weighted conflict resolution surfaces in Mission Control
- editable team role assignment from the cockpit
- rich workspace artifact review lanes

Those belong to later expansion once the local coworking model is stable.

## Collaboration Model

Wave 5 introduces a transient collaboration projection per mission.

Suggested fields:

- `mission_id`
- `participants`
- `active_owner`
- `handoff_edges`
- `collaboration_state`
- `budget_signal`
- `blocked_on`
- `last_recovery`
- `updated_at`

Suggested collaboration states:

- `solo`
- `delegating`
- `waiting_on_teammate`
- `reviewing`
- `blocked_on_approval`
- `recovering`

This does not need a new durable DB table in the first slice. It can remain a derived projection over mission-linked execution and runtime events.

## Participant Model

The user should see named participants rather than opaque internal IDs.

Initial participant sources:

- mission owner agent
- requested agent from linked `AgentRun`
- delegation source/target names from runtime events

The first slice should not require full durable participant history. It only needs a current collaboration view plus recent handoffs.

## Handoff Semantics

Handoffs should be visible as edges, not raw event spam.

First-slice rules:

- show the most recent delegation edge for a mission only when it is attributable through mission-linked local execution
- accumulate a small recent handoff set only from attributable mission-linked signals
- session-level delegation noise must not be over-attributed to a mission

This is enough to answer “who handed this off to whom?”

## Review Semantics

Wave 5 should start with lightweight review semantics, not a full PR-style review system.

A mission may enter `reviewing` when:

- a mission is done but still has a review-needed follow-up
- a mission-linked local execution is in a `verify_pending` or equivalent orchestrator-review-needed state

The first slice should not invent artificial review state where no signal exists.

Explicit non-sources for `reviewing` in the first slice:

- generic delegation count alone
- generic approval wait alone
- vague “multiple agents touched this mission”

## Budget And Recovery Semantics

Budget warnings and recovery decisions already exist as local runtime signals.

Wave 5 should:

- attach budget pressure to the mission collaboration summary
- attach latest recovery action to the mission collaboration summary
- avoid exposing low-level event spam first

Examples:

- `Budget pressure: 12/15 delegations`
- `Recovering: retry_with_hint`

## Mission Control Surface

Wave 5 should extend existing mission rows and detail rendering rather than building a separate collaboration dashboard first.

Suggested additions:

- participant summary on mission row or detail line
- handoff summary in mission detail
- collaboration-state badge or text
- budget/recovery hints when relevant

The first slice should stay compact enough for terminal use.

## Relationship To Wave 4 Loops

Loops and agenda stay above collaboration state.

Wave 5 should not turn loop rows into giant team dashboards. Instead:

- loops answer “what needs attention”
- coworking answers “how agents are collaborating on it”

These surfaces should complement each other.

## Relationship To External P2P Teams

The repo already has external team coordination primitives. Wave 5 should acknowledge them but not overcommit.

First-slice rule:

- external team data is not the primary source for coworking in Mission Control
- future slices may add an external-team adapter after the local built-in teammate story is stable

## Architecture Sketch

Suggested components:

- `CollaborationProjector`
  - mission-linked collaboration summary
  - participant extraction
  - handoff aggregation
  - budget/recovery attachment

- `CollaborationSourceAdapters`
  - delegation/runtime source
  - agent run source
  - mission execution-link source

The first slice should keep this as a projection layer, not a durable domain.

## Canonical Boundary

The first slice is specifically about **mission-linked local coworking**.

That means:

- `AgentRun`, mission execution links, and mission-linked local runtime signals are primary
- session-level delegation/budget/recovery events are only usable after mission attribution is proven
- external team coordination remains secondary until a dedicated adapter is added

## Non-Goals

Wave 5 does not aim to:

- build a full external team cockpit
- add cockpit controls for team formation or role editing
- create durable collaboration history storage
- solve artifact review/workspace review comprehensively

## Success Criteria

Wave 5 is successful when:

- Mission Control can show which agents are collaborating on a mission
- recent handoffs are visible in mission context
- blocked-on-approval / waiting-on-teammate / recovery / budget pressure become mission-visible collaboration state
- the first slice stays grounded in real local runtime signals
