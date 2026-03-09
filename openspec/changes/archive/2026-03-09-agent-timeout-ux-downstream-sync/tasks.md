## 1. Documentation Updates

- [x] 1.1 Add `agent.autoExtendTimeout` and `agent.maxRequestTimeout` rows to README.md config table after `agent.agentsDir`
- [x] 1.2 Add `autoExtendTimeout` and `maxRequestTimeout` to docs/configuration.md JSON example
- [x] 1.3 Add `agent.autoExtendTimeout` and `agent.maxRequestTimeout` rows to docs/configuration.md config table
- [x] 1.4 Add `agent.progress`, `agent.warning`, `agent.error` events to docs/gateway/websocket.md events table
- [x] 1.5 Add progressive thinking item to docs/features/channels.md Channel Features list

## 2. TUI Settings

- [x] 2.1 Add `auto_extend_timeout` (InputBool) and `max_request_timeout` (InputText) fields to NewAgentForm in forms_impl.go
- [x] 2.2 Add `auto_extend_timeout` and `max_request_timeout` case handlers to UpdateConfigFromForm in state_update.go
- [x] 2.3 Update TestNewAgentForm_AllFields wantKeys to include new field keys in forms_impl_test.go

## 3. Verification

- [x] 3.1 Run `go build ./...` to verify no build errors
- [x] 3.2 Run `go test ./internal/cli/settings/...` to verify TUI form tests pass
- [x] 3.3 Run `go test ./internal/cli/tuicore/...` to verify state update tests pass
