## Why

Provenance currently ships only the checkpoint slice of the intended capability. Session tree persistence is still memory-only, attribution is only a type sketch with CLI stubs, and provenance bundles are not implemented. That leaves the system unable to provide durable multi-agent lineage, auditable workspace contributions, or portable provenance exchange over P2P.

## What Changes

- Complete session provenance end-to-end: persistent session tree, git-aware attribution, portable bundles, and working CLI.
- Replace session and attribution placeholders with implemented storage, services, CLI commands, and P2P transport.
- Reuse existing RunLedger, session, workspace gitbundle, token usage, wallet, and DID identity layers rather than inventing a parallel provenance stack.

## Capabilities

### Modified Capabilities

- `session-provenance`: Promote the feature from checkpoint-first scaffolding to a shipped provenance system covering session tree, attribution, and bundle transport.

## Impact

- `internal/provenance/`: add session tree Ent store, attribution store/service, bundle service, import/export, and verification helpers
- `internal/ent/schema/`: add `provenance_attribution` and wire existing `session_provenance` for runtime use
- `internal/app/`: wire provenance runtime capture from workspace git operations and child-session lifecycle
- `internal/cli/provenance/`: replace stubs with working session, attribution, and bundle commands
- `internal/p2p/`: add dedicated provenance bundle transport with DID-verifiable signatures
- Docs/specs/README updated to remove placeholder language and document the real feature surface
