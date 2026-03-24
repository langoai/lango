## ADDED Requirements

### Requirement: Single-column cockpit regions
The interactive TUI SHALL render as a single-column coding-agent cockpit with four vertical regions: header, turn status strip, transcript viewport, and composer/help footer.

#### Scenario: Default idle layout
- **WHEN** the user runs `lango` and the TUI enters idle state
- **THEN** the screen SHALL show a header, a turn status strip, a transcript viewport, and a composer/help footer in that top-to-bottom order

#### Scenario: Streaming layout
- **WHEN** the agent is actively streaming a response
- **THEN** the same four regions SHALL remain visible and the turn status strip SHALL indicate that generation is in progress

### Requirement: Approval interrupt card
Approval requests SHALL be rendered as an interrupt card inside the single-column layout, visually tied to the active turn instead of opening a separate modal or panel.

#### Scenario: Approval request shown inline
- **WHEN** a tool approval request is raised during a turn
- **THEN** the TUI SHALL display an approval card in the main column with the tool name, summary, critical parameters, and action keys

#### Scenario: Approval resolution preserved in transcript
- **WHEN** an approval request is approved or denied
- **THEN** the transcript SHALL retain a short status entry describing the approval event outcome
