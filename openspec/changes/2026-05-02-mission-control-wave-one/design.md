# Design

## Problem

The cockpit already exposes chat, tasks, approvals, sessions, status, and other operational pages, but the default experience still makes users hunt for current work and the latest pending decision. At the same time, the runtime does not yet have a durable mission domain model, so the UI cannot honestly pretend that one already exists.

The design must therefore deliver a mission-first surface using only existing runtime facts plus narrowly-scoped producer hardening.

## Goals

- make Mission Control the default `lango` surface
- keep `lango chat` as a direct fallback
- project active missions, proposed missions, the latest live decision, and recent activity from existing data
- keep approvals on the existing pending response path
- keep the implementation cockpit-owned and session-scoped

## Non-Goals

- durable mission persistence or mission IDs backed by a new domain engine
- general agent-authored proposed mission creation beyond current learning suggestions
- LLM-written activity summaries
- message-level transcript streaming outside the chat model
- page-level EventBus subscription teardown

## Product Shape

Wave 1 adds a Mission Control page that becomes the default first screen for `lango`. The page renders:

- current missions derived primarily from background tasks and optional runtime readers
- the latest live decision derived from the latest pending approval request
- proposed missions derived from learning suggestion events
- a deterministic activity timeline plus the shared chat composer path
- a narrow header summary from runtime state already available to cockpit

Existing cockpit detail pages remain reachable. They are no longer the default mental model.

## Projection Model

Mission Control is a UI projection, not a mission engine.

### Missions

- background tasks project to active missions
- optional RunLedger and AgentRun readers may enrich status, blockers, and owner summaries
- learning suggestions project to proposed missions only as UI proposals
- accepting a proposed mission may route to existing prompt or approval flows, but does not create durable mission state in Wave 1

### Decision

- the latest pending approval request is the only required live decision in Wave 1
- approval history remains history-only data
- active approvals must resolve through the original pending response channel

### Activity

- timeline entries use deterministic, rule-based text
- retention is bounded to the latest 200 items
- switching pages does not clear the timeline
- session end resets the timeline

## Ownership Boundaries

Cockpit owns the shared session state required by Mission Control:

- `PendingApprovalRegistry` holds the latest pending approval request and its response channel
- `LearningSuggestionBuffer` holds recent suggestion events for projection
- `MissionActivityBuffer` holds recent deterministic activity items
- cockpit-lifetime subscriptions populate those stores

The chat model remains the direct conversation surface, but when mounted inside cockpit it must cooperate with the shared pending approval owner instead of independently owning pending approval state.

## Routing Decisions

- bare `lango` opens Mission Control first
- `lango chat` still opens focused chat directly
- existing cockpit pages remain available as detail routes
- Wave 1 should minimize shortcut churn; existing global page shortcuts remain stable, while Mission Control becomes the default route and sidebar/page-selection target

## Responsive Behavior

Mission Control must define responsive behavior before code lands:

- width `>= 120`: missions and the latest live decision render side by side above the timeline/composer region
- width `80-119`: missions render first, then the latest live decision in a compact stacked section
- width `< 80`: one focused lane renders at a time and `Tab` cycles lanes
- height `< 24`: the page prioritizes header, focused lane, and footer; the composer path stays available but the input opens inline on typing intent instead of remaining persistently visible

## Failure and Degraded States

Wave 1 distinguishes:

- `loading`: before the first projector snapshot is ready
- `empty`: no missions and no pending live decision after data load
- `degraded`: optional readers such as RunLedger or AgentRun are unavailable

Degraded state must omit unavailable details rather than inventing placeholder runtime facts.

## Risks

- if cockpit and chat each own separate pending approval state, the latest approval can diverge across surfaces
- if Mission Control claims mission persistence semantics too early, the UI will overpromise behavior the runtime cannot sustain
- if keyboard routing is changed too aggressively, operators lose established cockpit muscle memory

## Validation

The change is valid only if the OpenSpec delta stays inside Wave 1:

- Mission Control is the default surface, but not a new mission engine
- live approvals use the shared pending channel path
- learning suggestions stay informational until a later mission lifecycle slice exists
- no spec text implies durable mission storage, transcript event streaming, or page-lifetime unsubscribe support
