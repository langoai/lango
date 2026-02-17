## 1. Prompt Builder Extensions

- [x] 1.1 Add `SectionAgentIdentity` constant to `internal/prompt/section.go`
- [x] 1.2 Add `Clone()` method to `internal/prompt/builder.go`
- [x] 1.3 Add Clone tests to `internal/prompt/builder_test.go`

## 2. Per-Agent Loader

- [x] 2.1 Add `agentSectionFiles` map and `LoadAgentFromDir()` to `internal/prompt/loader.go`
- [x] 2.2 Add LoadAgentFromDir tests to `internal/prompt/loader_test.go`

## 3. Orchestration Integration

- [x] 3.1 Add `SubAgentPromptFunc` type to `internal/orchestration/orchestrator.go`
- [x] 3.2 Add `SubAgentPrompt` field to orchestration `Config`
- [x] 3.3 Use `SubAgentPromptFunc` in `BuildAgentTree` when non-nil
- [x] 3.4 Add SubAgentPromptFunc tests to `internal/orchestration/orchestrator_test.go`

## 4. App Wiring

- [x] 4.1 Add `buildSubAgentPromptFunc()` to `internal/app/wiring.go`
- [x] 4.2 Wire `SubAgentPrompt` into orchestration Config in `initAgent`

## 5. UI & Documentation

- [x] 5.1 Update Prompts Directory placeholder in `internal/cli/settings/forms_impl.go`
- [x] 5.2 Add "Per-Agent Prompt Customization" section to README.md

## 6. Verification

- [x] 6.1 Run `go build ./...` — build passes
- [x] 6.2 Run `go test ./internal/prompt/...` — all prompt tests pass
- [x] 6.3 Run `go test ./internal/orchestration/...` — all orchestration tests pass
- [x] 6.4 Run `go test ./...` — full test suite passes
