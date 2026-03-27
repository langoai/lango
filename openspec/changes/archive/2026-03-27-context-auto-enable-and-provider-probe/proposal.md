## Why

All context subsystems (Knowledge, Memory, Retrieval) require explicit `cfg.X.Enabled = true`. Users with EntStore available but no explicit config get no knowledge features. The config loader also lacks explicit key tracking, preventing the system from distinguishing "user set false" from "default false".

## What Changes

- Implement `LoadResult{Config, ExplicitKeys, AutoEnabled}` return type for `config.Load()` (spec existed but was unimplemented)
- Implement `collectExplicitKeys()` to detect which config keys the user explicitly wrote
- Add `ResolveContextAutoEnable()` shared resolver — auto-enables Knowledge, Memory, Retrieval when deps detectable at config level and not explicitly disabled
- Add `ProbeEmbeddingProvider()` with conservative cost policy (local-first, single-remote-only)
- Add `FeatureStatus.AutoEnabled` field for diagnostics
- Store `explicitKeys` inside encrypted profile payload (`configstore.profilePayload`)
- Wire auto-enable resolution in both `config.Load()` and `bootstrap.phaseLoadProfile` paths
- Add `contextProfile` validation to `Validate()`
- Fix pre-existing `configstore/migrate.go` build error

## Capabilities

### New Capabilities
- `context-auto-enable`: ExplicitKeys tracking, auto-enable resolution, embedding provider probe

### Modified Capabilities
- `config-system`: `Load()` returns `*LoadResult`, `collectExplicitKeys` + `ApplyContextProfile` + `ResolveContextAutoEnable` in pipeline, `Validate()` checks contextProfile
- `feature-status`: `FeatureStatus.AutoEnabled` field added

## Impact

- **MODIFY**: `internal/config/loader.go` — LoadResult, Load() return type, Validate() contextProfile
- **MODIFY**: `internal/configstore/store.go` — profilePayload, Save/Load/LoadActive signatures
- **MODIFY**: `internal/configstore/migrate.go` — use LoadResult, pass ExplicitKeys
- **MODIFY**: `internal/bootstrap/bootstrap.go` — Result.ExplicitKeys, Result.AutoEnabled
- **MODIFY**: `internal/bootstrap/phases.go` — ApplyContextProfile + ResolveContextAutoEnable in phaseLoadProfile
- **MODIFY**: `internal/types/feature_status.go` — AutoEnabled field
- **MODIFY**: `cmd/lango/main.go`, `cli/settings/settings.go`, `cli/onboard/onboard.go`, `cli/configcmd/profile.go` — Save/Load caller updates
- **NEW**: `internal/config/auto_enable.go` — collectExplicitKeys, ResolveContextAutoEnable, ProbeEmbeddingProvider, AutoEnabledSet, PresetExplicitKeys
- **NEW**: `internal/config/auto_enable_test.go` — 22 test cases
