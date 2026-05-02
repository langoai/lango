# mission-control-tui Specification

## Purpose
TBD - created by archiving change 2026-05-02-mission-control-wave-one. Update Purpose after archive.
## Requirements
### Requirement: Mission Control is the default `lango` TUI surface
Running `lango` on an interactive terminal SHALL open Mission Control as the default cockpit surface. The page SHALL make ongoing work, the latest live decision, recent activity, and the shared composer path available without requiring the user to navigate to another page first. `lango chat` SHALL remain the direct focused-chat fallback.

#### Scenario: Bare `lango` enters Mission Control
- **WHEN** the user runs `lango` on an interactive terminal
- **THEN** the initial active cockpit page SHALL be Mission Control
- **AND** the first screen SHALL include a short hint that chat remains available through typing and through `lango chat`

#### Scenario: `lango chat` remains direct chat fallback
- **WHEN** the user runs `lango chat`
- **THEN** the application SHALL bypass Mission Control and start the focused chat surface directly

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

Mission Control SHALL continue rendering learning suggestions as transient `proposed` missions, but Wave 2 SHALL treat proposal acceptance as a real durable write path. Durable mission rows SHALL begin at `prepared` or `active`; the transient `proposed` state itself SHALL NOT be persisted as a durable mission row.

#### Scenario: Proposal acceptance creates the first durable mission row
- **WHEN** the user accepts a proposed learning-based mission from Mission Control
- **THEN** the application SHALL create a durable mission row with a new `mission_id`
- **AND** the first durable mission status SHALL be `prepared` or `active`
- **AND** the accepted proposal SHALL stop being only a transient overlay

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

