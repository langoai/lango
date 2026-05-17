# Tasks

## 1. Planning

- [x] 1.1 Audit bootstrap source comments, tests, and specs for phase inventory drift.
- [x] 1.2 Add focused OpenSpec artifacts for bootstrap phase copy sync.
- [x] 1.3 Validate the OpenSpec change before implementation.

## 2. Regression Test

- [ ] 2.1 Add a failing bootstrap source-comment guard for phase count and phase names.
- [ ] 2.2 Confirm the guard fails against the stale comments before editing source comments.

## 3. Implementation

- [ ] 3.1 Update `DefaultPhases` source comment to say 12 phases.
- [ ] 3.2 Update `Run` source comment to list the current 12-phase inventory.
- [ ] 3.3 Keep runtime phase order unchanged.

## 4. Verification

- [ ] 4.1 Run focused bootstrap phase inventory tests.
- [ ] 4.2 Run the full bootstrap package tests.
- [ ] 4.3 Run `go build ./...`.
- [ ] 4.4 Run `go test ./...`.
- [ ] 4.5 Run `openspec validate --all --strict`.
- [ ] 4.6 Run `git diff --check`.
- [ ] 4.7 Archive the change after verification.
