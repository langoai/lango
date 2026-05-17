# Tasks

## 1. Planning

- [x] 1.1 Confirm existing sandbox worker protocol and test-coverage specs.
- [x] 1.2 Add focused OpenSpec artifacts for the stdio seam change.
- [x] 1.3 Validate the OpenSpec change before implementation.

## 2. Regression Test

- [x] 2.1 Add a failing test proving public `RunWorker` reads from injected stdin and writes to injected stdout.
- [x] 2.2 Confirm the new test fails before production code changes.

## 3. Implementation

- [x] 3.1 Add unexported worker stdio seams with production defaults.
- [x] 3.2 Route `RunWorker` through the seams while leaving `RunWorkerWithIO` unchanged.

## 4. Verification

- [x] 4.1 Run focused sandbox worker tests.
- [x] 4.2 Run the full sandbox package tests.
- [x] 4.3 Run `go build ./...`.
- [x] 4.4 Run `go test ./...`.
- [x] 4.5 Run `openspec validate --all --strict`.
- [x] 4.6 Archive the change after verification.
