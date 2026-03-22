## Why

Provenance bundles and the dedicated P2P provenance protocol already exist in runtime code, but users still have no direct way to exchange provenance with peers. The current `lango p2p` CLI exposes git/workspace flows while provenance exchange remains hidden behind internal services.

## What Changes

- Add a server-backed `lango p2p provenance` CLI surface for remote provenance exchange
- Extend the provenance P2P protocol with fetch support in addition to push/import
- Add gateway endpoints that bridge local CLI requests to the running app's P2P session, provenance bundle, and transport services
- Document the new commands and their requirement for an active P2P session plus a running server

## Capabilities

### Modified Capabilities

- `session-provenance`: add remote provenance bundle push/fetch via CLI + gateway
- `p2p-network`: expose provenance exchange over the existing authenticated peer session model

## Impact

- `internal/cli/p2p/`: new provenance command group and HTTP gateway client helpers
- `internal/app/p2p_routes.go`: new `/api/p2p/provenance/push` and `/api/p2p/provenance/fetch` routes
- `internal/p2p/provenanceproto/`: add `fetch_bundle` request/response
- README and P2P CLI docs updated with the new commands
