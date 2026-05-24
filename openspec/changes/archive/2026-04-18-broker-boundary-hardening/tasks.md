## 1. Facade Hardening

- [x] 1.1 Add capability-specific facade methods for production app/CLI paths that currently depend on generic Ent/SQL handles.
- [x] 1.2 Remove production use of generic `EntClient()` / `RawDB()` accessors from app and CLI code.

## 2. App Wiring

- [x] 2.1 Move ontology, observability alerts, workflow state, and P2P reputation/settlement wiring onto facade capabilities.
- [x] 2.2 Keep bootstrap/test-only wiring functional without reintroducing production raw-handle consumers.

## 3. Verification

- [x] 3.1 Add or update tests for the new storage capability paths where needed.
- [x] 3.2 Update affected docs/specs and run `go build ./...`, `go test ./...`, and `openspec validate --type change broker-boundary-hardening`.
