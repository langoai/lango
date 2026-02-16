## 1. Core Types

- [x] 1.1 Add SafetyLevel enum type with iota+1 to `internal/agent/runtime.go`
- [x] 1.2 Add String() and IsDangerous() methods to SafetyLevel
- [x] 1.3 Add SafetyLevel field to Tool struct
- [x] 1.4 Add SafetyLevel unit tests (`internal/agent/safety_level_test.go`)

## 2. Config Types and Migration

- [x] 2.1 Add ApprovalPolicy type and constants to `internal/config/types.go`
- [x] 2.2 Add ExemptTools field to InterceptorConfig
- [x] 2.3 Mark ApprovalRequired as deprecated in InterceptorConfig
- [x] 2.4 Add migrateApprovalPolicy function to `internal/config/loader.go`
- [x] 2.5 Update DefaultConfig with Interceptor defaults (Enabled=true, Policy=dangerous)
- [x] 2.6 Add viper SetDefault for approvalPolicy
- [x] 2.7 Add migration unit tests (`internal/config/migrate_test.go`)

## 3. Approval Request Enhancement

- [x] 3.1 Add Summary field to ApprovalRequest in `internal/approval/approval.go`

## 4. Tool SafetyLevel Assignments

- [x] 4.1 Add SafetyLevel to all exec tools (exec=Dangerous, exec_bg=Dangerous, exec_status=Safe, exec_stop=Dangerous)
- [x] 4.2 Add SafetyLevel to all filesystem tools (fs_read=Safe, fs_list=Safe, fs_write=Dangerous, fs_edit=Dangerous, fs_mkdir=Moderate, fs_delete=Dangerous)
- [x] 4.3 Add SafetyLevel to all browser tools (browser_navigate=Dangerous, browser_action=Dangerous, browser_screenshot=Safe)
- [x] 4.4 Add SafetyLevel to all crypto tools (crypto_encrypt=Dangerous, crypto_decrypt=Dangerous, crypto_sign=Dangerous, crypto_hash=Safe, crypto_keys=Safe)
- [x] 4.5 Add SafetyLevel to all secrets tools (secrets_store=Dangerous, secrets_get=Dangerous, secrets_list=Safe, secrets_delete=Dangerous)
- [x] 4.6 Add SafetyLevel to all meta tools (save_knowledge=Moderate, search_knowledge=Safe, save_learning=Moderate, search_learnings=Safe, create_skill=Moderate, list_skills=Safe)

## 5. Approval Logic Refactoring

- [x] 5.1 Add needsApproval(tool, interceptorConfig) function in `internal/app/tools.go`
- [x] 5.2 Add buildApprovalSummary(toolName, params) function with per-tool summaries
- [x] 5.3 Add truncate helper function
- [x] 5.4 Refactor wrapWithApproval to accept InterceptorConfig and use needsApproval
- [x] 5.5 Preserve SafetyLevel in wrapWithApproval wrapper
- [x] 5.6 Preserve SafetyLevel in wrapWithLearning wrapper
- [x] 5.7 Add needsApproval and buildApprovalSummary unit tests (`internal/app/approval_test.go`)

## 6. Application Wiring

- [x] 6.1 Update approval gate in `internal/app/app.go` to use policy-based logic

## 7. Summary Rendering in Providers

- [x] 7.1 Add Summary to GatewayProvider message in `internal/approval/gateway.go`
- [x] 7.2 Add Summary to TTYProvider prompt in `internal/approval/tty.go`
- [x] 7.3 Add Summary to HeadlessProvider audit log in `internal/approval/headless.go`
- [x] 7.4 Add Summary to Telegram message in `internal/channels/telegram/approval.go`
- [x] 7.5 Add Summary to Discord message (code block) in `internal/channels/discord/approval.go`
- [x] 7.6 Add Summary to Slack Block Kit text in `internal/channels/slack/approval.go`

## 8. Verification

- [x] 8.1 Run `go build ./...` — verify no compile errors
- [x] 8.2 Run `go test ./...` — verify all tests pass
