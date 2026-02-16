## 1. Session Key Context Propagation

- [x] 1.1 Add `ctx = WithSessionKey(ctx, sessionKey)` in `runAgent()` at `internal/app/channels.go`

## 2. CompositeProvider Sentinel Error

- [x] 2.1 Change fail-closed path in `CompositeProvider.RequestApproval` to return `fmt.Errorf("no approval provider for session %q", req.SessionKey)` at `internal/approval/composite.go`
- [x] 2.2 Update `TestCompositeProvider_FailClosed` to expect non-nil error at `internal/approval/approval_test.go`

## 3. Approval Error Message Improvement

- [x] 3.1 Add session key check in `wrapWithApproval` deny branch to differentiate "no channel" vs "user denied" at `internal/app/tools.go`

## 4. System Prompt Guidance

- [x] 4.1 Add "Tool Approval" section to `prompts/TOOL_USAGE.md` with guidance for denial and missing channel scenarios

## 5. Verification

- [x] 5.1 Run `go build ./...` and verify no build errors
- [x] 5.2 Run `go test ./...` and verify all tests pass
