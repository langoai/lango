## 1. OpenSpec + Contracts

- [x] 1.1 Add delta spec updates for `session-provenance` covering persistent session tree, attribution, bundle export/import, and P2P-verifiable transport
- [x] 1.2 Update proposal/design artifacts to describe wallet/DID-based signing and verify-and-store import semantics

## 2. Core Provenance Storage

- [x] 2.1 Implement Ent-backed `SessionTreeStore`
- [x] 2.2 Add `provenance_attribution` Ent schema and generate Ent code
- [x] 2.3 Add `AttributionStore` interface with memory/Ent implementations as needed for tests and runtime
- [x] 2.4 Add attribution service/report logic joining token usage and checkpoints

## 3. Bundle + Verification

- [x] 3.1 Extend `ProvenanceBundle` with signer and signature metadata
- [x] 3.2 Implement canonical bundle serialization, redaction, export, import, and verify-and-store behavior
- [x] 3.3 Add DID/public-key signature verification helpers using wallet-compatible secp256k1 signatures

## 4. Runtime Wiring

- [x] 4.1 Expand provenance module values to expose session tree, attribution, and bundle services
- [x] 4.2 Wire child-session lifecycle into runtime multi-agent path and persist fork/merge/discard into session provenance
- [x] 4.3 Hook workspace git operations to emit git-aware attribution records

## 5. CLI + P2P

- [x] 5.1 Replace session tree/list CLI placeholders with working commands
- [x] 5.2 Replace attribution show/report CLI placeholders with working commands
- [x] 5.3 Add provenance bundle export/import CLI commands
- [x] 5.4 Add dedicated provenance P2P transport with DID/signature verification

## 6. Downstream + Verification

- [x] 6.1 Update README/docs/spec text to remove placeholder provenance language and document the new commands
- [x] 6.2 Add or update tests for session tree persistence, attribution capture/reporting, bundle round-trip/redaction/signature verification, and P2P transport
- [x] 6.3 Run `go build ./...`
- [x] 6.4 Run `go test ./...`
- [x] 6.5 Complete the OpenSpec follow-through: verify, sync specs, and archive the change
