## 1. Wrapper Guard

- [x] 1.1 Add a shared required-integer extraction helper.
- [x] 1.2 Tighten team workflow tool wrappers to enforce declared required params.
- [x] 1.3 Add regression coverage for team wrapper missing-parameter paths.

## 2. Verification

- [x] 2.1 Run `go test ./internal/toolparam ./internal/p2p/team -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `team-connectivity`, `production-readiness`, and `toolparam-extraction` coverage for the wrapper guard contract.
- [ ] 3.2 Validate and archive the OpenSpec change.
