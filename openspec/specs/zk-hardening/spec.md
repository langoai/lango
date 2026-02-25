# ZK Proof Hardening Spec

## Overview

Hardens all four ZK circuits with proper testing, timestamp freshness, capability binding, structured attestation data, and production SRS support.

## Circuit Changes

### ResponseAttestationCircuit
- **Added public inputs**: `MinTimestamp`, `MaxTimestamp`
- **New constraints**: `MinTimestamp <= Timestamp <= MaxTimestamp`
- Ensures attestation proofs cannot be replayed outside the freshness window

### AgentCapabilityCircuit
- **Added public input**: `AgentTestBinding` (MiMC(TestHash, AgentDIDHash))
- **Fixed constraint**: `api.AssertIsEqual(hAgent.Sum(), c.AgentTestBinding)` (was `_ = hAgent.Sum()`)
- Makes the agent-test binding verifiable externally

### WalletOwnershipCircuit & BalanceRangeCircuit
- No structural changes, test coverage added

## Test Coverage

### Circuit Tests (circuits_test.go)
- 15 test cases across 4 circuits
- Framework: gnark `test.NewAssert(t)` with `test.WithCurves(ecc.BN254)`
- Both plonk and groth16 proving systems tested automatically
- MiMC hash computation via native `bn254/fr/mimc` package

### ProverService Tests (zkp_test.go)
- 6 integration tests: compile, prove, verify (valid/invalid), idempotent compile, uncompiled error
- Both plonk and groth16 schemes tested

## AttestationData Wire Format

```go
type AttestationData struct {
    Proof        []byte `json:"proof"`
    PublicInputs []byte `json:"publicInputs"`
    CircuitID    string `json:"circuitId"`
    Scheme       string `json:"scheme"`
}
```

### Firewall Integration
- `AttestationResult` struct in firewall package (avoids circular imports)
- `ZKAttestFunc` returns `*AttestationResult` instead of `[]byte`
- `AttestResponse()` returns structured data

### Remote Agent Verification
- `ZKAttestVerifyFunc` callback type for attestation verification
- `P2PRemoteAgent.SetAttestVerifier()` setter
- Verification logged in `InvokeTool()` response handling

### Backward Compatibility
- `Response.AttestationProof []byte` field retained (deprecated)
- New `Response.Attestation *AttestationData` field added
- Handler sets both fields for backward compat

## SRS Production Path

| Config Key | Type | Default | Description |
|------------|------|---------|-------------|
| `p2p.zkp.srsMode` | string | "unsafe" | SRS generation: "unsafe" or "file" |
| `p2p.zkp.srsPath` | string | "" | Path to SRS file |
| `p2p.zkp.maxCredentialAge` | string | "24h" | Max credential age |

## Credential Revocation

- `GossipService.revokedDIDs map[string]time.Time`
- `RevokeDID(did)` / `IsRevoked(did) bool`
- `SetMaxCredentialAge(d time.Duration)`
- Credential rejection: expired (ExpiresAt), stale (IssuedAt + maxCredentialAge), revoked (IsRevoked)
