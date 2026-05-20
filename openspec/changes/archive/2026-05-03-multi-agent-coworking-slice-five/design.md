# Design

## Problem

By the end of Slice 4, Mission Control can show owned work and operator loops, but it still treats collaboration largely as transcript noise. Delegation, teammate waiting state, review-needed execution state, budget pressure, and recovery activity exist in runtime signals, yet the user cannot reliably see those signals as mission-linked coworking state.

At the same time, Slice 5 must avoid inventing a second durable work model or overclaiming external team support. The first slice should make local coworking legible while keeping attribution strict and projection-only.

## Goals

- add a mission-linked local coworking projection
- make participants, handoffs, blocked-on-approval, waiting-on-teammate, budget, and recovery visible in mission context
- keep collaboration state derived from real local runtime sources
- keep external P2P team state secondary in the first slice

## Non-Goals

- adding a new durable collaboration database table
- exposing full external P2P team UX as a primary Mission Control surface
- adding cockpit controls for role editing, team formation, or conflict resolution
- treating session-level runtime signals as mission-linked when attribution is ambiguous

## Collaboration Model

Slice 5 adds a collaboration projection attached to a mission rather than a new durable model.

The first slice needs, at minimum:

- `mission_id`
- participant summary
- active owner or current collaborating agent summary
- recent handoff edges
- collaboration state
- budget signal
- blocked-on summary
- recovery signal
- `updated_at`

Suggested collaboration states in this slice:

- `solo`
- `delegating`
- `waiting_on_teammate`
- `reviewing`
- `blocked_on_approval`
- `recovering`

This projection is derived and transient. It does not replace `Mission` or create a second durable work graph.

## First-Slice Source Contract

The first slice uses only real local sources already present in the codebase:

- mission execution links
- linked `AgentRun` runtime state
- mission-attributable delegation signals
- mission-attributable budget and recovery signals
- linked local execution review state

External P2P team coordination remains explicitly secondary in this slice.

## Attribution Rules

Strict mission-linked attribution is the central rule of Slice 5.

### Delegation and Handoff Attribution

Delegation signals may only appear on a mission when they can be tied to a mission-linked local execution.

If delegation is only visible at the broader session level and mission attribution is ambiguous:

- the signal remains session-level runtime context
- Mission Control SHALL NOT project it as mission coworking state

### Budget and Recovery Attribution

Budget warnings and recovery actions may only appear on a mission when they can be attributed through that mission's linked active local execution.

Otherwise:

- they remain runtime context
- they SHALL NOT be projected as mission collaboration state

### Review Attribution

The first slice may show `reviewing` only from linked local execution state such as `verify_pending` or equivalent orchestrator-review-needed status.

The first slice SHALL NOT infer review state from generic multi-agent activity alone.

## Mission Control Behavior

Slice 5 extends mission rows and mission detail context without replacing the mission board.

The first slice should add compact coworking context such as:

- participant summary
- recent handoff summary
- collaboration-state label
- budget or recovery hint when attributable

The surface should remain terminal-friendly and additive.

## External Team Boundary

External P2P team state already exists in the repo, but Slice 5 keeps it secondary.

The first slice SHALL:

- prioritize local built-in teammate collaboration signals
- avoid implying unified external team UX in Mission Control
- leave external team adapters for later work

## Validation

This change is valid only if:

- coworking is mission-linked and local-first
- collaboration remains a projection rather than a new durable model
- local handoff, participant, blocked, budget, and recovery visibility are captured
- mission attribution rules remain strict
- external P2P team UX remains secondary in the first slice
