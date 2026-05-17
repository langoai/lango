# Tasks

## 1. Planning

- [x] 1.1 Audit P2P tool isolation defaults in code, README, docs, and specs.
- [x] 1.2 Add focused OpenSpec artifacts for public config docs default parity.
- [x] 1.3 Validate the OpenSpec change before implementation.
- [ ] 1.4 Commit the OpenSpec planning artifacts separately.

## 2. Regression Test

- [ ] 2.1 Add a failing documentation guard that compares public P2P tool isolation default rows against `config.DefaultConfig()`.
- [ ] 2.2 Confirm the guard fails against the stale README `p2p.toolIsolation.maxMemoryMB` default.

## 3. Implementation

- [ ] 3.1 Correct stale public documentation values found by the guard.
- [ ] 3.2 Keep runtime configuration defaults unchanged.

## 4. Review

- [ ] 4.1 Request teammate review for spec compliance and code quality.
- [ ] 4.2 Address any actionable findings before archiving.

## 5. Verification

- [ ] 5.1 Run focused `internal/testutil` tests.
- [ ] 5.2 Run `go build ./...`.
- [ ] 5.3 Run `go test ./...`.
- [ ] 5.4 Run `openspec validate --all --strict`.
- [ ] 5.5 Run `git diff --check`.
- [ ] 5.6 Archive the change after verification.
