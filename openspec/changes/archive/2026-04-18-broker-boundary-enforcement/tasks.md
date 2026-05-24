## 1. Architecture Gates

- [x] 1.1 Add archtests that forbid removed raw storage accessors in production packages
- [x] 1.2 Add archtests that forbid payment-side `entStore.Client()` regressions and test-only storage wiring helpers in production packages

## 2. Payment Boundary Cleanup

- [x] 2.1 Remove the remaining production payment fallback that reconstructs transaction storage from `session.EntStore.Client()`
- [x] 2.2 Update docs to reflect boundary enforcement and storage-facing payment setup

## 3. Verification

- [x] 3.1 Verify broker transport regression tests remain covered under `go test ./...`
- [x] 3.2 Run `go build ./...`, `go test ./...`, and `openspec validate --type change broker-boundary-enforcement`
