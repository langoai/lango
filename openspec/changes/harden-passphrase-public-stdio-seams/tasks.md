# Tasks

## 1. Planning

- [x] 1.1 Define the passphrase stdio seam scope.
- [x] 1.2 Limit specs to passphrase acquisition and executable coverage.

## 2. Tests

- [ ] 2.1 Add failing wrapper-level tests for `Acquire` stdin and stderr seams.
- [ ] 2.2 Add failing wrapper-level test for `AcquireNonInteractive` stderr seam.

## 3. Implementation

- [ ] 3.1 Route public passphrase wrappers through package-level stdio seams.
- [ ] 3.2 Preserve existing helper APIs and runtime behavior.

## 4. Verification

- [ ] 4.1 Run focused passphrase tests.
- [ ] 4.2 Run `go build ./...`.
- [ ] 4.3 Run `go test ./...`.
- [ ] 4.4 Run `openspec validate --all --strict`.
- [ ] 4.5 Archive the OpenSpec change after implementation is verified.
