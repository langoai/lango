## ADDED Requirements

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
