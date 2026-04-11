## Tasks

### Wave 1 — KEM primitives

- [x] Promote `cloudflare/circl` from indirect to direct dependency
- [x] Create `internal/security/kem.go` — `AlgorithmX25519MLKEM768`, `KEMDecapsulator` type, `GenerateEphemeralKEM()`, `KEMEncapsulate()`
- [x] Create `internal/security/kem_test.go` — roundtrip, invalid pubkey rejection, invalid ciphertext rejection
- [x] Create `internal/security/derive_session.go` — `DeriveSessionKey()`
- [x] Create `internal/security/derive_session_test.go` — determinism, wrong size rejection, different DIDs = different keys

### Wave 2 — Handshake protocol v1.2

- [x] Add KEM fields to Challenge/ChallengeResponse structs (omitempty)
- [x] Add `ProtocolIDv12` constant and `PreferredProtocols()` helper
- [x] Add `EnablePQKEM` to Config, `kemEnabled` to Handshaker
- [x] Extend `challengeCanonicalPayload()` with KEM transcript binding
- [x] Add `responseCanonicalPayload()` for response transcript binding
- [x] Update `verifyChallengeSignature()` and `verifyResponse()` for v1.2 payloads
- [x] Implement KEM exchange in `Initiate()` — generate keypair, decapsulate, derive session key
- [x] Implement KEM exchange in `HandleIncoming()` — encapsulate, derive session key, sign transcript-bound payload
- [x] Add `EncryptionKey`/`KEMUsed` to Session struct
- [x] Add EncryptionKey zeroing to `Create()` overwrite, `Remove()`, `Invalidate()`, `InvalidateAll()`, `InvalidateByCondition()`, `Cleanup()`
- [x] Add `StreamHandlerV12()` method
- [x] Add handshake tests — KEM roundtrip, graceful degradation, key zeroing, transcript binding

### Wave 3 — Wiring + config + CLI

- [x] Add `EnablePQHandshake` to `P2PConfig`
- [x] Wire `EnablePQKEM` in `wiring_p2p.go`, register v1.2 handler
- [x] Update `tools_p2p.go` initiator to use `PreferredProtocols()`
- [x] Add PQ KEM status to `cli/security/status.go`

### Wave 4 — OpenSpec finalize

- [x] Verify change
- [x] Sync delta specs
- [x] Archive change
