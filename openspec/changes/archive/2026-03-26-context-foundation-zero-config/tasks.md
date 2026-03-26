## 1. Config Loader: LoadResult + ExplicitKeys

- [x] 1.1 Add `LoadResult` struct and `collectExplicitKeys()` to `internal/config/loader.go`; change `Load()` return type to `(*LoadResult, error)`
- [x] 1.2 Update all `config.Load()` callers to use `result.Config`: only `internal/configstore/migrate.go` (others use bootstrap, not config.Load directly)
- [x] 1.3 Add unit tests for `collectExplicitKeys()`: verify only file-present keys are returned, verify nil when no config file, verify SetDefault-only keys are excluded

## 2. Context Profile System

- [x] 2.1 Create `internal/config/context_profile.go` with `ContextProfileName` type, constants (`Off`/`Lite`/`Balanced`/`Full`), and `ApplyContextProfile(cfg *Config, explicitKeys map[string]bool)`
- [x] 2.2 Add `ContextProfile ContextProfileName` field to `Config` in `internal/config/types.go` with mapstructure tag `contextProfile`
- [x] 2.3 Wire `ApplyContextProfile` call in `Load()` after Unmarshal, before PostLoad
- [x] 2.4 Add profile validation in `Validate()`: reject unknown profile names
- [x] 2.5 Add table-driven tests: each profile sets correct downstream booleans, explicit overrides are preserved, empty profile is no-op, invalid profile fails validation

## 3. FeatureStatus Shared Type

- [x] 3.1 Create `internal/types/feature_status.go` with `FeatureStatus` struct (Name, Enabled, Healthy, Reason, Suggestion)
- [x] 3.2 Create `internal/app/wiring_status.go` with `StatusCollector` (Add, All, SilentDisabledCount)
- [x] 3.3 Add unit tests for `StatusCollector`: verify Add, All, SilentDisabledCount logic

## 4. Wiring Functions: FeatureStatus Return

- [x] 4.1 Update `initEmbedding()` in `internal/app/wiring_embedding.go` to return `(*embeddingComponents, *types.FeatureStatus)` and populate status on every exit path
- [x] 4.2 Update `initKnowledge()` in `internal/app/wiring_knowledge.go` to return `(*knowledgeComponents, *types.FeatureStatus)`
- [x] 4.3 Update memory init in `internal/app/wiring_memory.go` (or equivalent) to return `*types.FeatureStatus`
- [x] 4.4 Update `initGraphStore()` in `internal/app/wiring_graph.go` to return `*types.FeatureStatus`
- [x] 4.5 Update librarian init to return `*types.FeatureStatus`
- [x] 4.6 Wire all returned statuses into `StatusCollector` in the `app.New()` orchestration

## 5. Doctor Check: Context Engineering

- [x] 5.1 Create `internal/cli/doctor/checks/context_health.go` implementing `ContextHealthCheck` with Name/Run/Fix methods
- [x] 5.2 Register `ContextHealthCheck` in `AllChecks()` in `internal/cli/doctor/checks/checks.go` before Embedding/Graph/Memory checks
- [x] 5.3 Create `internal/cli/doctor/checks/adapter.go` with `FeatureStatusToDoctorResult()` adapter function
- [x] 5.4 Add unit tests: balanced profile with silent disabled, all healthy, no profile set

## 6. Status Command: Profile + Reason

- [x] 6.1 Create `internal/cli/status/adapter.go` with `FeatureStatusToFeatureInfo()` adapter function
- [x] 6.2 Update `collectFeatures()` in `internal/cli/status/status.go` to include profile detail and use FeatureStatus reason for context-related features where available
- [x] 6.3 Add profile name display to the status dashboard output

## 7. Build Verification + Downstream

- [x] 7.1 Run `go build ./...` and fix all compilation errors from Load() signature change
- [x] 7.2 Run `go test ./...` and fix all test failures
- [x] 7.3 Run `go vet ./...` and fix any warnings
- [x] 7.4 Verify `lango doctor` shows "Context Engineering" check
- [x] 7.5 Verify `lango status` shows profile name and feature reasons
