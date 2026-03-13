## 1. Tool Output Truncation Middleware

- [x] 1.1 Create `internal/toolchain/mw_truncate.go` with `WithTruncate(maxChars int) Middleware`
- [x] 1.2 Create `internal/toolchain/mw_truncate_test.go` with table-driven tests (under/over limit, map, error, nil, default)
- [x] 1.3 Add `MaxOutputChars int` field to `ToolsConfig` in `internal/config/types.go`
- [x] 1.4 Wire truncation middleware at step 7a in `internal/app/app.go` (before hooks at 7b)

## 2. Response Sanitizer

- [x] 2.1 Create `internal/gatekeeper/sanitizer.go` with `Sanitizer` struct and `NewSanitizer`/`Sanitize`/`Enabled` methods
- [x] 2.2 Implement thought tag stripping with code block protection (placeholder-based)
- [x] 2.3 Implement internal marker line removal (`[INTERNAL]`, `[DEBUG]`, `[SYSTEM]`, `[OBSERVATION]`)
- [x] 2.4 Implement large JSON code block replacement (configurable threshold, default 500)
- [x] 2.5 Implement custom regex pattern application and blank line normalization
- [x] 2.6 Create `internal/gatekeeper/sanitizer_test.go` with comprehensive test coverage
- [x] 2.7 Add `GatekeeperConfig` struct to `internal/config/types.go` with `*bool` toggles
- [x] 2.8 Add `Gatekeeper GatekeeperConfig` field to root `Config` struct

## 3. App Wiring

- [x] 3.1 Add `Sanitizer *gatekeeper.Sanitizer` field to `App` struct in `internal/app/types.go`
- [x] 3.2 Initialize sanitizer at step 1b in `internal/app/app.go`
- [x] 3.3 Wire sanitizer to gateway via `SetSanitizer()` after gateway creation
- [x] 3.4 Apply sanitization in `runAgent()` in `internal/app/channels.go` before returning response

## 4. Gateway Integration

- [x] 4.1 Add `sanitizer` field and `SetSanitizer()` method to `gateway.Server`
- [x] 4.2 Apply sanitization to streaming chunks in `handleChatMessage()` callback
- [x] 4.3 Suppress empty chunks after sanitization
- [x] 4.4 Apply sanitization to final response before returning

## 5. System Prompt Output Principles

- [x] 5.1 Create `prompts/OUTPUT_PRINCIPLES.md` with 6 output principles
- [x] 5.2 Add `SectionOutputPrinciples` to `internal/prompt/section.go` (Valid + Values)
- [x] 5.3 Add section to `DefaultBuilder()` in `internal/prompt/defaults.go` at priority 350
- [x] 5.4 Add `OUTPUT_PRINCIPLES.md` to `sectionFiles` map in `internal/prompt/loader.go`
- [x] 5.5 Update existing tests in `defaults_test.go` and `sections_test.go` for new section

## 6. Verification

- [x] 6.1 Full project build passes (`go build ./...`)
- [x] 6.2 All toolchain tests pass (`go test ./internal/toolchain/...`)
- [x] 6.3 All gatekeeper tests pass (`go test ./internal/gatekeeper/...`)
- [x] 6.4 All prompt tests pass (`go test ./internal/prompt/...`)
