## Purpose

Capability spec for companion-discovery. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Bonjour Service Discovery
The system SHALL discover companion apps on local network using mDNS/Bonjour.

#### Scenario: Discover companion on startup
- **WHEN** Lango server starts
- **THEN** it SHALL browse for services of type `_lango-companion._tcp`
- **AND** attempt connection to discovered instances

#### Scenario: Companion appears after startup
- **WHEN** a companion app starts and advertises its service
- **THEN** Lango SHALL detect the new service within 5 seconds
- **AND** initiate WebSocket connection

### Requirement: Companion Connection
The system SHALL maintain WebSocket connection to companion app.

#### Scenario: Successful connection
- **WHEN** companion is discovered
- **THEN** Lango connects via WebSocket to the advertised address
- **AND** performs mTLS handshake

#### Scenario: Connection lost
- **WHEN** WebSocket connection to companion is lost
- **THEN** the system continues operating with local fallback
- **AND** attempts reconnection every 30 seconds

### Requirement: Manual Companion Configuration
The system SHALL support manual companion address configuration.

#### Scenario: Configure via config file
- **WHEN** lango.json contains `security.companion.address`
- **THEN** the system connects to that address instead of using discovery
- **AND** skips Bonjour browsing

### Requirement: Companion Status Reporting
The system SHALL report companion connection status.

#### Scenario: Query companion status
- **WHEN** status is queried via gateway API
- **THEN** the system returns connection state, companion address, and last heartbeat time
