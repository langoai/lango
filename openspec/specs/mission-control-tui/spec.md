# mission-control-tui Specification

## Purpose
TBD - created by archiving change 2026-05-02-mission-control-wave-one. Update Purpose after archive.
## Requirements
### Requirement: Mission Control is the default `lango` TUI surface

Mission Control SHALL remain the shared mission-native surface used by the interactive workbench and available inside cockpit. It SHALL continue to make ongoing work, the latest live decision, recent activity, and the shared composer path available without requiring the user to navigate to another workflow first.

#### Scenario: Workbench mounts Mission Control directly
- **WHEN** the user runs bare `lango` on an interactive terminal
- **THEN** the workbench SHALL mount Mission Control as its primary body
- **AND** the first screen SHALL include a short hint that `lango chat` remains available as focused chat
- **AND** the first Wave 6 slice SHALL also hint that `lango cockpit` remains available as the explicit dashboard

#### Scenario: `lango chat` remains direct chat fallback
- **WHEN** the user runs `lango chat`
- **THEN** the application SHALL bypass Mission Control and start the focused chat surface directly

#### Scenario: Cockpit still exposes Mission Control as a page
- **WHEN** the user runs `lango cockpit`
- **THEN** Mission Control SHALL remain available through the cockpit page set
- **AND** Wave 6 SHALL NOT require a second Mission Control domain or projection system

### Requirement: Active missions are projected deterministically from existing runtime facts

Mission Control SHALL use durable mission rows as the primary read source for session work once Wave 2 mission persistence is available. Runtime producers such as background tasks, live approvals, learning suggestions, and optional execution readers SHALL remain available as overlays for unmatched or not-yet-linked work so the page does not hide active runtime state.

#### Scenario: Durable mission row renders as the primary mission record
- **WHEN** a durable mission row exists for the current session
- **THEN** Mission Control SHALL render that mission from the durable mission record instead of synthesizing the row only from background task state
- **AND** the row SHALL use the durable `mission_id` as its stable identity

#### Scenario: Unmatched runtime work remains visible beside durable rows
- **WHEN** runtime work exists that is not linked to any durable mission row
- **THEN** Mission Control SHALL continue to surface that work as unmatched runtime overlay content
- **AND** the page SHALL retain that visibility until the work is linked to a durable mission or dismissed by the product flow

### Requirement: The latest pending approval renders as one live decision
Mission Control SHALL render the latest pending approval request as one live decision using the shared pending approval owner. Resolving the decision SHALL write to the original approval response channel and remove the pending item from other cockpit surfaces on the next render.

#### Scenario: Pending approval appears as live decision
- **WHEN** cockpit receives a pending `ApprovalRequestMsg`
- **THEN** Mission Control SHALL render a live decision row showing the requested action, reason, effect summary, and risk
- **AND** Wave 1 SHALL NOT require Mission Control to queue or render multiple simultaneous pending approvals

#### Scenario: Second pending approval arrives while one is already visible
- **WHEN** a second pending `ApprovalRequestMsg` arrives before the visible pending approval is resolved
- **THEN** Wave 1 SHALL continue to promise only one visible live decision at a time
- **AND** this change SHALL NOT promise multi-pending queueing, ordering, or concurrent rendering behavior

#### Scenario: Decision resolution uses the shared pending path
- **WHEN** the user approves, denies, or allows for session from Mission Control
- **THEN** the response SHALL be sent through the original pending approval response channel
- **AND** the same pending request SHALL disappear from chat-derived or approvals-derived cockpit surfaces on the next render

### Requirement: Learning suggestions render as actionable proposed missions

Mission Control SHALL back proposed missions with a transient proposal registry instead of rendering raw learning-buffer rows directly. In this Wave 3 slice, `LearningSuggestionEvent` is the only active proposal producer. Proposed missions SHALL remain transient and SHALL move through explicit proposal states such as `suggested`, `preparing`, and `prepared` before acceptance or dismissal.

#### Scenario: Learning suggestion becomes a transient proposal record
- **WHEN** a `LearningSuggestionEvent` is emitted for the current session
- **THEN** Mission Control SHALL create or update one transient proposal record using the suggestion as the proposal source
- **AND** the proposal SHALL NOT create a durable `Mission` row before user acceptance

#### Scenario: Proposed mission no longer depends on raw learning-buffer row shape
- **WHEN** Mission Control renders active proposals
- **THEN** the page SHALL read proposal state from the transient proposal registry
- **AND** the learning buffer MAY remain only as a producer input or compatibility source, not the primary rendered proposal state

### Requirement: Mission Control presents timeline and header as first-class Wave 1 outputs

Mission Control SHALL add a real direct mission-start write path in Wave 2 while preserving timeline and header behavior from Wave 1.

#### Scenario: Direct mission start creates durable mission state
- **WHEN** the user starts a mission directly from Mission Control
- **THEN** the application SHALL create a durable mission row immediately
- **AND** the resulting mission SHALL appear in Mission Control through the durable-first read path

### Requirement: Mission Control defines loading, empty, degraded, and responsive states
Mission Control SHALL distinguish first-load, empty-data, degraded-reader, and narrow-terminal states instead of rendering the same fallback for every case.

#### Scenario: Loading state before first snapshot
- **WHEN** Mission Control has not yet produced its first projector snapshot
- **THEN** the page SHALL render a loading view instead of empty-state copy

#### Scenario: Empty state after data load
- **WHEN** data has loaded and there are no missions and no pending live decision
- **THEN** the page SHALL render an empty-state view with the shared composer path still available on the page

#### Scenario: Degraded state omits unavailable optional fields
- **WHEN** optional readers such as RunLedger or AgentRun are unavailable
- **THEN** the page SHALL render a compact degraded note
- **AND** it SHALL omit unavailable mission or header fields instead of inventing values

#### Scenario: Narrow terminal collapses to focused lane
- **WHEN** terminal width is less than 80 columns
- **THEN** Mission Control SHALL render one focused lane at a time
- **AND** `Tab` SHALL cycle the focused lane while the footer reports the active lane and pending decision count

#### Scenario: Compact height opens composer inline
- **WHEN** terminal height is less than 24 rows
- **THEN** Mission Control SHALL prioritize header, focused lane, and footer
- **AND** the composer SHALL open inline on typing intent instead of remaining persistently visible

### Requirement: Waiting for user direction is stored as coarse durable mission state

Wave 2 SHALL represent decision-paused mission progress as coarse durable `waiting_decision` mission state. This durable state may include latest decision kind or summary, but it SHALL NOT become a durable approval queue or require Mission Control to persist every live approval request as a durable decision item. A mission SHALL only enter this state when the paused work can be deterministically attributed to that mission, and the durable mission row SHALL store at most one latest decision marker at a time.

#### Scenario: Mission pauses on user direction without durable queue semantics
- **WHEN** mission progress is paused pending user approval or direction
- **THEN** the durable mission row SHALL move into `waiting_decision`
- **AND** the durable state MAY store coarse latest-decision summary fields
- **BUT** Wave 2 SHALL NOT require a durable per-request approval queue for Mission Control

#### Scenario: Decision state requires deterministic mission attribution
- **WHEN** paused approval or direction work cannot be deterministically attributed to a durable mission
- **THEN** Mission Control MAY still show a live runtime decision surface
- **BUT** the system SHALL NOT invent a durable `waiting_decision` mission transition for an unrelated or ambiguous mission row

#### Scenario: Durable mission row stores only the latest coarse decision marker
- **WHEN** a mission is already in `waiting_decision`
- **AND** a later mission-attributed approval or direction update arrives
- **THEN** the durable mission row SHALL keep only one latest decision marker
- **AND** Wave 2 SHALL NOT require durable multi-pending approval semantics for that mission

### Requirement: Proposed missions can prepare a deterministic brief before acceptance

Mission Control SHALL support deterministic low-risk preparation for transient proposals in this slice. Preparation SHALL generate a prepared brief from source-native evidence already available on the proposal source and SHALL NOT require broad heuristic agent work.

#### Scenario: Proposal reaches prepared with deterministic source-native brief
- **WHEN** a learning-based proposal is prepared
- **THEN** the proposal SHALL move into `prepared`
- **AND** the prepared brief SHALL be derived from source event fields such as proposed rule, rationale, confidence, and other stable source metadata
- **AND** the slice SHALL NOT require an LLM call to generate that prepared brief

#### Scenario: Generic proposal-owned execution remains out of scope
- **WHEN** a proposal moves through `suggested`, `preparing`, or `prepared`
- **THEN** the slice SHALL NOT require launching generic proposal-owned background execution or generic proposal-owned `RunLedger` execution
- **AND** the prepared state SHALL still be satisfiable through deterministic brief generation alone

### Requirement: Accepting a prepared proposal preserves prepared context in the durable mission path

Accepting a prepared proposal SHALL create the first durable mission row while preserving the prepared brief context that justified the proposal. `proposed` itself SHALL remain transient; the durable mission row SHALL still begin at `prepared` or `active`.

#### Scenario: Prepared proposal acceptance creates durable mission without losing prepared brief
- **WHEN** the user accepts a prepared proposal from Mission Control
- **THEN** the application SHALL create a durable mission through the mission lifecycle write path
- **AND** the resulting durable mission SHALL begin at `prepared` or `active`
- **AND** the prepared brief context SHALL remain available on the resulting mission path instead of being discarded during acceptance

### Requirement: Non-learning proposal producers remain explicitly disabled in this slice

Wave 3 SHALL keep the first proactive slice narrowly scoped. Librarian-gap producers, runtime-failure producers, and other future proposal sources SHALL remain explicit non-goals until they have dedicated adapters and source contracts.

#### Scenario: Librarian and runtime-failure producers do not create proposals in this slice
- **WHEN** librarian gap/inquiry signals or runtime failure signals exist
- **THEN** the Wave 3 first slice SHALL NOT require those signals to create transient proposals
- **AND** Mission Control SHALL NOT imply that those producers are active before their dedicated adapters exist

### Requirement: Mission Control can project operator loops from real existing sources
Mission Control SHALL support an operator loop surface in addition to durable missions, proposals, and live decisions. In the first Wave 4 slice, loop rows SHALL be projected only from real existing sources rather than invented integrations or placeholder data.

#### Scenario: Loop rows are derived only from first-slice real sources
- **WHEN** Mission Control projects operator loops for the current session
- **THEN** it SHALL derive those loops only from real existing sources such as durable missions, pending inquiries, dead-letter or retry backlog, cron-job schedule state, and deterministic follow-up signals from current mission or proposal state
- **AND** the first slice SHALL NOT require fabricated calendar, inbox, or external task-system loop sources

#### Scenario: Loop surface does not replace durable missions
- **WHEN** loop rows are available for the current session
- **THEN** Mission Control SHALL keep durable missions visible as the primary owned work surface
- **AND** loop rows SHALL remain an additive coordination surface rather than a replacement for durable mission records

### Requirement: Mission Control agenda ordering is deterministic
Mission Control SHALL derive an agenda ordering for unresolved loops using deterministic category ordering rather than broad heuristic prioritization.

#### Scenario: Agenda order follows fixed Wave 4 priority
- **WHEN** multiple unresolved loop rows are visible in the same Mission Control session
- **THEN** the agenda SHALL order them by fixed category priority: `waiting_user`, `blocked`, `active`, `scheduled`, `needs_review`, `resolved`
- **AND** rows within the same category SHALL order newer updates first

### Requirement: Scheduled automation loops are cron-job only in the first slice
Wave 4 SHALL keep the first scheduled automation source narrow. Mission Control SHALL surface scheduled automation loops only from cron-job sources until another source has a dedicated adapter.

#### Scenario: Cron job can appear as scheduled automation loop
- **WHEN** a cron job source exists with active, failed, or attention-needing state relevant to the current session
- **THEN** Mission Control MAY project one scheduled automation loop from that cron source

#### Scenario: Workflow runs remain deferred without a dedicated adapter
- **WHEN** workflow-run state exists but no dedicated loop adapter has been introduced for Wave 4
- **THEN** Mission Control SHALL NOT imply workflow-run loops are part of the first scheduled automation slice

### Requirement: Follow-up loops use explicit deterministic predicates
Wave 4 follow-up loops SHALL be generated only from explicit deterministic predicates over existing source facts.

#### Scenario: Follow-up loop is generated from deterministic source fact
- **WHEN** one of the approved first-slice predicates holds, such as an accepted proposal with no active linked execution yet, a recently completed mission still needing review, a failed recurring cron automation, or an unresolved inquiry older than a threshold
- **THEN** Mission Control SHALL project a follow-up loop for that fact
- **AND** the projected loop SHALL be traceable to the underlying source state

#### Scenario: Speculative follow-up loops are not generated
- **WHEN** no approved deterministic predicate matches the current source state
- **THEN** Mission Control SHALL NOT invent a narrative or heuristic follow-up loop

### Requirement: Unsupported external work-life integrations remain explicit non-goals
The first Wave 4 slice SHALL remain honest about unavailable sources. Mission Control SHALL NOT imply support for external work-life operating loops when no real adapter exists.

#### Scenario: Calendar, inbox, and external task integrations stay disabled
- **WHEN** Mission Control renders the Wave 4 loop surface
- **THEN** it SHALL NOT claim calendar events, inbox threads, or third-party external task systems are first-slice loop sources unless a real adapter exists in the application
- **AND** those unsupported domains SHALL remain explicit non-goals for this change

### Requirement: Mission Control can project mission-linked local coworking context

Mission Control SHALL support a collaboration projection attached to durable missions. In the first Wave 5 slice, this projection SHALL be derived from mission-linked local coworking signals rather than a new durable collaboration model.

#### Scenario: Mission row can show local collaboration context
- **WHEN** a durable mission has attributable local coworking signals
- **THEN** Mission Control SHALL surface compact participant, handoff, blocked, budget, or recovery context on that mission row or mission detail view
- **AND** the collaboration context SHALL remain attached to the durable mission instead of becoming a separate durable work model

### Requirement: Collaboration attribution must be mission-linked and local-first

Wave 5 SHALL use strict mission-linked attribution for coworking signals. Session-level local runtime signals SHALL only appear on a mission when attribution is provable through that mission's linked local execution. External P2P team state SHALL remain secondary in the first slice.

#### Scenario: Attributable local handoff can appear on a mission
- **WHEN** a recent local delegation or teammate handoff can be attributed through a mission-linked local execution
- **THEN** Mission Control SHALL be allowed to project that handoff as mission collaboration context

#### Scenario: Session-level local noise is not over-attributed to a mission
- **WHEN** delegation, budget, or recovery signals exist only at the session level and no mission-linked local attribution is provable
- **THEN** Mission Control SHALL NOT project those signals as mission-specific coworking state

#### Scenario: External P2P team remains secondary
- **WHEN** Wave 5 first-slice collaboration context is rendered
- **THEN** local built-in teammate signals SHALL be the primary collaboration source
- **AND** Mission Control SHALL NOT imply a full external P2P team collaboration surface is part of this slice

### Requirement: Mission Control collaboration states are grounded in real local runtime signals

Wave 5 collaboration visibility SHALL be grounded in real local signals such as linked `AgentRun` state, local handoffs, linked review-needed execution state, and mission-attributed budget or recovery runtime signals.

#### Scenario: Blocked, budget, and recovery visibility come from real local sources
- **WHEN** a mission-linked local execution is blocked on approval, waiting on a teammate, under budget pressure, or recovering from a recent runtime action
- **THEN** Mission Control SHALL project those states in collaboration context
- **AND** the projected state SHALL remain traceable to the underlying local runtime source

#### Scenario: Reviewing state is not inferred from vague multi-agent activity
- **WHEN** a mission has generic multi-agent activity but no linked local review-needed execution signal
- **THEN** Mission Control SHALL NOT invent a `reviewing` collaboration state for that mission

