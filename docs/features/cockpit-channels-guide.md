---
title: Cockpit Channel Integration Guide
---

# Cockpit Channel Integration Guide

An operator's guide for using Telegram, Discord, and Slack channels within the cockpit TUI.

For full channel setup instructions (bot tokens, configuration keys, security settings), see [Channels](channels.md).

## Overview

The cockpit TUI can operate as a live channel operator console. When launched with the `--with-channels` flag, the cockpit starts channel adapters (Telegram, Discord, Slack) alongside the local chat interface. Inbound messages from all connected channels appear in the Chat page transcript, and approval requests from remote channel users surface directly in the cockpit for the operator to act on.

This integration uses the EventBus to bridge channel events into the Bubble Tea TUI event loop, so all channel activity is visible without polling or page-switching.

## Launching with Channels

```bash
lango cockpit --with-channels
```

Without the `--with-channels` flag, the cockpit runs in local-chat mode (same as `lango chat` but with the multi-panel layout). Channel adapters are not started, and no channel events appear.

The `--with-channels` flag uses `AppModeCockpit`, which starts core components plus channel adapters but skips the gateway server, automation, and network lifecycle. This avoids conflicts when a separate `lango serve` instance is already running with the same credentials.

**Important**: Do not run `lango cockpit --with-channels` and `lango serve` simultaneously with the same channel credentials. Both would attempt to poll/connect to the same bot tokens, causing duplicate message handling or API conflicts.

## Channel Setup

Each channel requires credentials configured via `lango settings` or `lango onboard`. A brief summary:

| Channel | Required Credentials |
|---------|---------------------|
| **Telegram** | `botToken` from BotFather |
| **Discord** | `botToken` + `applicationId` from Discord Developer Portal |
| **Slack** | `botToken` + `appToken` + `signingSecret` from Slack API |

For detailed setup instructions, prerequisite steps, allowlist configuration, and security recommendations, see [Channels -- Setup](channels.md#setup).

## Message Routing

Channel messages flow through the following path:

1. A channel adapter (Telegram/Discord/Slack) receives an inbound message
2. The adapter publishes a `ChannelMessageReceivedEvent` on the EventBus
3. `SubscribeChannelEvents` (wired in `runCockpit()`) converts the event to a `ChannelMessageMsg` and sends it to the Bubble Tea program
4. The cockpit's `Update` handler forwards `ChannelMessageMsg` directly to the chat child model -- **regardless of which page is currently active**
5. The chat model appends the message to the transcript via `appendChannel`

This design ensures that channel messages are never lost when the operator is browsing a non-chat page (Settings, Tools, Status, Tasks, etc.). The message is always recorded in the chat transcript and will be visible when the operator switches back to the Chat page.

### Channel Message Display

Each channel message renders as a single-line block in the chat transcript with:

- A **colored channel badge** (Telegram blue `#0088cc`, Discord blurple `#5865F2`, Slack aubergine `#4A154B`)
- The **sender name** prefixed with `@` (when available)
- The **message text**, sanitized (ANSI escape sequences stripped, newlines collapsed) and truncated to fit the terminal width

External input is sanitized before rendering to prevent terminal control injection from remote users.

## Channel Approval Flow

When a remote channel user triggers a tool that requires approval, the request flows through the cockpit:

1. The channel adapter's approval handler creates an `ApprovalRequestMsg` and sends it to the TUI program
2. The cockpit receives `ApprovalRequestMsg` and **automatically switches to the Chat page** -- even if the operator is on another page (Tasks, Settings, etc.)
3. The approval prompt renders in the chat view using the two-tier system:
   - **Tier 1 (Inline Strip)** for safe/moderate tools
   - **Tier 2 (Fullscreen Dialog)** for dangerous tools with diff preview and risk badge
4. The operator responds with keyboard keys:
   - `a` -- approve this invocation
   - `s` -- approve and allow for the rest of the session
   - `d` or `Esc` -- deny the invocation
5. For critical-risk tools, `a` and `s` require a double-press confirmation
6. The response is sent back through the approval channel to the originating channel adapter

The automatic page switch is essential -- without it, approval requests from background tasks or channel users would remain invisible and eventually time out if the operator happened to be on another page.

## Context Panel Channel Status

Toggle the context panel with `Ctrl+P` to see live channel status. The **Channels** section appears only when at least one channel is connected or has been seeded by the tracker.

Each channel shows:

```
Channels
────────────────────
  ● discord   3 msgs
  ○ slack     0 msgs
  ● telegram  12 msgs
```

- `●` (green) -- connected and operational
- `○` (red) -- disconnected or failed to start

The status indicator and message count update on each context panel tick (every 5 seconds). The `ChannelTracker` aggregates data from `ChannelMessageReceivedEvent` and `ChannelMessageSentEvent` events on the EventBus.

### How Channel Status is Seeded

When the cockpit starts:

1. All configured channels are registered with `tracker.SeedChannel(name, false)` -- initially marked disconnected
2. Each channel's `Start()` runs in a goroutine
3. On successful start, `tracker.SeedChannel(name, true)` updates the status to connected
4. If start fails, the channel remains marked disconnected and an error is printed to stderr

This means the context panel may briefly show all channels as disconnected during startup until the goroutines complete.

## Multi-Channel Operation

### Session Isolation

Each channel adapter maintains independent sessions. A Telegram user's conversation is isolated from a Discord user's conversation. The `SessionKey` field in channel messages identifies the originating session (e.g., `telegram-<chatID>`, `discord-<channelID>`).

All channel messages appear interleaved in the cockpit's chat transcript, but the agent processes each session independently.

### Message Attribution

Every channel message in the transcript includes the channel badge, so the operator can distinguish which platform a message originated from at a glance. The colored badges (blue for Telegram, purple for Discord, aubergine for Slack) provide quick visual differentiation.

## Troubleshooting

### Channel messages not appearing in cockpit

**Cause**: The `--with-channels` flag was not passed when launching the cockpit.

**Fix**: Restart with `lango cockpit --with-channels`. Without this flag, channel adapters are never started and no EventBus subscriptions are created.

### Channel shows disconnected (red circle) in context panel

**Possible causes**:

- **Expired or invalid bot token** -- verify credentials with `lango settings` or check the channel-specific configuration in [Channels](channels.md)
- **Network connectivity** -- the channel adapter could not reach the platform API
- **Startup still in progress** -- channels start asynchronously; wait a few seconds after launch

**Diagnostic**: Check `~/.lango/cockpit.log` (or `<dataRoot>/cockpit.log`) for channel start errors. Failed starts print to stderr and log the specific error.

### Approval request times out

**Cause**: The operator was on a non-chat page when the request arrived, or they did not respond in time.

**Context**: `ApprovalRequestMsg` automatically switches to the Chat page, so the operator should see the prompt immediately. However, if the terminal is backgrounded or minimized, the visual notification may be missed.

**Fix**: Keep the cockpit terminal visible when channels are active. The approval timeout is configured per channel (default 30 seconds for Telegram).

### Duplicate messages or bot conflicts

**Cause**: Running `lango cockpit --with-channels` and `lango serve` simultaneously with the same channel credentials.

**Fix**: Only one process should own a channel's bot token at a time. Either:
- Use `lango serve` for production channel handling, and `lango cockpit` (without `--with-channels`) for local monitoring
- Stop `lango serve` before launching `lango cockpit --with-channels`

### Channel section missing from context panel

**Cause**: No channels are configured or enabled in the profile configuration.

**Fix**: Run `lango onboard` and select Channel Setup, or enable channels directly in `lango settings`. At least one channel must have `enabled: true` in the configuration for the Channels section to appear.

## Related

- [Channels](channels.md) -- full channel setup, credentials, and security configuration
- [Cockpit TUI](cockpit.md) -- cockpit layout, pages, keyboard shortcuts
- [Cockpit Tasks Guide](cockpit-tasks-guide.md) -- background task management in cockpit
