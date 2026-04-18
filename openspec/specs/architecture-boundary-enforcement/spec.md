## Purpose

Capability spec for architecture-boundary-enforcement. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: archtest enforces economy/p2p boundary
The `internal/archtest` package SHALL contain a test that fails when any `internal/economy/**` package imports any `internal/p2p/**` package.

#### Scenario: Clean boundary passes
- **WHEN** `go test ./internal/archtest/...` is run and no economy→p2p imports exist
- **THEN** the test passes with 0 violations

#### Scenario: Violation fails the test
- **WHEN** an `internal/economy/` package adds an import of `internal/p2p/`
- **THEN** `go test ./internal/archtest/...` fails with an error identifying the violation

### Requirement: archtest enforces p2p-infra boundary
The `internal/archtest` package SHALL enforce that P2P infrastructure packages (`discovery`, `handshake`, `identity`, `firewall`, `protocol`, `agentpool`) do NOT import `internal/economy/**`, `internal/payment/**`, or `internal/wallet/**`. The matching logic SHALL correctly match both exact package paths and sub-packages.

#### Scenario: p2p infrastructure packages have no commerce imports
- **WHEN** the import graph is scanned
- **THEN** no p2p infrastructure package imports economy, payment, or wallet packages

#### Scenario: p2p/identity included in networking boundary
- **WHEN** the import graph is scanned
- **THEN** `internal/p2p/identity` SHALL be subject to the same economy/payment/wallet import restrictions

#### Scenario: Exact package path matching
- **WHEN** `internal/p2p/handshake` imports `internal/wallet` (no sub-package)
- **THEN** archtest SHALL detect and report the violation

### Requirement: archtest enforces provenance boundary

The archtest SHALL enforce that `internal/provenance` packages do NOT import `internal/p2p/identity` packages. Verification implementations are injected at the app/cli wiring layer.

#### Scenario: provenance importing p2p/identity is a violation
- **WHEN** `internal/provenance` imports `internal/p2p/identity`
- **THEN** archtest SHALL report a boundary violation

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

### Requirement: p2p/handshake and p2p/identity are included in depguard wallet restriction

Since `p2p/handshake` uses a consumer-local `Signer` interface and `p2p/identity` uses a consumer-local `KeyProvider` interface (neither imports `internal/wallet`), both SHALL be included in the `p2p-infra-no-economy` depguard rule file list.

#### Scenario: handshake wallet import triggers depguard violation
- **WHEN** `golangci-lint run ./internal/p2p/handshake/...` is executed and handshake imports `internal/wallet`
- **THEN** depguard SHALL report a violation

#### Scenario: identity wallet import triggers depguard violation
- **WHEN** `golangci-lint run ./internal/p2p/identity/...` is executed and identity imports `internal/wallet`
- **THEN** depguard SHALL report a violation

### Requirement: archtest forbids removed storage facade raw accessors in production packages
The `internal/archtest` package MUST fail when production packages reintroduce removed raw storage facade accessors.

#### Scenario: Production package uses removed raw accessor
- **WHEN** any production package under `internal/app`, `internal/cli`, or selected domain packages references `EntClient()`, `RawDB()`, `FTSDB()`, or `PaymentClient()`
- **THEN** `go test ./internal/archtest/...` fails and reports the offending line

### Requirement: archtest forbids payment-side Ent extraction regressions
The `internal/archtest` package MUST fail when production payment wiring reintroduces direct `session.EntStore.Client()` extraction.

#### Scenario: Payment wiring extracts Ent client directly
- **WHEN** production payment wiring or CLI payment initialization calls `entStore.Client()`
- **THEN** `go test ./internal/archtest/...` fails and reports the offending line

### Requirement: archtest forbids test-only storage wiring helpers in production packages
The `internal/archtest` package MUST fail when production packages use test-only facade wiring helpers such as `storage.WithEntClient` or `storage.WithRawDB`.

#### Scenario: Production package uses test-only facade helper
- **WHEN** a production package references `storage.WithEntClient(...)` or `storage.WithRawDB(...)`
- **THEN** `go test ./internal/archtest/...` fails and reports the offending line

