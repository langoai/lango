## 1. Broker Runtime

- [x] 1.1 Add the internal broker mode and parent-side spawn/healthcheck/shutdown flow.
- [x] 1.2 Define the coarse-grained stdio JSON RPC envelope and fail-closed semantics.
- [x] 1.3 Move SQLite open, ent schema migration, and auxiliary table/index initialization into the broker open-db handshake.

## 2. Storage Facade Migration

- [x] 2.1 Introduce a storage facade that covers config/security/session/knowledge/memory/agent-memory/runledger/provenance/turntrace/audit domains.
- [x] 2.2 Change bootstrap to return the storage client instead of direct DB handles.
- [x] 2.3 Remove direct DB access from app wiring and CLI status/doctor/settings/config helpers.

## 3. Boundary Enforcement

- [x] 3.1 Recompute protected paths from resolved runtime DB/envelope/keyfile/graph paths.
- [x] 3.2 Enforce close-on-exec and no-FD-inheritance rules for non-broker subprocesses.
- [x] 3.3 Add architecture and integration tests for broker-only DB ownership.
