# Tasks

## 1. Planning

- [x] 1.1 Confirm existing sandbox worker protocol and test-coverage specs.
- [x] 1.2 Add focused OpenSpec artifacts for the stdio seam change.
- [x] 1.3 Validate the OpenSpec change before implementation.

## 2. Regression Test

- [ ] 2.1 Add a failing test proving public `RunWorker` reads from injected stdin and writes to injected stdout.
- [ ] 2.2 Confirm the new test fails before production code changes.

## 3. Implementation

- [ ] 3.1 Add unexported worker stdio seams with production defaults.
- [ ] 3.2 Route `RunWorker` through the seams while leaving `RunWorkerWithIO` unchanged.

## 4. Verification

- [ ] 4.1 Run focused sandbox worker tests.
- [ ] 4.2 Run the full sandbox package tests.
- [ ] 4.3 Run `go build ./...`.
- [ ] 4.4 Run `go test ./...`.
- [ ] 4.5 Run `openspec validate --all --strict`.
- [ ] 4.6 Archive the change after verification.
