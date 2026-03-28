## 1. Config Core

- [x] 1.1 Create `internal/config/auto_enable.go` with `collectExplicitKeys`, `contextRelatedKeys`, `nestedKeyExists`, `ResolveContextAutoEnable`, `AutoEnabledSet`, `ProbeEmbeddingProvider`, `PresetExplicitKeys`
- [x] 1.2 Add `LoadResult` type to `internal/config/loader.go`
- [x] 1.3 Change `Load()` return type from `(*Config, error)` to `(*LoadResult, error)`
- [x] 1.4 Wire `collectExplicitKeys` → `ApplyContextProfile` → `ResolveContextAutoEnable` → `PostLoad` in `Load()` pipeline
- [x] 1.5 Add `contextProfile` validation to `Validate()`
- [x] 1.6 Create `internal/config/auto_enable_test.go` with auto-enable and provider probe tests

## 2. Configstore

- [x] 2.1 Add `profilePayload{Config, ExplicitKeys}` internal type to `internal/configstore/store.go`
- [x] 2.2 Update `Save()` to accept `explicitKeys` parameter, marshal as profilePayload
- [x] 2.3 Update `Load()` to return `(*Config, map[string]bool, error)` with legacy fallback
- [x] 2.4 Update `LoadActive()` to return `(string, *Config, map[string]bool, error)` with legacy fallback
- [x] 2.5 Add `decryptPayload()` helper for new/legacy format detection
- [x] 2.6 Update `configstore/store_test.go` for new signatures

## 3. Bootstrap

- [x] 3.1 Add `ExplicitKeys` and `AutoEnabled` fields to `bootstrap.Result`
- [x] 3.2 Update `phaseLoadProfile` to call `ApplyContextProfile` + `ResolveContextAutoEnable` after profile load
- [x] 3.3 Update `handleNoProfile` to pass nil explicitKeys to `Save()`

## 4. Types

- [x] 4.1 Add `AutoEnabled bool` field to `types.FeatureStatus`

## 5. Caller Updates

- [x] 5.1 Update `configstore/migrate.go` to use `result.Config` and `result.ExplicitKeys`
- [x] 5.2 Update `cmd/lango/main.go` Save call with nil explicitKeys
- [x] 5.3 Update `cli/settings/settings.go` Save/Load calls
- [x] 5.4 Update `cli/onboard/onboard.go` Save/Load calls
- [x] 5.5 Update `cli/configcmd/profile.go` Save/Load calls with PresetExplicitKeys

## 6. Verification

- [x] 6.1 `CGO_ENABLED=1 go build -tags fts5 ./...` — full build passes (0 errors)
- [x] 6.2 Config, configstore, bootstrap, types package tests pass
- [x] 6.3 Pre-existing `configstore/migrate.go` build error resolved
