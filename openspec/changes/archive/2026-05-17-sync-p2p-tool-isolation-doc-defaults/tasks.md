# Tasks

## 1. Planning

- [x] 1.1 Audit P2P tool isolation defaults in code, README, docs, and specs.
- [x] 1.2 Add focused OpenSpec artifacts for public config docs default parity.
- [x] 1.3 Validate the OpenSpec change before implementation.
- [x] 1.4 Commit the initial OpenSpec planning artifacts separately before implementation.

## 2. Regression Test

- [x] 2.1 Add a failing documentation guard that compares public P2P default rows against `config.DefaultConfig()`.
- [x] 2.2 Confirm the guard fails against stale public P2P defaults.

## 3. Implementation

- [x] 3.1 Correct stale public documentation values found by the guard.
- [x] 3.2 Keep runtime configuration defaults unchanged.

## 4. Review

- [x] 4.1 Request teammate review for spec compliance and code quality.
- [x] 4.2 Address any actionable findings before archiving.

## 5. Verification

- [x] 5.1 Run focused `internal/testutil` tests.
- [x] 5.2 Run `go build ./...`.
- [x] 5.3 Run `go test ./...`.
- [x] 5.4 Run `openspec validate --all --strict`.
- [x] 5.5 Run `git diff --check`.
- [x] 5.6 Archive the change after verification.
