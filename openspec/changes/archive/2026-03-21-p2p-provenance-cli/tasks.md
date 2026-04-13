## 1. OpenSpec

- [x] 1.1 Add delta spec updates for remote provenance CLI and protocol fetch behavior
- [x] 1.2 Capture server-backed execution and active-session requirements in proposal/design

## 2. Protocol + Runtime

- [x] 2.1 Extend `internal/p2p/provenanceproto` with `fetch_bundle` request/response helpers
- [x] 2.2 Add gateway routes for provenance push/fetch using running app services
- [x] 2.3 Reuse active P2P session tokens and DID-derived peer resolution in the route handlers

## 3. CLI + Docs

- [x] 3.1 Add `lango p2p provenance push` and `lango p2p provenance fetch`
- [x] 3.2 Update README and P2P CLI docs with the new commands and server requirement

## 4. Verification

- [x] 4.1 Add or update tests for no-server, no-session, push, fetch, and tampered bundle paths
- [x] 4.2 Run `go build ./...`
- [x] 4.3 Run `go test ./...`
- [x] 4.4 Validate and archive the change
