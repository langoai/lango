## ADDED Requirements

### Requirement: Storage broker subprocess mode
The system SHALL provide an internal storage broker subprocess mode inside the main `lango` binary. The broker SHALL be activated by a dedicated internal flag and SHALL not expose a separate user-facing binary.

#### Scenario: Broker mode selected
- **WHEN** the current process is launched with the storage broker flag
- **THEN** the process SHALL run the storage broker entry point instead of the normal CLI command tree

### Requirement: Persistent stdio JSON protocol
The storage broker SHALL communicate over a persistent stdio JSON protocol using request/response envelopes with `id`, `method`, `deadline_ms`, and `payload` fields on requests and `id`, `ok`, `result`, and `error` fields on responses.

#### Scenario: Health round-trip
- **WHEN** a client sends a `health` request
- **THEN** the broker SHALL return a successful response containing whether the broker currently owns an open application database

#### Scenario: Unknown method
- **WHEN** a client sends a request for an unknown method name
- **THEN** the broker SHALL return a failed response with an error string

### Requirement: Broker-owned open-db handshake
The storage broker SHALL expose an `open_db` request that opens the application SQLite database in read-write mode and prepares it for use before the parent runtime proceeds.

#### Scenario: Open database through broker
- **WHEN** the broker receives a valid `open_db` request with a database path
- **THEN** the broker SHALL open the SQLite database
- **AND** run schema migration and auxiliary table/index initialization before reporting success

#### Scenario: Missing database path
- **WHEN** the broker receives an `open_db` request without a database path
- **THEN** the broker SHALL return a failed response
