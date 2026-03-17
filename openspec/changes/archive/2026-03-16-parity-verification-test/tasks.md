## 1. Lifecycle Registry Enhancement

- [x] 1.1 Add `Names()` method to `internal/lifecycle/registry.go`
- [x] 1.2 Add `TestRegistry_Names` and `TestRegistry_Names_Empty` to `internal/lifecycle/registry_test.go`
- [x] 1.3 Verify lifecycle package builds and tests pass

## 2. Layer 1 — Helper Unit Tests

- [x] 2.1 Create `internal/app/parity_test.go` with `TestBuildCatalogFromEntries_Basic`
- [x] 2.2 Add `TestBuildCatalogFromEntries_DuplicateCategory` for same-category accumulation
- [x] 2.3 Add `TestRegisterPostBuildLifecycle_Names` with no-channels and with-channels subtests
- [x] 2.4 Verify Layer 1 tests pass

## 3. Layer 2 — Integration Parity Tests

- [x] 3.1 Add `TestAppNew_DefaultConfig_Parity` — fixture with `DefaultConfig` + `TestEntClient`, verify enabled/disabled categories, tool count, dispatcher tools, lifecycle names, non-nil/nil fields
- [x] 3.2 Add `TestAppNew_FeaturesEnabled_Parity` — fixture with knowledge/graph/memory/cron enabled, verify additional categories, lifecycle components, and field population
- [x] 3.3 Verify Layer 2 tests pass with `go test ./internal/app/... -run TestAppNew -v`

## 4. Final Verification

- [x] 4.1 Run full app package test suite: `go test ./internal/app/... -count=1`
- [x] 4.2 Run full lifecycle package test suite: `go test ./internal/lifecycle/... -count=1`
