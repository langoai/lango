## 1. Dead Code Removal

- [x] 1.1 Delete `NewSQLiteStore` wrapper function from `internal/session/store.go`

## 2. ApprovalRequired Legacy Removal

- [x] 2.1 Remove `ApprovalRequired` field from `InterceptorConfig` in `internal/config/types.go`
- [x] 2.2 Remove `migrateApprovalPolicy()` function and its call in `internal/config/loader.go`
- [x] 2.3 Remove `TestMigrateApprovalPolicy` from `internal/config/migrate_test.go`; keep `TestDefaultConfig_ApprovalPolicy`

## 3. SystemPromptPath Legacy Removal

- [x] 3.1 Remove `SystemPromptPath` field from `AgentConfig` in `internal/config/types.go`
- [x] 3.2 Remove legacy single-file prompt fallback from `buildPromptBuilder()` in `internal/app/wiring.go`
- [x] 3.3 Remove `system_prompt_path` form field from `NewAgentForm()` in `internal/cli/onboard/forms_impl.go`
- [x] 3.4 Remove `system_prompt_path` case from `UpdateConfigFromForm()` in `internal/cli/onboard/state_update.go`
- [x] 3.5 Update tests in `forms_impl_test.go` to remove `system_prompt_path` references

## 4. EmbeddingConfig.Provider Legacy Cleanup

- [x] 4.1 Update `Provider` field comment in `EmbeddingConfig` to document local-only usage
- [x] 4.2 Simplify `ResolveEmbeddingProvider()` — remove type-search fallback, keep only ProviderID and local paths
- [x] 4.3 Update embedding form in `forms_impl.go` — restrict Provider fallback to local only
- [x] 4.4 Remove backward-compatibility auto-resolve in `state_update.go` `emb_provider_id` case
- [x] 4.5 Update doctor check in `internal/cli/doctor/checks/embedding.go` — remove legacy Provider references
- [x] 4.6 Update tests: `types_test.go`, `embedding_test.go`, `forms_impl_test.go`

## 5. EventsAdapter Legacy Fallback

- [x] 5.1 Add `DefaultTokenBudget = 32000` constant; use it when `tokenBudget <= 0`
- [x] 5.2 Remove 100-message hardcap from `truncatedHistory()` in `internal/adk/state.go`
- [x] 5.3 Update `TestEventsAdapter_Truncation` and `TestEventsAdapter_LegacyFallback` in `state_test.go`
- [x] 5.4 Update legacy cap comment in `internal/memory/integration_test.go`

## 6. Comment Cleanup

- [x] 6.1 Remove misleading backward-compatibility comment from `GetSalt()` in `internal/session/ent_store.go`

## 7. Verification

- [x] 7.1 `go build ./...` succeeds
- [x] 7.2 `go test ./...` passes
- [x] 7.3 Grep confirms no remaining references to removed symbols in `internal/`
