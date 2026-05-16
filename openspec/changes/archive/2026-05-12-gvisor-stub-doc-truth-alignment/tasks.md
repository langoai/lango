## 1. Runtime Truth Alignment

- [x] 1.1 Add a regression that checks explicit `gvisor` runtime requests return actionable unavailable wording.
- [x] 1.2 Update public configuration docs to describe `gvisor` as the current stub implementation.

## 2. Verification

- [x] 2.1 Run `go test ./internal/sandbox -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update production-readiness and downstream-docs-sync specs for the gVisor stub contract.
- [ ] 3.2 Validate and archive the OpenSpec change.
