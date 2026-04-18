## ADDED Requirements

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
