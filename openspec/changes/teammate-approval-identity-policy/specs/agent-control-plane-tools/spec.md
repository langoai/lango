## ADDED Requirements

### Requirement: Approval identity is exposed consistently
The control-plane blocked-state surface for built-in teammate runs SHALL expose approval request identity consistently with the capability-layer policy.

#### Scenario: Repeated blocked requests remain interpretable
- **WHEN** `agent_wait` reports an approval-blocked teammate run
- **THEN** the surfaced `grant_request_id` SHALL identify the stable logical blocked request
- **AND** any distinction between repeated attempts of that same request SHALL come from separate attempt metadata if exposed
