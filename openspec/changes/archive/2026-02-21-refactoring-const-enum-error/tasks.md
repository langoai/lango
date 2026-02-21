## 1. Foundation — Shared Types Package

- [x] 1.1 Create `internal/types/enum.go` with `Enum[T any]` interface (`Valid() bool`, `Values() []T`)
- [x] 1.2 Create `internal/types/callback.go` with `EmbedCallback`, `ContentCallback`, `Triple`, `TripleCallback` types
- [x] 1.3 Create `internal/types/token.go` with `EstimateTokens()` and `isCJK()` from `memory/token.go`
- [x] 1.4 Migrate `knowledge/store.go` and `memory/store.go` to import `EmbedCallback` from `types`
- [x] 1.5 Migrate `learning/graph_engine.go`, `memory/graph_hooks.go`, `librarian/types.go` to import `TripleCallback` from `types`
- [x] 1.6 Delete `learning/token.go` and update references to use `types.EstimateTokens`
- [x] 1.7 Update `memory/token.go` to delegate to `types.EstimateTokens`
- [x] 1.8 Rename `internal/cli/common/` to `internal/cli/clitypes/` and update 3 importers

## 2. Core Enum Types in `internal/types/`

- [x] 2.1 Create `internal/types/channel.go` — `ChannelType` with `ChannelTelegram`, `ChannelDiscord`, `ChannelSlack` + `Valid()`/`Values()`
- [x] 2.2 Create `internal/types/provider.go` — `ProviderType` with openai/anthropic/gemini/google/ollama/github + `Valid()`/`Values()`
- [x] 2.3 Create `internal/types/role.go` — `MessageRole` with user/assistant/tool/function/model + `Valid()`/`Values()`/`Normalize()`
- [x] 2.4 Create `internal/types/confidence.go` — `Confidence` with high/medium/low + `Valid()`/`Values()`
- [x] 2.5 Create `internal/types/sender.go` — `RPCSenderFunc` type consolidating duplicates

## 3. Enum Consumers — Replace Magic Strings

- [x] 3.1 Update `app/sender.go` to use `types.ChannelType` constants
- [x] 3.2 Update `app/channels.go` to use `types.ChannelType` constants
- [x] 3.3 Update `app/tools.go` to use `types.ChannelType` constants
- [x] 3.4 Update `cli/settings/forms_impl.go`, `cli/onboard/wizard.go`, `cli/onboard/steps.go` to use `types.ChannelType`
- [x] 3.5 Update `supervisor/supervisor.go` to switch on `types.ProviderType` constants
- [x] 3.6 Update `adk/session_service.go` to use `types.MessageRole` with `Normalize()`
- [x] 3.7 Update `librarian/types.go` and `learning/parse.go` to use `types.Confidence`
- [x] 3.8 Update `security/rpc_provider.go` and `wallet/rpc_wallet.go` to use `types.RPCSenderFunc`
- [x] 3.9 Update `config/types.go` `ProviderConfig.Type` field to `types.ProviderType`

## 4. Package-Local Enum Types

- [x] 4.1 Convert `graph/store.go` predicate constants to typed `Predicate string` + `Valid()`/`Values()`
- [x] 4.2 Create `knowledge.KnowledgeCategory` typed enum in `knowledge/types.go` + `Valid()`/`Values()`
- [x] 4.3 Create `skill.SkillStatus` and `skill.SkillType` typed enums in `skill/types.go` + `Valid()`/`Values()`

## 5. Add Valid/Values to Existing Enums

- [x] 5.1 Add `Valid()`/`Values()` to `ApprovalPolicy` in `config/types.go`
- [x] 5.2 Add `Valid()`/`Values()` to `PIICategory` in `agent/pii_pattern.go`
- [x] 5.3 Add `Valid()`/`Values()` to `KeyType` in `security/key_registry.go`
- [x] 5.4 Add `Valid()`/`Values()` to `SectionID` in `prompt/section.go`
- [x] 5.5 Add `Valid()`/`Values()` to `StreamEventType` in `provider/provider.go`
- [x] 5.6 Add `Valid()`/`Values()` to `SafetyLevel` in `agent/runtime.go`
- [x] 5.7 Add `Valid()`/`Values()` to `Status` in `background/task.go`
- [x] 5.8 Add `Valid()`/`Values()` to `ContextLayer` in `knowledge/types.go`

## 6. Sentinel Errors

- [x] 6.1 Create `session/errors.go` with `ErrSessionNotFound`, `ErrDuplicateSession`
- [x] 6.2 Update `adk/session_service.go` to use `errors.Is()` instead of `strings.Contains`
- [x] 6.3 Create `gateway/errors.go` with `ErrNoCompanion`, `ErrApprovalTimeout`, `ErrAgentNotReady`, `RPCError` type
- [x] 6.4 Update `gateway/server.go` to use `RPCError` named type
- [x] 6.5 Create `workflow/errors.go` with `ErrWorkflowNameEmpty`, `ErrNoWorkflowSteps`, `ErrStepIDEmpty`
- [x] 6.6 Create `knowledge/errors.go` with `ErrKnowledgeNotFound`, `ErrLearningNotFound`
- [x] 6.7 Create `security/errors.go` with `ErrKeyNotFound`, `ErrNoEncryptionKeys`, `ErrDecryptionFailed`

## 7. Error Message Cleanup — Remove "failed to" Prefix

- [x] 7.1 Fix `session/ent_store.go` (~30 occurrences)
- [x] 7.2 Fix `tools/filesystem/filesystem.go` (~12 occurrences)
- [x] 7.3 Fix `security/secrets_store.go` (~8 occurrences)
- [x] 7.4 Fix `security/key_registry.go` (~6 occurrences)
- [x] 7.5 Fix `security/local_provider.go` (~6 occurrences)
- [x] 7.6 Fix `adk/agent.go` (~5 occurrences)
- [x] 7.7 Fix `gateway/auth.go` (~4 occurrences)
- [x] 7.8 Fix `app/app.go` (~4 occurrences)
- [x] 7.9 Fix remaining 20+ files with "failed to" errors

## 8. Verification

- [x] 8.1 Run `go build ./...` — zero errors
- [x] 8.2 Run `go test ./...` — all tests pass
- [x] 8.3 Verify zero "failed to" remaining: `grep -r "failed to" internal/ --include="*.go" | grep -v _test.go | grep -v ent/`
