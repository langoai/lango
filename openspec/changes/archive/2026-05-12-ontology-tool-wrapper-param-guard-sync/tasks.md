## 1. Wrapper Guard

- [x] 1.1 Tighten ontology governance wrappers to enforce declared required params.
- [x] 1.2 Tighten dynamic ontology action wrappers to enforce declared required params.
- [x] 1.3 Add regression coverage for the missing-parameter wrapper paths.

## 2. Verification

- [x] 2.1 Run `go test ./internal/ontology -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `ontology-tools` and `production-readiness` coverage for the wrapper guard contract.
- [ ] 3.2 Validate and archive the OpenSpec change.
