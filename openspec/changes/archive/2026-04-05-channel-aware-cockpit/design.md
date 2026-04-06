## Context

Cockpit v1 runs in `AppModeLocalChat` which skips channel initialization entirely. The gateway server (`lango serve`) handles channels but has no TUI. This change bridges the gap by allowing the cockpit to optionally start channels and observe their activity via EventBus.

## Goals / Non-Goals

**Goals:**
- Make channel messages visible in cockpit transcript
- Show channel origin on approval requests
- Display channel connection status in context panel
- Default to safe mode (no channels) to prevent duplicate bot sessions

**Non-Goals:**
- Gateway WebSocket client (cross-process channel observation) — deferred
- Channel message sending from cockpit (cockpit is observe-only for channels)
- Channel configuration UI (use config file or `lango config`)

## Decisions

### D1. AppModeCockpit with opt-in channel start

New `AppModeCockpit` mode sits between `AppModeLocalChat` (no channels) and `AppModeServer` (full gateway). It calls `initChannels()` to create adapter objects and wire handlers, but skips `registerPostBuildLifecycle()` (no HTTP gateway). Channel `Start/Stop` is managed manually in `runCockpit()` outside the lifecycle registry because `SetMaxPriority(PriorityBuffer)` would skip `PriorityNetwork` components.

Default is `--with-channels=false` to prevent duplicate bot sessions when `lango serve` is already running with the same credentials.

### D2. EventBus for same-process channel event delivery

Two new event types (`ChannelMessageReceivedEvent`, `ChannelMessageSentEvent`) published from `app/channels.go` handlers. The cockpit subscribes via `SubscribeChannelEvents(bus, program)` which converts events to `tea.Msg`. This avoids network hops and keeps the bridge simple.

**Alternative considered:** Gateway WebSocket client. Rejected as too complex for Phase 2 — requires reconnection logic, URL configuration, authentication.

### D3. Channel.Name() interface extension

Added `Name() string` to the `Channel` interface so cockpit can identify channels without config-index alignment. Each adapter returns its platform name ("telegram", "discord", "slack"). This is a minimal breaking change to the interface (only `parity_test.go` mock needed updating).

### D4. ANSI sanitization of external channel text

Channel messages from external users are rendered in the cockpit terminal. `ansi.Strip()` removes ANSI/OSC escape sequences before rendering to prevent terminal control injection. Applied before newline collapse and width truncation.

### D5. ChannelTracker with SeedChannel for connection status

`ChannelTracker` subscribes to EventBus for message counting and exposes `SeedChannel(name, connected)` for cockpit to report start success/failure. This distinguishes "connected, no messages yet" from "start failed". Snapshot is pushed to context panel on each 5-second tick.

### D6. Channel message routing bypasses active page

`ChannelMessageMsg` is intercepted in `cockpit.go` Update() before `forwardToActive()` and always forwarded to the chat child. This prevents message loss when the operator is on a non-chat page (Settings, Status, etc.).

## Risks / Trade-offs

**[R1] Duplicate bot sessions** — If operator runs `lango --with-channels` while `lango serve` is active, both processes compete for platform updates. → Mitigation: `--with-channels` defaults to false; documentation warns against concurrent use.

**[R2] ChannelMessageSentEvent timing** — Published before adapter actually delivers the message (handler returns response, adapter sends after). → Mitigation: Documented limitation; event reflects "response ready" not "delivery confirmed".

**[R3] Channel Start blocks on network** — Channel adapters connect to external APIs during Start(). If network is slow, cockpit boot is not blocked (Start runs in goroutines). → Mitigation: SeedChannel updates connection status asynchronously.
