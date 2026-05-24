## 1. Planning

- [x] 1.1 Identify the storage broker stderr routing gap.
- [x] 1.2 Document the scoped OpenSpec change.

## 2. TDD

- [x] 2.1 Add a failing unit test for injected broker stderr routing.
- [x] 2.2 Implement the storage broker stderr seam and command helper.
- [x] 2.3 Run the focused storagebroker tests.

## 3. Verification

- [x] 3.1 Run `go build ./...`.
- [x] 3.2 Run `go test ./...`.
- [x] 3.3 Run `openspec validate --all --strict`.
- [x] 3.4 Archive the OpenSpec change and confirm no active change remains.
