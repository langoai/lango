## 1. Core Approval Interface

- [x] 1.1 Create `internal/approval/approval.go` — ApprovalRequest struct and Provider interface
- [x] 1.2 Create `internal/approval/composite.go` — CompositeProvider with session-key routing, TTY fallback, fail-closed
- [x] 1.3 Create `internal/approval/tty.go` — TTYProvider with stdin terminal prompt
- [x] 1.4 Create `internal/approval/gateway.go` — GatewayProvider with GatewayApprover interface
- [x] 1.5 Create `internal/approval/approval_test.go` — Tests for routing, fallback, fail-closed, concurrency

## 2. Wiring Migration

- [x] 2.1 Add `ApprovalTimeoutSec` field to `InterceptorConfig` in `internal/config/types.go`
- [x] 2.2 Add `ApprovalProvider approval.Provider` field to `App` struct in `internal/app/types.go`
- [x] 2.3 Refactor `wrapWithApproval` in `internal/app/tools.go` — change signature from `*gateway.Server` to `approval.Provider`, delete `requestToolApproval` and `promptTTYApproval`
- [x] 2.4 Wire CompositeProvider in `internal/app/app.go` — create composite, register GatewayProvider, set TTY fallback

## 3. Telegram Approval

- [x] 3.1 Add `Request` method to BotAPI interface in `internal/channels/telegram/bot.go`
- [x] 3.2 Create `internal/channels/telegram/approval.go` — ApprovalProvider with InlineKeyboard, HandleCallback, CanHandle
- [x] 3.3 Modify event loop in `internal/channels/telegram/telegram.go` — route CallbackQuery to approval provider, add GetApprovalProvider method
- [x] 3.4 Create `internal/channels/telegram/approval_test.go` — Tests for approve, deny, timeout, context cancellation
- [x] 3.5 Update MockBotAPI in `telegram_test.go` — add Request method

## 4. Discord Approval

- [x] 4.1 Add `InteractionRespond` and `ChannelMessageEditComplex` methods to Session interface in `internal/channels/discord/session.go`
- [x] 4.2 Create `internal/channels/discord/approval.go` — ApprovalProvider with Button components, HandleInteraction, CanHandle
- [x] 4.3 Modify `internal/channels/discord/discord.go` — register InteractionCreate handler, add GetApprovalProvider method
- [x] 4.4 Create `internal/channels/discord/approval_test.go` — Tests for approve, deny, timeout
- [x] 4.5 Update MockSession in `discord_test.go` — add InteractionRespond and ChannelMessageEditComplex methods

## 5. Slack Approval

- [x] 5.1 Add `UpdateMessage` method to Client interface in `internal/channels/slack/client.go`
- [x] 5.2 Create `internal/channels/slack/approval.go` — ApprovalProvider with Block Kit actions, HandleInteractive, CanHandle
- [x] 5.3 Modify `internal/channels/slack/slack.go` — handle EventTypeInteractive, add GetApprovalProvider method
- [x] 5.4 Create `internal/channels/slack/approval_test.go` — Tests for approve, deny, timeout
- [x] 5.5 Update MockClient in `slack_test.go` — add UpdateMessage method

## 6. Channel Wiring

- [x] 6.1 Modify `internal/app/channels.go` — register each channel's ApprovalProvider with CompositeProvider after channel creation

## 7. Verification

- [x] 7.1 Run `go build ./...` — verify full project builds
- [x] 7.2 Run `go test ./internal/approval/...` — core interface tests pass
- [x] 7.3 Run `go test ./internal/channels/telegram/...` — Telegram approval tests pass
- [x] 7.4 Run `go test ./internal/channels/discord/...` — Discord approval tests pass
- [x] 7.5 Run `go test ./internal/channels/slack/...` — Slack approval tests pass
- [x] 7.6 Run `go test ./internal/app/...` — wiring and tool wrapping tests pass
- [x] 7.7 Run `go vet ./...` — static analysis pass
