## 1. Tests First

- [x] 1.1 Add a failing `internal/ontology` package test that rejects production `panic(` calls in schema exchange code.
- [x] 1.2 Run the focused test and confirm it fails on the existing digest marshal panic.

## 2. Implementation

- [x] 2.1 Replace panic-based digest marshaling with a checked helper returning `(string, error)`.
- [x] 2.2 Preserve the existing `ComputeDigest(types, predicates) string` API and successful digest output.
- [x] 2.3 Propagate checked digest errors through `ExportSchema`.

## 3. Verification

- [x] 3.1 Run focused ontology tests.
- [x] 3.2 Run `openspec validate ontology-digest-no-panic --strict`.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Sync/archive the OpenSpec change after verification.
