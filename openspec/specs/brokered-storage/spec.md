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

### Requirement: Runtime readers use broker-backed storage capabilities
Runtime app and CLI reader paths MUST be able to obtain learning history, pending inquiries, workflow state, alert history, and reputation details through broker-backed storage capabilities.

#### Scenario: Reader path resolved through broker capability
- **WHEN** a runtime app or CLI path needs one of those reader surfaces while broker mode is active
- **THEN** it can obtain the data through broker-backed storage capabilities without opening or querying the application DB directly in the parent process

### Requirement: Payment mutators use storage-facing transaction capabilities
Production payment and settlement mutator paths MUST obtain payment transaction persistence and spending-limit collaborators through explicit storage-facing capabilities instead of extracting raw parent-side Ent clients.

#### Scenario: Payment CLI mutator setup avoids raw Ent access
- **WHEN** a payment CLI command initializes send, balance, or info dependencies
- **THEN** it obtains transaction persistence and spending-limit collaborators through storage-facing capabilities
- **AND** it does not extract a raw `*ent.Client` from the session store

#### Scenario: App payment and settlement setup avoids raw Ent access
- **WHEN** app wiring initializes payment service or P2P settlement persistence
- **THEN** it uses explicit storage-facing transaction capabilities
- **AND** it does not reconstruct those dependencies from raw parent-side ORM handles

### Requirement: Broker transport regression tests remain release gates
The broker storage boundary MUST continue to verify transport correctness for concurrent writes, large JSON responses, and graceful shutdown as part of standard test execution.

#### Scenario: Broker transport regressions are exercised in standard test runs
- **WHEN** `go test ./...` is executed
- **THEN** broker transport regression tests cover concurrent request serialization, responses larger than 64 KiB, and shutdown-before-close behavior

### Requirement: Payment mutator setup no longer falls back to direct Ent extraction
Production payment setup MUST use storage-facing transaction persistence and limiter capabilities without falling back to `session.EntStore.Client()`.

#### Scenario: Payment setup resolved through storage capability only
- **WHEN** app or CLI payment setup initializes payment transaction persistence
- **THEN** it resolves the collaborator through storage-facing capabilities
- **AND** it does not rebuild the dependency from a session-store Ent client

