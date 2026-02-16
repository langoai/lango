## 1. Prompt Package Core

- [x] 1.1 Create `internal/prompt/section.go` with SectionID constants and PromptSection interface
- [x] 1.2 Create `internal/prompt/sections.go` with StaticSection implementation (ID, Priority, Render)
- [x] 1.3 Create `internal/prompt/builder.go` with Builder (Add, Remove, Has, Build with priority sorting)
- [x] 1.4 Create `internal/prompt/sections_test.go` verifying Render with/without title, empty content
- [x] 1.5 Create `internal/prompt/builder_test.go` verifying Add/Remove/Replace/Priority/Empty

## 2. Default Sections and Loader

- [x] 2.1 Create `internal/prompt/defaults.go` with four default sections (Identity, Safety, ConversationRules, ToolUsage) and DefaultBuilder()
- [x] 2.2 Create `internal/prompt/loader.go` with LoadFromDir (known file mapping, custom section handling, error fallback)
- [x] 2.3 Create `internal/prompt/defaults_test.go` verifying all sections present and correct order
- [x] 2.4 Create `internal/prompt/loader_test.go` verifying override, custom files, empty files, non-existent dir

## 3. Config Extension

- [x] 3.1 Add `PromptsDir` field to `AgentConfig` in `internal/config/types.go`

## 4. Wiring Integration

- [x] 4.1 Replace `_defaultSystemPrompt` and `loadSystemPrompt()` with `buildPromptBuilder()` in `internal/app/wiring.go`
- [x] 4.2 Update `ContextAwareModelAdapter` constructor to accept `*prompt.Builder` in `internal/adk/context_model.go`
- [x] 4.3 Update all call sites in `initAgent()` to pass builder instead of string

## 5. Verification

- [x] 5.1 Run `go build ./...` — no errors
- [x] 5.2 Run `go test ./internal/prompt/...` — all tests pass
- [x] 5.3 Run `go test ./...` — full test suite passes
