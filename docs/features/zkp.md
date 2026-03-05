---
title: Zero-Knowledge Proofs
---

# Zero-Knowledge Proofs

Lango uses zero-knowledge proofs (ZKPs) for privacy-preserving identity verification, capability attestation, and response authenticity in the P2P network. The ZKP system is built on the [gnark](https://github.com/Consensys/gnark) library using the BN254 elliptic curve.

## Proving Schemes

Two proving schemes are supported:

| Scheme | Use Case | Trade-offs |
|--------|----------|------------|
| **PlonK** | Default, general purpose | Universal setup (one SRS for all circuits), slightly larger proofs |
| **Groth16** | Performance-critical | Per-circuit trusted setup, smallest proofs, fastest verification |

Configure via `p2p.zkp.provingScheme`:

```json
{
  "p2p": {
    "zkp": {
      "provingScheme": "plonk"
    }
  }
}
```

## Circuits

Lango defines four ZKP circuits, each proving a specific statement without revealing private data.

### 1. Wallet Ownership (`WalletOwnershipCircuit`)

Proves knowledge of a secret response that produces the expected public key hash when combined with a challenge.

| Input | Visibility | Description |
|-------|-----------|-------------|
| `PublicKeyHash` | Public | Expected hash of the agent's public key |
| `Challenge` | Public | Random challenge value |
| `Response` | Private | Secret response (witness) |

**Constraint**: `MiMC(Response, Challenge) == PublicKeyHash`

### 2. Agent Capability (`AgentCapabilityCircuit`)

Proves that an agent possesses a specific capability with a score meeting a minimum threshold, without revealing the actual score or test details.

| Input | Visibility | Description |
|-------|-----------|-------------|
| `CapabilityHash` | Public | Hash of the capability proof |
| `AgentDIDHash` | Public | Hash of the agent's DID |
| `MinScore` | Public | Minimum required score |
| `AgentTestBinding` | Public | Binding between agent and capability test |
| `ActualScore` | Private | Agent's actual capability score |
| `TestHash` | Private | Hash of the capability test |

**Constraints**:
- `ActualScore >= MinScore`
- `MiMC(TestHash, ActualScore) == CapabilityHash`
- `MiMC(TestHash, AgentDIDHash) == AgentTestBinding`

### 3. Balance Range (`BalanceRangeCircuit`)

Proves that a private balance meets a minimum threshold without revealing the actual amount.

| Input | Visibility | Description |
|-------|-----------|-------------|
| `Threshold` | Public | Minimum required balance |
| `Balance` | Private | Actual balance value |

**Constraint**: `Balance >= Threshold`

### 4. Response Attestation (`ResponseAttestationCircuit`)

Proves that an agent produced a response derived from specific source data, with timestamp freshness guarantees.

| Input | Visibility | Description |
|-------|-----------|-------------|
| `ResponseHash` | Public | Hash of the response |
| `AgentDIDHash` | Public | Hash of the agent's DID |
| `Timestamp` | Public | Response timestamp |
| `MinTimestamp` | Public | Minimum valid timestamp |
| `MaxTimestamp` | Public | Maximum valid timestamp |
| `SourceDataHash` | Private | Hash of the source data |
| `AgentKeyProof` | Private | Agent's private key proof |

**Constraints**:
- `MiMC(AgentKeyProof) == AgentDIDHash`
- `MiMC(SourceDataHash, AgentKeyProof, Timestamp) == ResponseHash`
- `MinTimestamp <= Timestamp <= MaxTimestamp`

## Structured Reference String (SRS)

PlonK requires a Structured Reference String (SRS) for the trusted setup. Two modes are supported:

| Mode | Description | Use Case |
|------|-------------|----------|
| `unsafe` | Deterministic SRS generated at runtime | Development and testing |
| `file` | SRS loaded from a pre-generated file | Production deployments |

When `file` mode is configured but the SRS file is missing, the system falls back to `unsafe` mode with a warning.

```json
{
  "p2p": {
    "zkp": {
      "srsMode": "file",
      "srsPath": "/path/to/ceremony-srs.bin"
    }
  }
}
```

The SRS file contains two KZG commitments (canonical and Lagrange) written sequentially in binary format.

## Prover Service

The `ProverService` manages the full ZKP lifecycle:

1. **Compile** — Compile a circuit and generate proving/verifying keys
2. **Prove** — Generate a proof from a circuit assignment (witness)
3. **Verify** — Check whether a proof is valid

Compiled circuits are cached in memory by circuit ID. The cache directory defaults to `~/.lango/zkp/cache/`.

### Proof Structure

```json
{
  "data": "<base64-encoded-proof>",
  "publicInputs": "<base64-encoded-public-witness>",
  "circuitId": "attestation",
  "scheme": "plonk"
}
```

## P2P Integration

### ZK Handshake

When `p2p.zkHandshake` is enabled, peer authentication includes a zero-knowledge proof of DID ownership using the `WalletOwnershipCircuit`.

### ZK Attestation

When `p2p.zkAttestation` is enabled, P2P responses include a `ResponseAttestationCircuit` proof with timestamp freshness bounds. The attestation data is structured as:

```json
{
  "proof": "<base64>",
  "publicInputs": ["<agent-id-hash>", "<min-ts>", "<max-ts>"],
  "circuitId": "attestation",
  "scheme": "plonk"
}
```

### Credential Revocation

ZK credentials have a configurable maximum age (`p2p.zkp.maxCredentialAge`). Credentials older than this duration are rejected during agent card validation, even if not explicitly revoked.

## Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| `p2p.zkHandshake` | `false` | Enable ZK proof during peer handshake |
| `p2p.zkAttestation` | `false` | Enable ZK attestation on P2P responses |
| `p2p.requireSignedChallenge` | `false` | Reject unsigned (v1.0) handshake challenges |
| `p2p.zkp.provingScheme` | `"plonk"` | Proving scheme: `plonk` or `groth16` |
| `p2p.zkp.srsMode` | `"unsafe"` | SRS source: `unsafe` or `file` |
| `p2p.zkp.srsPath` | `""` | Path to SRS file (when srsMode is `file`) |
| `p2p.zkp.maxCredentialAge` | `"24h"` | Maximum age for ZK credentials |
| `p2p.zkp.proofCacheDir` | `"~/.lango/zkp"` | Directory for ZKP cache files |

## CLI Commands

```bash
lango p2p zkp status         # Show ZKP configuration and compiled circuits
lango p2p zkp circuits       # List available circuits with constraint counts
```
