# Tasks

## 1. Planning

- [x] 1.1 Audit current direct standard-stream references under `cmd/` and `internal/`.
- [x] 1.2 Add focused OpenSpec artifacts for a core stdio stream guard.
- [x] 1.3 Validate the OpenSpec change before implementation.

## 2. Regression Test

- [ ] 2.1 Add a failing fixture test proving the new core stream scanner rejects direct `os.Stdout` usage.
- [ ] 2.2 Confirm the new fixture test fails before implementing the scanner.

## 3. Implementation

- [ ] 3.1 Add the core production stream scanner and repository guard.
- [ ] 3.2 Add an explicit allowlist for existing intentional non-CLI core stdio seams.
- [ ] 3.3 Keep `cmd/`, `internal/cli/`, generated ent code, and test guard fixtures under their existing guard ownership.

## 4. Verification

- [ ] 4.1 Run focused `internal/testutil` guard tests.
- [ ] 4.2 Run `go build ./...`.
- [ ] 4.3 Run `go test ./...`.
- [ ] 4.4 Run `openspec validate --all --strict`.
- [ ] 4.5 Run `git diff --check`.
- [ ] 4.6 Archive the change after verification.
