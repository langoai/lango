## ADDED Requirements

### Requirement: archtest enforces economy/p2p boundary
The `internal/archtest` package SHALL contain a test that fails when any `internal/economy/**` package imports any `internal/p2p/**` package.

#### Scenario: Clean boundary passes
- **WHEN** `go test ./internal/archtest/...` is run and no economy→p2p imports exist
- **THEN** the test passes with 0 violations

#### Scenario: Violation fails the test
- **WHEN** an `internal/economy/` package adds an import of `internal/p2p/`
- **THEN** `go test ./internal/archtest/...` fails with an error identifying the violation

### Requirement: archtest enforces p2p-infra boundary
The `internal/archtest` package SHALL enforce that P2P infrastructure packages (`discovery`, `handshake`, `firewall`, `protocol`, `agentpool`) do NOT import `internal/economy/**`, `internal/payment/**`, or `internal/wallet/**`.

#### Scenario: p2p infrastructure packages have no commerce imports
- **WHEN** the import graph is scanned
- **THEN** no p2p infrastructure package imports economy, payment, or wallet packages

### Requirement: archtest uses go list without external dependencies
The archtest implementation SHALL use `go list -json` via `os/exec` to parse the import graph. It MUST NOT add `golang.org/x/tools/go/packages` or any other external dependency to `go.mod`.

#### Scenario: No new go.mod dependencies from archtest
- **WHEN** `internal/archtest/boundary_test.go` is compiled
- **THEN** no new entries are added to `go.mod` beyond what already exists

### Requirement: archtest checks production imports only
The archtest SHALL check production code imports only. Test-only imports (`_test.go` files) are excluded from boundary enforcement.

#### Scenario: Test file cross-domain import is not flagged
- **WHEN** a test file in `internal/economy/` imports `internal/p2p/` for test fixtures
- **THEN** archtest does NOT report this as a violation

### Requirement: depguard rules in golangci-lint
The `.golangci.yml` SHALL include depguard rules that block `economy→p2p` and `p2p-infra→economy/wallet` imports at the linter level.

#### Scenario: Linter catches economy→p2p import
- **WHEN** `golangci-lint run` is executed on code where `internal/economy/` imports `internal/p2p/`
- **THEN** depguard reports a violation

### Requirement: p2p/handshake is exempt from depguard wallet restriction
The depguard configuration SHALL NOT include `p2p/handshake` in the restricted package list, because it legitimately uses `wallet.WalletProvider` for cryptographic signing.

#### Scenario: handshake wallet import does not trigger lint error
- **WHEN** `golangci-lint run ./internal/p2p/handshake/...` is executed
- **THEN** no depguard violations are reported for the wallet import
