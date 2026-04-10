## MODIFIED Requirements

### Requirement: archtest enforces p2p-infra boundary

The `p2pNetworkingPrefixes` list SHALL include `internal/p2p/identity` in addition to discovery, handshake, firewall, protocol, and agentpool. The `forbiddenForP2P` matching logic SHALL correctly match both exact package paths and sub-packages (not rely on trailing slashes).

#### Scenario: p2p/identity included in networking boundary
- **WHEN** the import graph is scanned
- **THEN** `internal/p2p/identity` SHALL be subject to the same economy/payment/wallet import restrictions as other p2p infrastructure packages

#### Scenario: Exact package path matching (no trailing slash)
- **WHEN** `internal/p2p/handshake` imports `internal/wallet` (no sub-package)
- **THEN** archtest SHALL detect and report the violation

### Requirement: archtest enforces provenance boundary

The archtest SHALL enforce that `internal/provenance` packages do NOT import `internal/p2p/identity` packages. Verification implementations are injected at the app/cli wiring layer.

#### Scenario: provenance importing p2p/identity is a violation
- **WHEN** `internal/provenance` imports `internal/p2p/identity`
- **THEN** archtest SHALL report a boundary violation

### Requirement: p2p/handshake is no longer exempt from wallet restriction

Since `p2p/handshake` now uses a consumer-local `Signer` interface instead of `wallet.WalletProvider`, the depguard exemption for `p2p/handshake` wallet import is no longer needed.

#### Scenario: handshake wallet import triggers lint error
- **WHEN** `p2p/handshake` imports `internal/wallet`
- **THEN** depguard SHALL report a violation

## REMOVED Requirements

### Requirement: p2p/handshake is exempt from depguard wallet restriction
