## 1. Config Layer

- [x] 1.1 Add `ProviderID` field to `EmbeddingConfig` in `internal/config/types.go`
- [x] 1.2 Add exported `ProviderTypeToEmbeddingType` mapping (provider type → embedding backend type)
- [x] 1.3 Implement `ResolveEmbeddingProvider()` method on `Config` with ProviderID priority, legacy fallback, and unsupported type handling
- [x] 1.4 Add table-driven tests for `ResolveEmbeddingProvider` covering ProviderID, legacy, precedence, and edge cases in `internal/config/types_test.go`

## 2. Wiring

- [x] 2.1 Update `initEmbedding` in `internal/app/wiring.go` to use `ResolveEmbeddingProvider()` instead of hardcoded switch/map lookups
- [x] 2.2 Update disabled check to consider both `Provider` and `ProviderID` empty
- [x] 2.3 Update log output to include `providerID` and use resolved `backendType`

## 3. Onboard TUI

- [x] 3.1 Update `NewEmbeddingForm` in `internal/cli/onboard/forms_impl.go` to build provider options from user's registered providers + "local"
- [x] 3.2 Add `emb_provider_id` handler in `internal/cli/onboard/state_update.go` with auto-resolution of provider type
- [x] 3.3 Update embedding form test in `forms_impl_test.go` to expect `emb_provider_id` key and verify dynamic options
- [x] 3.4 Add test for `emb_provider_id` local selection clearing ProviderID

## 4. Doctor Checks

- [x] 4.1 Update `EmbeddingCheck.Run` in `internal/cli/doctor/checks/embedding.go` to use `ResolveEmbeddingProvider()`
- [x] 4.2 Add tests in `embedding_test.go` for ProviderID resolution, not-found, no-key, legacy, local, and unconfigured cases

## 5. Documentation

- [x] 5.1 Add `embedding.providerID` row to Configuration Reference table in `README.md`
