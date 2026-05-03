# Design

## Problem

Mission Control already knows about durable missions, transient proposals, live decisions, and runtime activity. That is enough to steer individual work, but not enough to answer the higher-level operator question:

> "What open loops does Lango think still need my attention?"

Without a deterministic loop layer, the user must manually infer which blocked missions, stale inquiries, failed automation, or retryable dead letters are still part of the active agenda.

## Goals

- add an operator-facing loop projection over real current sources
- keep durable missions as the primary owned work surface
- add deterministic agenda ordering
- restrict the first slice to sources that already exist in code
- define follow-up generation only through explicit deterministic predicates

## Non-Goals

- full calendar integration
- inbox or email-thread integration
- external task-system integration
- workflow-run scheduled loops without a dedicated adapter
- a new durable loop table
- vague productivity or life-planner heuristics

## First-Slice Source Contract

Wave 4 first slice uses only real existing sources:

- durable missions
- pending librarian inquiries
- dead-letter and retry backlog
- deterministic follow-up signals derived from current mission/proposal/session facts
- cron-job schedule state

Explicitly excluded in this slice:

- workflow-run loops
- calendar events
- inbox threads
- third-party tasks or reminders

## Loop Projection Model

The first slice introduces a transient loop projection rather than new durable storage.

Suggested loop kinds:

- `mission_cluster`
- `inquiry`
- `scheduled_automation`
- `dead_letter`
- `follow_up`

Suggested statuses:

- `waiting_user`
- `blocked`
- `active`
- `scheduled`
- `needs_review`
- `resolved`

Each loop row is derived from real source records and must remain traceable back to those records.

## Deterministic Agenda Ordering

Agenda order in the first slice is fixed:

1. `waiting_user`
2. `blocked`
3. `active`
4. `scheduled`
5. `needs_review`
6. `resolved`

Within the same category:

- newer `updated_at` first

No other heuristic priority signal is allowed in this slice.

## Scheduled Automation

Scheduled automation loops are cron-backed only in this slice.

- a cron job may produce a loop when it is active, recently failed, or otherwise needs attention
- workflow runs remain deferred until a dedicated adapter exists

Wave 4 must not imply that any generic scheduled or workflow-owned loop model is already live.

## Follow-Up Predicates

Follow-up rows in the first slice are allowed only from explicit deterministic predicates:

- accepted proposal with no active linked execution after `10m`
- mission in `done` state updated within `24h` and still needing review
- pending inquiry older than `24h`
- cron job whose most recent execution failed within `24h`
- dead-letter entry that remains retryable

These thresholds are fixed constants in the first slice and must be test-controlled through an injectable clock.

## Mission Control Contract

Wave 4 extends Mission Control rather than replacing it.

- durable missions remain the primary owned work surface
- loop rows are additive coordination surfaces
- the first slice should prefer a compact loop lane or agenda band rather than another major shell rewrite

The slice is successful when Mission Control can surface real open loops honestly without implying unsupported integrations.
