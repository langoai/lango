## ADDED Requirements

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
Mission Control SHALL derive active mission rows from existing runtime producers without introducing durable mission entities. Background tasks are the required Wave 1 source. Optional RunLedger and AgentRun readers may enrich display fields when available.

#### Scenario: Background task renders as active mission
- **WHEN** a background task snapshot is available
- **THEN** Mission Control SHALL render an `active` mission row using a deterministic title derived from the task prompt
- **AND** the row SHALL use task-derived status and updated time instead of synthetic mission persistence fields

#### Scenario: Optional readers enrich without becoming required
- **WHEN** RunLedger or AgentRun readers are available
- **THEN** Mission Control MAY enrich owner, blocked-state, or next-action fields
- **AND** their absence SHALL NOT prevent background-task missions from rendering

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
Mission Control SHALL project buffered learning suggestions as `proposed` missions for inspection, acceptance, or dismissal. In Wave 1, this is an actionable UI proposal that MAY route into an existing prompt or approval flow, but it SHALL NOT imply durable mission creation or automatic learning persistence by itself.

#### Scenario: Learning suggestion becomes proposed mission
- **WHEN** a `LearningSuggestionEvent` is buffered for Mission Control
- **THEN** the page SHALL render a `proposed` mission row with a deterministic title beginning with `Apply learning rule`

#### Scenario: Proposal acceptance may route into an existing flow without creating durable state
- **WHEN** the user accepts a proposed learning-suggestion mission
- **THEN** the UI MAY route the intent into an existing prompt or approval flow
- **BUT** Wave 1 SHALL NOT claim that a durable mission object was created
- **AND** Wave 1 SHALL NOT claim that learning persistence happened unless an existing downstream flow actually performs it

### Requirement: Mission Control presents timeline and header as first-class Wave 1 outputs
Mission Control SHALL expose a recent activity timeline and a compact header summary as part of the default first screen, using only cockpit-available data.

#### Scenario: Timeline shows recent deterministic activity
- **WHEN** runtime, approval, or learning events are available to cockpit-owned buffers
- **THEN** Mission Control SHALL render recent activity entries using deterministic text
- **AND** the timeline SHALL behave as a recent activity surface rather than a message-by-message transcript promise

#### Scenario: Header shows compact available status
- **WHEN** cockpit already has compact runtime status such as active-agent summary, pending-decision count, or degraded-reader notice
- **THEN** Mission Control SHALL surface that status in the header
- **AND** it SHALL NOT require new synthesized context beyond data already available to cockpit

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
