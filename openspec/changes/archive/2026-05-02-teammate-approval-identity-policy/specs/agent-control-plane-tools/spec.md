## ADDED Requirements

### Requirement: Approval identity is exposed consistently
The control-plane blocked-state surface for built-in teammate runs SHALL expose stable logical approval identity together with attempt metadata.

#### Scenario: agent_wait exposes logical identity and attempt metadata
- **WHEN** `agent_wait` reports an approval-blocked teammate run
- **THEN** the response SHALL include `grant_request_id`
- **AND** the response SHALL expose `grant_attempt`
- **AND** the response SHALL expose `grant_state`
- **AND** `grant_attempt` SHALL be at least `1` whenever the run is currently `blocked_waiting_approval`
