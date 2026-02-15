## Why

Sensitive tool execution approval currently relies solely on the Gateway Companion WebSocket. Users interacting through Telegram, Discord, or Slack channels cannot participate in the approval process — if no companion is connected, tool execution is unconditionally denied. This blocks channel users from using sensitive tools entirely.

## What Changes

- Introduce a unified `approval.Provider` interface in `internal/approval/` that abstracts approval request handling across all channels.
- Implement `CompositeProvider` that routes approval requests to the correct channel based on session key prefix, with TTY fallback and fail-closed semantics.
- Extract existing TTY prompt logic into `TTYProvider` and gateway logic into `GatewayProvider`.
- Implement channel-native approval providers:
  - **Telegram**: InlineKeyboard buttons (approve/deny callbacks)
  - **Discord**: Message Component buttons (InteractionCreate handler)
  - **Slack**: Block Kit action buttons (EventTypeInteractive handler)
- Migrate `wrapWithApproval` to use the new `approval.Provider` interface instead of direct `gateway.Server` dependency.
- Wire channel approval providers into the composite at channel initialization.
- Add `ApprovalTimeoutSec` config field to `InterceptorConfig`.

## Capabilities

### New Capabilities
- `channel-approval`: Unified approval provider interface and channel-specific approval implementations (Telegram InlineKeyboard, Discord Buttons, Slack Block Kit)

### Modified Capabilities
- `ai-privacy-interceptor`: Tool approval now routes through CompositeProvider instead of gateway-only; adds ApprovalTimeoutSec config
- `channel-telegram`: Adds CallbackQuery handling for approval InlineKeyboard and `Request` method to BotAPI interface
- `channel-discord`: Adds InteractionCreate handler for approval buttons and `InteractionRespond`/`ChannelMessageEditComplex` to Session interface
- `channel-slack`: Adds EventTypeInteractive handling for approval block actions and `UpdateMessage` to Client interface

## Impact

- **Core package**: New `internal/approval/` package (no external dependencies)
- **App layer**: `wrapWithApproval` signature changed from `*gateway.Server` to `approval.Provider`; `App` struct gains `ApprovalProvider` field
- **Channel packages**: Each channel's SDK interface extended with 1-2 methods; event loops modified to route interactive callbacks
- **Config**: `InterceptorConfig` struct extended with `ApprovalTimeoutSec` field
- **Existing tests**: Channel mock types updated to satisfy new interface methods
