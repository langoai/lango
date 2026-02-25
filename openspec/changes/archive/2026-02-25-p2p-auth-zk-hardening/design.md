# Design: P2P Auth & ZK Hardening

## Architecture Decisions

### 1. Dual Protocol Versioning
**Decision**: Register both `/lango/handshake/1.0.0` and `/lango/handshake/1.1.0` handlers.
**Rationale**: Zero-downtime migration. Old peers continue working, new peers get signed challenges.
**Alternative rejected**: Breaking protocol change with forced upgrade.

### 2. NonceCache as Struct (not interface)
**Decision**: Concrete `NonceCache` struct with Start/Stop lifecycle.
**Rationale**: Simple, single implementation needed. Goroutine cleanup matches existing buffer patterns (EmbeddingBuffer, GraphBuffer).

### 3. AttestationResult in Firewall Package
**Decision**: Define `AttestationResult` struct in `firewall` package instead of `protocol`.
**Rationale**: Avoids circular imports (`firewall` → `protocol` → `firewall`). Firewall is the producer, protocol is the consumer.

### 4. ZKAttestVerifyFunc as Callback
**Decision**: Use callback pattern for attestation verification on remote agent.
**Rationale**: Consistent with existing codebase patterns (ApprovalFunc, ZKProverFunc, ZKVerifierFunc). Avoids import cycles.

### 5. SRS Mode as Config (not Build Tag)
**Decision**: Runtime config `srsMode: "unsafe"|"file"` instead of build tags.
**Rationale**: Build tags are for dependency isolation (KMS providers). SRS is a runtime choice, not a dependency.

## Dependency Flow

```
config/types.go (RequireSignedChallenge, SRSMode, MaxCredentialAge)
    ↓
config/loader.go (defaults)
    ↓
app/wiring.go (creates NonceCache, wires to Handshaker, registers dual protocols,
               updates attestation callback, passes config to gossip)
    ↓
p2p/handshake/ (NonceCache + signed challenge + timestamp validation)
p2p/firewall/ (AttestationResult)
p2p/protocol/ (AttestationData + verification callback)
p2p/discovery/ (credential revocation)
zkp/circuits/ (attestation freshness + capability binding)
zkp/ (SRS file support)
```

## Files Modified

| File | Layer | Changes |
|------|-------|---------|
| handshake/nonce_cache.go | Core | NEW — TTL nonce cache |
| handshake/nonce_cache_test.go | Test | NEW — 7 test cases |
| handshake/handshake.go | Core | Signed challenge, timestamp validation, nonce cache |
| circuits/attestation.go | Core | MinTimestamp/MaxTimestamp |
| circuits/capability.go | Core | AgentTestBinding fix |
| circuits/circuits_test.go | Test | NEW — 15 circuit tests |
| zkp/zkp.go | Core | SRS file support |
| zkp/zkp_test.go | Test | NEW — 6 prover tests |
| protocol/messages.go | Core | AttestationData struct |
| protocol/handler.go | Application | Structured attestation construction |
| protocol/remote_agent.go | Application | Attestation verification |
| firewall/firewall.go | Core | AttestationResult, typed ZKAttestFunc |
| discovery/gossip.go | Application | Credential revocation |
| config/types.go | Core | New config fields |
| config/loader.go | Core | Default values |
| app/wiring.go | Application | NonceCache, dual protocol, attestation wiring |
