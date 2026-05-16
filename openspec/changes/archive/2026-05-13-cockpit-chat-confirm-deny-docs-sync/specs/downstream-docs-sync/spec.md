## MODIFIED Requirements
### Requirement: Cockpit feature docs describe the Chat operator surface
The public cockpit feature reference SHALL describe the current Chat page beyond simple roster availability.

#### Scenario: Cockpit feature page describes confirm-pending deny path
- **WHEN** the public cockpit feature reference describes the critical-risk double-press guardrail
- **THEN** it SHALL explain that `d` or `Esc` still deny the request immediately while confirm-pending is active
