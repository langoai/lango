# brokered-storage Specification

## Purpose
TBD - created by archiving change brokered-storage-boundary. Update Purpose after archive.
## Requirements
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

### Requirement: Production code uses capability-specific storage access
Production app and CLI code MUST consume storage through capability-specific facade methods instead of generic Ent/SQL handle access.

#### Scenario: CLI storage readers do not use generic ent accessors
- **WHEN** learning history, librarian inquiry inspection, workflow state, payment setup, or reputation inspection runs through the CLI
- **THEN** those code paths use storage-provided readers or factories
- **AND** they do not call generic `EntClient()` accessors from production code

#### Scenario: App wiring uses facade dependency bundles
- **WHEN** app initialization wires ontology, observability alerts, workflow state, or P2P reputation/settlement components
- **THEN** it resolves those dependencies from facade capability methods
- **AND** it does not reconstruct them from generic production ent/sql handles

### Requirement: Broker client serializes request writes
The storage broker client MUST serialize request writes on the shared stdin pipe so concurrent RPC calls cannot interleave request bytes.

#### Scenario: Concurrent RPC calls do not corrupt request framing
- **WHEN** multiple goroutines issue broker RPCs concurrently
- **THEN** each request is written atomically with respect to other requests on the shared stdin pipe

### Requirement: Broker client supports large JSON responses
The storage broker client MUST support JSON response frames larger than 64 KiB for payload protection traffic.

#### Scenario: Large encrypt/decrypt response is decoded successfully
- **WHEN** the broker emits a response whose JSON line exceeds 64 KiB
- **THEN** the client read loop continues decoding responses successfully

### Requirement: Broker client sends graceful shutdown before closing transport
The storage broker client MUST attempt the shutdown RPC path before it marks itself closed and tears down stdio.

#### Scenario: Close triggers shutdown RPC
- **WHEN** `Client.Close()` is called on an open broker client
- **THEN** the client attempts `methodShutdown`
- **AND** only after that begins transport teardown and process wait

