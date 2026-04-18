## 1. Broker Config/Profile APIs

- [x] 1.1 Add broker protocol/server/client methods for config profile load/save/set-active/list/delete/exists.
- [x] 1.2 Add a broker-backed `storage.ConfigProfileStore` adapter and wire it into bootstrap profile loading.

## 2. Broker Session Bootstrap APIs

- [x] 2.1 Add broker protocol/server/client methods for session bootstrap/store operations needed by runtime wiring.
- [x] 2.2 Add a broker-backed `session.Store` adapter and connect it to `storage.Facade.OpenSessionStore`.

## 3. Verification

- [ ] 3.1 Add tests for broker-backed config profile and session adapter round-trips.
- [ ] 3.2 Run `go build ./...`, `go test ./...`, and `openspec validate --type change broker-ownership-bootstrap`.
