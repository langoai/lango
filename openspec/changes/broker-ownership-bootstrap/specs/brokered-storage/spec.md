## ADDED Requirements

### Requirement: Broker exposes bootstrap config/profile operations
The storage broker MUST expose config profile operations needed by bootstrap so profile loading can proceed without requiring direct ent/sql access in the parent process.

#### Scenario: Bootstrap profile load through broker
- **WHEN** bootstrap needs to load or resolve the active config profile
- **THEN** it can do so through broker-backed config/profile RPCs

### Requirement: Broker exposes session bootstrap operations
The storage broker MUST expose session-store operations needed for bootstrap/runtime session wiring.

#### Scenario: Session store opened through broker-backed adapter
- **WHEN** runtime wiring requests a session store while broker mode is enabled
- **THEN** the session store can be constructed from broker-backed operations instead of direct parent DB access
