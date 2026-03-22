## Context

The repo now has:

- signed provenance bundle export/import in `internal/provenance`
- dedicated provenance transport in `internal/p2p/provenanceproto`
- active P2P session tokens in `handshake.SessionStore`
- a gateway router that already exposes authenticated `/api/p2p/*` endpoints

What is missing is a supported user path that reuses those runtime services instead of creating an extra ephemeral node/session model in the CLI.

## Goals / Non-Goals

**Goals**
- Provide `lango p2p provenance push` and `lango p2p provenance fetch`
- Keep the CLI server-backed, matching the existing runtime-only P2P workspace/git pattern
- Reuse active peer session tokens from the running app
- Allow fetch to request a remote bundle by `session-key` and `redaction`

**Non-Goals**
- Standalone direct CLI P2P node management for provenance exchange
- Implicit session creation or handshake from provenance commands
- Changing provenance import semantics away from verify-and-store

## Decisions

### D1. CLI is a thin HTTP client

The CLI will POST to new gateway endpoints instead of constructing its own P2P node or session store. This keeps session ownership inside the running app, where the active peer session token already exists.

### D2. Peer resolution uses DID parsing, not discovery lookup

`did:lango:<pubkey>` already deterministically embeds a public key and peer identity. The CLI routes resolve the remote peer from the DID directly, without requiring gossip/discovery state to be present.

### D3. Fetch is protocol-level, not two-step remote tool invocation

The provenance transport gets a `fetch_bundle` request. A remote peer exports and signs the bundle on demand, returns it over the provenance protocol, and the local side immediately verify-and-store imports it.

### D4. Active session token is mandatory

Push/fetch both fail fast if the running app has no active authenticated session for the target peer DID. Handshake remains the responsibility of the existing `lango p2p` flow.

## Risks / Trade-offs

- **Server dependency**: commands require `lango serve` to be running. This is intentional and matches existing runtime-bound workspace/git commands.
- **Wallet dependency on fetch**: the remote peer must be able to sign the exported bundle. If wallet identity is unavailable, fetch fails explicitly instead of falling back to unsigned transport.
