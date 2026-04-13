## 1. Foundation — EventBus Events + App Mode

- [x] 1.1 Add `ChannelMessageReceivedEvent` and `ChannelMessageSentEvent` to `internal/eventbus/events.go`
- [x] 1.2 Add `AppModeCockpit` mode + `WithCockpit()` option to `internal/app/types.go`
- [x] 1.3 Modify `internal/app/app.go` mode checks: cockpit caps at PriorityBuffer, calls initChannels, skips registerPostBuildLifecycle
- [x] 1.4 Add `Name() string` to `Channel` interface in `internal/app/types.go`
- [x] 1.5 Implement `Name()` on Telegram, Discord, Slack adapters
- [x] 1.6 Update `noopChannel` mock in `parity_test.go`
- [x] 1.7 Add EventBus event tests in `internal/eventbus/events_test.go`

## 2. Channel Handler Publishing

- [x] 2.1 Publish `ChannelMessageReceivedEvent` in `handleTelegramMessage` (with chatID metadata)
- [x] 2.2 Publish `ChannelMessageSentEvent` in `handleTelegramMessage` after runAgent success
- [x] 2.3 Publish events in `handleDiscordMessage` (with channelID, guildID metadata)
- [x] 2.4 Publish events in `handleSlackMessage` (with channelID, threadTS metadata)
- [x] 2.5 Guard all EventBus publishing with `if a.EventBus != nil`
- [x] 2.6 Add `internal/app/channels_test.go` with event struct and publish/subscribe tests

## 3. TUI Transcript Integration

- [x] 3.1 Add `ChannelMessageMsg` tea.Msg to `internal/cli/chat/messages.go`
- [x] 3.2 Add `itemChannel` kind + `appendChannel()` to `internal/cli/chat/chatview.go`
- [x] 3.3 Preserve sessionKey and metadata in transcript item meta
- [x] 3.4 Add `ChannelMessageMsg` handler in `internal/cli/chat/chat.go` Update()
- [x] 3.5 Create `internal/cli/chat/render_channel.go` with per-channel colored badge renderer
- [x] 3.6 Sanitize external text with `ansi.Strip()` before rendering
- [x] 3.7 Collapse multiline text with `strings.ReplaceAll(text, "\n", " ")`
- [x] 3.8 Add `internal/cli/chat/render_channel_test.go`

## 4. EventBus-to-TUI Bridge

- [x] 4.1 Create `internal/cli/cockpit/channelbridge.go` with `SubscribeChannelEvents(bus, sender)`
- [x] 4.2 Define local `msgSender` interface in cockpit package
- [x] 4.3 Create `ChannelTracker` with EventBus subscription + `SeedChannel(name, connected)`
- [x] 4.4 Implement `Snapshot()` returning sorted channel statuses
- [x] 4.5 Add `internal/cli/cockpit/channelbridge_test.go`

## 5. Cockpit Wiring

- [x] 5.1 Add `EventBus *eventbus.Bus` to cockpit `Deps`
- [x] 5.2 Add `--with-channels` flag to `cockpitCmd()` (default false)
- [x] 5.3 Switch `runCockpit()` to `WithCockpit()` only when `--with-channels` is set
- [x] 5.4 Start channel loops manually in goroutines after `SubscribeChannelEvents`
- [x] 5.5 Cancel ctx before ch.Stop() in shutdown defer (prevent deadlock)
- [x] 5.6 Wire `SubscribeChannelEvents(bus, p)` before channel start
- [x] 5.7 Create tracker, seed channels, wire to cockpit model

## 6. Cockpit Message Routing

- [x] 6.1 Intercept `ChannelMessageMsg` in `cockpit.go` Update() and always forward to chat child
- [x] 6.2 Add `SetChannelTracker(tracker)` to cockpit Model
- [x] 6.3 Push tracker.Snapshot() to context panel on contextTickMsg

## 7. Approval Channel Origin

- [x] 7.1 Add `formatChannelOrigin(sessionKey)` helper in `approval.go`
- [x] 7.2 Add `formatChannelBadge(sessionKey)` helper in `approval.go`
- [x] 7.3 Show origin line in `renderApprovalBanner`
- [x] 7.4 Prepend channel badge in `renderApprovalStrip`
- [x] 7.5 Show origin line in `renderApprovalDialog`
- [x] 7.6 Add `internal/cli/chat/approval_origin_test.go`

## 8. Context Panel Channel Section

- [x] 8.1 Add `channelStatus` struct and `SetChannelStatuses()` to `contextpanel.go`
- [x] 8.2 Add `renderChannelStatus()` with connected/disconnected indicators
- [x] 8.3 Insert channel section into `View()` with graceful degradation
- [x] 8.4 Extend `contextpanel_test.go` with channel status tests

## 9. Verification

- [x] 9.1 `go build ./...` passes
- [x] 9.2 `go test ./...` passes
- [x] 9.3 `go vet ./...` passes
