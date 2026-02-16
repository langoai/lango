## 1. Session Context Helpers (P0)

- [x] 1.1 Create `internal/session/context.go` with `WithSessionKey` and `SessionKeyFromContext`
- [x] 1.2 Remove `sessionKeyCtxKey`, `WithSessionKey`, `SessionKeyFromContext` from `internal/app/tools.go`
- [x] 1.3 Update `internal/app/tools.go` references to use `session.SessionKeyFromContext`
- [x] 1.4 Update `internal/app/channels.go` to use `session.WithSessionKey`

## 2. Context Propagation Fixes (P0)

- [x] 2.1 Add `ctx context.Context` field to Discord `Channel` struct
- [x] 2.2 Store `Start(ctx)` context in `Channel.ctx` field
- [x] 2.3 Replace `context.Background()` with `c.ctx` in `onMessageCreate`
- [x] 2.4 Inject session key into Gateway `handleChatMessage` context via `session.WithSessionKey`

## 3. Silent Failure Fixes (P0)

- [x] 3.1 Change `TTYProvider.RequestApproval` to return error when stdin is not a terminal
- [x] 3.2 Create `internal/approval/tty_test.go` with non-terminal error test

## 4. Safety Improvements (P1)

- [x] 4.1 Add comma-ok guard to Discord `HandleInteraction` type assertion
- [x] 4.2 Add comma-ok guard to Telegram `HandleCallback` type assertion
- [x] 4.3 Add comma-ok guard to Slack `HandleInteractive` type assertions (both Load and LoadAndDelete)
- [x] 4.4 Replace `_ = store.SaveAuditLog` with error logging in `save_knowledge` handler
- [x] 4.5 Replace `_ = store.SaveAuditLog` with error logging in `save_learning` handler
- [x] 4.6 Replace `_ = store.SaveAuditLog` with error logging in `create_skill` handler
- [x] 4.7 Add `default` case with warning log to Anthropic `convertParams` role switch

## 5. Verification

- [x] 5.1 `go build ./...` passes
- [x] 5.2 `go test ./...` passes
- [x] 5.3 `go vet ./...` passes
