## ADDED Requirements

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
