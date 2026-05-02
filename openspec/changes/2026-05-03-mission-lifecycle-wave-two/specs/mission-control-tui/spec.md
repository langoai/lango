## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Waiting for user direction is stored as coarse durable mission state

Wave 2 SHALL represent decision-paused mission progress as coarse durable `waiting_decision` mission state. This durable state may include latest decision kind or summary, but it SHALL NOT become a durable approval queue or require Mission Control to persist every live approval request as a durable decision item.

#### Scenario: Mission pauses on user direction without durable queue semantics
- **WHEN** mission progress is paused pending user approval or direction
- **THEN** the durable mission row SHALL move into `waiting_decision`
- **AND** the durable state MAY store coarse latest-decision summary fields
- **BUT** Wave 2 SHALL NOT require a durable per-request approval queue for Mission Control
