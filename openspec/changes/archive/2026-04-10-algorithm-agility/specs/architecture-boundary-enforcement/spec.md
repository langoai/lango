## MODIFIED Requirements

### Requirement: archtest enforces p2p-infra boundary

The depguard `p2p-infra-no-economy` rule SHALL include `p2p/handshake` and `p2p/identity` in the protected file list, since both packages no longer import `internal/wallet` (resolved in Phase 0).

#### Scenario: handshake wallet import triggers depguard violation
- **WHEN** `golangci-lint run` is executed and `p2p/handshake` imports `internal/wallet`
- **THEN** depguard SHALL report a violation

#### Scenario: identity wallet import triggers depguard violation
- **WHEN** `golangci-lint run` is executed and `p2p/identity` imports `internal/wallet`
- **THEN** depguard SHALL report a violation

## REMOVED Requirements

### Requirement: p2p/handshake is exempt from depguard wallet restriction
