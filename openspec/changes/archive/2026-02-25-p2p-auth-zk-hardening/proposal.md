# P2P Security Hardening: Authentication (B- → A) & ZK Proofs (C → B+)

## Problem

The P2P security roadmap had all P0/P1/P2 items completed, but two critical areas remained at low grades:

1. **P2P Authentication (B-)**: P0-2 signature verification was complete, but Challenge messages were unsigned, allowing initiator identity spoofing. No nonce replay protection or timestamp validation existed.

2. **ZK Proofs (C)**: Four circuits were defined but had zero test coverage, ResponseAttestation had no timestamp freshness enforcement, AgentCapability circuit had a discarded binding (line 48: `_ = hAgent.Sum()`), and attestation data was opaque bytes with no structured verification.

## Solution

### Authentication Hardening (B- → A)
- Sign Challenge messages with ECDSA (nonce || bigEndian(timestamp) || senderDID → Keccak256 → secp256k1)
- Add NonceCache with TTL-based replay detection
- Validate challenge timestamps (5 min past + 30s future)
- Dual protocol versioning (v1.0 legacy + v1.1 signed)
- Configurable `requireSignedChallenge` for strict mode

### ZK Proof Hardening (C → B+)
- Complete test suite: 15 circuit tests + 6 ProverService tests
- Attestation timestamp freshness (MinTimestamp/MaxTimestamp range constraints)
- Capability binding fix (AgentTestBinding properly constrained)
- Structured AttestationData wire format (proof + publicInputs + circuitID + scheme)
- Attestation verification callback on remote agent
- SRS production file path support
- Credential revocation in gossip discovery

## Scope

- 16 files modified/created
- All changes backward compatible (default settings preserve existing behavior)
- No breaking protocol changes (dual version registration)
