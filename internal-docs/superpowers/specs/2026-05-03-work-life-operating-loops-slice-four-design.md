# Work And Life Operating Loops Slice 4 Design

## Strategic Goal

Slice 3 introduced proactive proposals. Slice 4 should turn Mission Control from a mission-and-proposal surface into a broader operating surface for recurring work, open loops, follow-ups, and scheduled activity.

The goal is not to pretend Lango already has universal calendar, email, or personal-life integrations. The goal is to make existing system state feel like an operator-grade loop manager:

- what is active now
- what is waiting on the user
- what is scheduled
- what is unresolved
- what should come back onto the agenda next

## Product Intent

Slice 4 should answer:

> "What ongoing loops does Lango think I am responsible for, and what needs attention next?"

This is a level above single missions. Some loops are:

- one durable mission
- a cluster of related missions
- a scheduled automation with unresolved follow-up
- a knowledge inquiry waiting for user input
- a dead-letter or retry backlog item

## Fixed Decisions

The following decisions are fixed for Slice 4:

- Slice 4 builds on durable missions and transient proposals; it does not replace them
- loops are a presentation and coordination model first, not a replacement for mission persistence
- the first slice must use real existing producers only
- unsupported domains such as calendar sync, inbox sync, or document collaboration remain explicit future work unless a real adapter exists
- loop state should remain legible in terminal form and must not become a vague productivity dashboard

## Why Loops Are Needed

By the end of Slice 3, Lango can:

- show durable missions
- show transient proposals
- surface live decisions

But the user still has to infer larger operational groupings manually. A mission system alone does not tell the user:

- which things are recurring
- which unresolved items keep resurfacing
- which scheduled systems are waiting on human follow-up
- which partial failures should stay on the agenda

Slice 4 should make those loops explicit.

## Initial Loop Sources

Slice 4 should start only from sources that already exist in code today.

### Source 1: Durable Missions

Durable missions remain the primary owned work units.

Slice 4 loop view should be able to group:

- active missions
- blocked missions
- waiting-decision missions
- recently completed but still relevant missions

### Source 2: Pending Knowledge Inquiries

The librarian already persists pending inquiries. These are legitimate open loops because they represent missing user input required to improve future system behavior or knowledge.

Initial loop role:

- unresolved inquiry
- pending answer from user

### Source 3: Scheduled Automation

Existing cron and workflow systems already model recurring or staged work.

Initial loop role:

- recurring automation
- workflow with recent activity
- scheduled work with failure or follow-up need

### Source 4: Dead Letters / Retry Backlog

The dead-letter and retry surfaces already exist. They are operational open loops by definition.

Initial loop role:

- unresolved failure requiring manual replay or retry
- blocked settlement/execution loop

### Source 5: Session Follow-Up Signals

Current sessions, runtime activity, and recent mission/proposal state can produce lightweight follow-up signals such as:

- mission recently completed but needs review
- proposal accepted but not yet started
- user-submitted mission with no linked execution yet

This should remain deterministic and conservative.

## Explicit Non-Sources For The First Slice

The first Slice 4 slice must not pretend the following are real loop sources unless adapters exist:

- calendar events
- inbox/email threads
- chat apps beyond the already modeled channel delivery surfaces
- personal reminders from outside Lango
- external docs/tasks from third-party systems

These can become future sources later, but not in the initial implementation.

## Loop Model

Slice 4 introduces a transient `LoopView` model.

Suggested fields:

- `loop_id`
- `session_key`
- `loop_kind`
- `title`
- `summary`
- `status`
- `priority`
- `source_refs`
- `next_action`
- `updated_at`

Suggested loop kinds:

- `mission_cluster`
- `inquiry`
- `scheduled_automation`
- `dead_letter`
- `follow_up`

Suggested statuses:

- `active`
- `waiting_user`
- `scheduled`
- `blocked`
- `needs_review`
- `resolved`

This does not need a new durable database table in the first slice. It can begin as a deterministic app-side projection built from existing durable and transient sources.

## Relationship To Missions

Loops are not a replacement for missions.

Rules:

- a durable mission can appear inside one loop
- multiple related durable missions can be grouped into one loop later, but the first slice may keep grouping simple
- proposals remain transient suggestions and should only appear in loop surfaces when they clearly contribute to an existing open loop

Slice 4 should not create a second durable mission-like table that competes with `Mission`.

## Initial Grouping Rules

The first slice should stay simple and deterministic.

### Durable Mission Loops

- each active/blocked/waiting-decision mission can appear as its own loop entry
- later grouping across multiple missions is explicitly future work

### Inquiry Loops

- each pending inquiry is its own loop entry

### Scheduled Automation Loops

- one cron job or workflow may produce one loop entry if it is currently active, recently failed, or waiting on attention

### Dead Letter Loops

- one dead-letter execution or retry backlog item is one loop entry

### Follow-Up Loops

- generated only from deterministic recent mission/proposal/session state
- no speculative heuristics such as “this feels important”

## Mission Control Surface

Slice 4 should extend Mission Control with a loop-aware overview rather than replacing the mission board entirely.

Possible shape:

- top summary remains missions + decisions
- secondary band or toggle shows loops / agenda
- loops should be scannable in a narrow terminal

The first slice should avoid overhauling the whole layout again. A toggle or compact loop lane is enough.

## Loop Prioritization

Slice 4 needs deterministic prioritization.

Suggested order:

1. waiting-user loops
2. blocked loops
3. active loops
4. scheduled loops
5. needs-review loops
6. resolved loops

Within the same category:

- newer updates first
- explicit mission priority may override later, but not in the first slice

## Agenda Semantics

Slice 4 should introduce an agenda concept carefully.

The first slice may define `agenda` as:

- the ordered list of highest-priority unresolved loops for the current session/operator surface

This is not yet a full personal planner.

It is simply the top of the loop stack that deserves attention now.

## Follow-Up Semantics

Follow-up signals should be generated only from real facts such as:

- accepted proposal with no active execution yet
- completed mission with unresolved review step
- failed or blocked recurring automation
- unresolved inquiry older than a threshold

The first slice should not generate arbitrary narrative follow-ups.

## Architecture Sketch

Suggested components:

- `LoopProjector`
  - reads durable missions
  - reads proposal registry
  - reads inquiry store
  - reads cron/workflow/dead-letter sources
  - produces deterministic loop view rows

- `AgendaProjector`
  - orders the highest-priority unresolved loops

- `LoopSourceAdapters`
  - one adapter per source category

The first slice should keep these in the application/UI boundary rather than introducing a new durable domain model immediately.

## Relationship To Slice 5

Slice 4 should stop short of multi-agent coworking. However, loop surfaces should leave room for later visibility such as:

- which agent last touched the loop
- whether review or handoff is pending
- whether a loop is blocked on another agent

That belongs to Slice 5.

## Non-Goals

Slice 4 does not aim to:

- implement full calendar or inbox integration
- replace durable missions with loops
- build a general life-planning ontology
- add broad heuristic prioritization
- add multi-agent coworking views yet

## Success Criteria

Slice 4 is successful when:

- Mission Control can surface real open loops from existing sources
- agenda ordering is deterministic and honest
- scheduled automation, inquiries, dead letters, and durable missions can all appear as operator-visible loops
- no fake integrations or speculative sources are implied
