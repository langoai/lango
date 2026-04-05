## Why

Cockpit v1 is local-chat only (`AppModeLocalChat`), blind to external channel activity. Operators running Slack/Discord/Telegram channels alongside the cockpit cannot see inbound messages, approval origins, or channel connection status without switching to `lango serve` logs.

## What Changes

- Introduce `AppModeCockpit` application mode that optionally initializes channel adapters (when `--with-channels` flag is set) while skipping the HTTP gateway
- Add `ChannelMessageReceivedEvent` and `ChannelMessageSentEvent` to EventBus so channel activity is observable in-process
- Publish EventBus events from all three channel handlers (Telegram, Discord, Slack) before/after agent execution
- Add `Name() string` to the `Channel` interface for runtime identification
- Add `itemChannel` transcript kind with per-channel colored badge renderer, ANSI-sanitized external text
- Create EventBus-to-TUI bridge (`SubscribeChannelEvents`) that converts channel events into `tea.Msg`
- Show channel origin info on approval requests (banner/strip/dialog) via `formatChannelOrigin` helper
- Add channel connection status section to cockpit context panel via `ChannelTracker`
- Default cockpit to `--with-channels=false` to prevent duplicate bot sessions when `lango serve` is already running

## Capabilities

### New Capabilities
- `tui-channel-transcript`: Channel message display in cockpit transcript with per-channel badge, sender, sanitized text
- `cockpit-channel-status`: Channel connection status and message count tracking in context panel

### Modified Capabilities
- `cockpit-shell`: Cockpit now intercepts `ChannelMessageMsg` and always forwards to chat child regardless of active page
- `channel-approval`: Approval surfaces (banner/strip/dialog) now display channel origin info parsed from session key
- `cockpit-context-panel`: Context panel gains a "Channels" section showing connection status and message counts

## Impact

- `internal/app/types.go`: New `AppModeCockpit` mode + `WithCockpit()` option + `Name()` on Channel interface
- `internal/app/app.go`: Mode checks for cockpit channel initialization
- `internal/app/channels.go`: EventBus publishing in all 3 handlers
- `internal/channels/{telegram,discord,slack}`: `Name()` method added to each adapter
- `internal/eventbus/events.go`: 2 new event types
- `internal/cli/chat/`: New message type, transcript kind, channel renderer, approval origin helpers
- `internal/cli/cockpit/`: Channel bridge, tracker, context panel section, cockpit routing fix
- `cmd/lango/main.go`: `--with-channels` flag, mode switch, channel lifecycle management
