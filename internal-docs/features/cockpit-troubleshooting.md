# Cockpit Troubleshooting Guide

This guide covers common issues when using the Lango cockpit TUI and how to resolve them.

---

## 1. Cockpit Won't Start

### Symptom

Running `lango` or `lango cockpit` exits immediately with an error or shows the help text instead of launching the TUI.

### Cause

- **Non-interactive terminal (non-TTY)**: The cockpit checks `prompt.IsInteractive()` before launching. If stdin is not a terminal (e.g., piped input, CI environment, or `ssh` without `-t`), the command prints help text or returns an error: `"cockpit requires an interactive terminal"`.
- **Missing or invalid config**: The bootstrap process (`cliboot.BootResult()`) fails if the configuration file is missing, malformed, or references unavailable resources (e.g., a database path that cannot be created).
- **Alt-screen not supported**: The cockpit uses `tea.WithAltScreen()`. Terminals that do not support the alternate screen buffer may render incorrectly or fail to initialize.

### Solution

1. Ensure you are running in an interactive terminal. For SSH sessions, use `ssh -t` to allocate a pseudo-TTY.
2. Run `lango doctor` to diagnose configuration issues. If the config file is missing, run `lango onboard` to create one.
3. Use a terminal emulator that supports the alternate screen buffer (iTerm2, WezTerm, kitty, or any modern xterm-compatible terminal).

---

## 2. Context Panel Is Empty

### Symptom

Pressing `Ctrl+P` opens the context panel, but it shows no data (no token usage, no tool stats, no system info).

### Cause

- **Panel not toggled on**: The context panel is hidden by default. It must be toggled on with `Ctrl+P`.
- **MetricsCollector not initialized**: If `application.MetricsCollector` is nil during bootstrap, the context panel renders placeholder text. This can happen if observability is not properly configured.
- **Tick cycle not yet fired**: The context panel refreshes on a 5-second tick cycle (`tea.Tick(5*time.Second, ...)`). After toggling the panel on, the first snapshot is taken immediately via `refreshSnapshot()` in `Start()`, but subsequent updates arrive every 5 seconds. If the collector has no data yet, the panel will appear empty until events are recorded.

### Solution

1. Press `Ctrl+P` to toggle the context panel on if it is not visible.
2. Wait at least 5 seconds after opening the panel for the first tick refresh to populate data.
3. If the panel consistently shows placeholder text, verify that your configuration enables observability features. Run `lango doctor` to check.

---

## 3. Runtime Status Not Showing

### Symptom

The context panel is open but the "Runtime" section does not appear, even though a turn is expected to be running.

### Cause

- **No active local turn**: The Runtime section only renders when `runtimeStatus.IsRunning` is `true`. This flag is set by `RuntimeTracker.StartTurn()`, which is triggered when the cockpit receives the first content event (`ToolStartedMsg`, `ThinkingStartedMsg`, or `ChunkMsg`). If no such event has been received, the section stays hidden.
- **Channel-originated turns**: The `RuntimeTracker` filters token events by session key. Events with a session key that does not match the local cockpit session (e.g., tokens from channel or background turns) are discarded. Channel-originated turns will not increment the runtime token counter or show the Runtime section.
- **Turn already completed**: When a `DoneMsg` is received, the cockpit calls `runtimeTracker.ResetTurn()`, which sets `turnActive = false` and clears the delegation count and active agent. The Runtime section disappears immediately after the turn ends.

### Solution

1. The Runtime section is designed to appear only during active local turns. This is expected behavior, not a bug.
2. To see runtime metrics for a turn, watch the context panel while actively chatting. The section appears when the agent starts processing and disappears when the response is complete.
3. Token usage summaries for completed turns are appended to the chat as `TurnTokenUsageMsg` after each turn ends.

---

## 4. Approval Times Out

### Symptom

A tool approval prompt appears but the timer expires before the user can respond, or the approval prompt is never seen at all.

### Cause

- **User was on a non-Chat page**: When an `ApprovalRequestMsg` arrives, the cockpit automatically switches to the Chat page (`handleApprovalRequest` calls `switchPage(PageChat)`). However, if the page switch fails or the user navigates away before responding, the approval may time out.
- **Default timeout**: Channel-based approval providers default to 30 seconds (`ApprovalTimeoutSec` in channel config, default 30). The TUI approval provider itself does not enforce a timeout -- it waits until the user responds or the context is cancelled. The timeout is controlled by the upstream caller (gateway or toolchain middleware).
- **Approval response keys not recognized**: Approval responses require pressing `a` (allow), `s` (allow for session), or `d`/`Esc` (deny). These keys only work when the approval prompt is actively displayed in the Chat page.

### Solution

1. If you see the approval prompt but on the wrong page, press `Ctrl+1` to switch to the Chat page manually.
2. Respond promptly using `a` (allow), `s` (allow for session), or `d` (deny) once the prompt is visible.
3. To adjust the timeout for channel-originated approvals, set `approvalTimeoutSec` in the channel configuration (e.g., `channels.telegram.approvalTimeoutSec`).
4. The approval middleware retries up to `MaxTurnApprovalTimeouts` (3) timeouts before failing the turn permanently.

---

## 5. Channel Messages Not Appearing

### Symptom

The cockpit is running but messages from Telegram, Discord, or Slack channels are not showing up in the Chat page.

### Cause

- **`--with-channels` flag not passed**: By default, `lango cockpit` runs in local-chat mode (`app.WithLocalChat()`). Channel adapters are only started when `--with-channels` is explicitly provided. Without it, no channel listeners are active.
- **Channel credentials not configured**: Even with `--with-channels`, each channel requires valid credentials (API tokens, bot tokens, etc.) in the configuration file. If credentials are missing or invalid, the channel fails to start and an error is printed to stderr.
- **EventBus subscription timing**: The cockpit subscribes to channel events via `cockpit.SubscribeChannelEvents(application.EventBus, p)` before starting channel polling loops. If this wiring is disrupted, events may be dropped silently.
- **Conflict with `lango serve`**: Running `lango cockpit --with-channels` while `lango serve` is already running with the same credentials can cause conflicts (e.g., Telegram bot webhook collisions).

### Solution

1. Start the cockpit with channel support: `lango cockpit --with-channels`.
2. Verify channel credentials are configured correctly. Run `lango doctor` to check.
3. Do not run `lango cockpit --with-channels` and `lango serve` simultaneously with the same channel credentials.
4. Check `cockpit.log` (located in your data root, default `~/.lango/cockpit.log`) for channel start errors.

---

## 6. Task Fails or Can't Be Retried

### Symptom

A background task shows an error, or the retry/cancel action does not respond when pressing `r` or `c` on the Tasks page.

### Cause

- **Retry is status-gated**: The `retrySelectedTask()` method only works when the selected task has status `"failed"` or `"cancelled"`. Pressing `r` on a `"running"` or `"pending"` task does nothing.
- **Cancel is status-gated**: The `cancelSelectedTask()` method only works when the selected task has status `"running"` or `"pending"`. Pressing `c` on a completed or failed task does nothing.
- **No BackgroundManager**: If `application.BackgroundManager` is nil, the Tasks page renders "No background tasks available" and no actions are possible.
- **Error details hidden**: The full error message is only visible in the detail panel. Press `Enter` to expand the detail view for the selected task.

### Solution

1. Press `Enter` on the selected task to open the detail panel and inspect the error message, result, origin channel, and token usage.
2. To retry a task, select a task with status `"failed"` or `"cancelled"` and press `r`.
3. To cancel a task, select a task with status `"running"` or `"pending"` and press `c`.
4. Use `j`/`k` or arrow keys to navigate the task list; use `Esc` to close the detail panel.

---

## 7. Keyboard Shortcuts Not Working

### Symptom

Pressing keyboard shortcuts has no effect, or keys seem to be routed to the wrong component.

### Cause

- **Focus state**: The cockpit has a focus toggle between the sidebar and the content area, controlled by `Tab`. When the sidebar is focused, key presses are routed to the sidebar (for navigation) rather than the active page. However, global shortcuts (`Ctrl+1` through `Ctrl+6`, `Ctrl+B`, `Ctrl+P`, `Ctrl+Y`) always work regardless of focus state.
- **Page-specific keys**: Some keys only work on specific pages or in specific states:
  - `a`, `s`, `d` -- only when an approval prompt is active on the Chat page
  - `c` (cancel), `r` (retry) -- only on the Tasks page when a task is selected
  - `Enter` -- toggles detail mode on the Tasks page
  - `Esc` -- closes detail panel on the Tasks page
- **Terminal key capture conflicts**: Some terminal emulators intercept `Ctrl+` combinations before they reach the application (e.g., `Ctrl+C` for SIGINT, `Ctrl+Z` for suspend).

### Solution

1. Press `Tab` to toggle focus between the sidebar and content area. The active area receives non-global key events.
2. Use the global shortcuts to switch pages regardless of focus: `Ctrl+1` (Chat), `Ctrl+2` (Settings), `Ctrl+3` (Tools), `Ctrl+4` (Status), `Ctrl+5` (Tasks), `Ctrl+6` (Approvals).
3. Other global shortcuts: `Ctrl+B` (toggle sidebar), `Ctrl+P` (toggle context panel), `Ctrl+Y` (copy to clipboard).
4. If your terminal captures certain key combinations, check your terminal emulator's settings or try a different emulator.

---

## 8. Terminal Rendering Issues (ANSI/CPR)

### Symptom

Stray characters appear in the input composer, the screen flickers, or the layout is corrupted. Characters like `[?1;2R` or other escape sequences leak into visible text.

### Cause

- **CPR (Cursor Position Report) sequences**: Some terminals send CPR responses (`ESC [ <row> ; <col> R`) or OSC (Operating System Command) sequences in response to queries. The cockpit includes a `cprFilter` state machine that intercepts and discards these sequences. It buffers ambiguous key messages (starting with `ESC`) and waits up to 50ms (`cprTimeout`) to determine if the sequence is a terminal response or a normal key press.
- **Unrecognized sequences**: If the terminal sends a response format that the `cprFilter` does not recognize, the buffered keys are flushed back to the input handler. This can cause partial escape sequences to appear as visible characters.
- **OSC sequences**: The filter also handles OSC sequences (terminated by BEL or `ESC \`), but non-standard OSC formats may not be caught.

### Solution

1. Use a terminal emulator known to work well with TUI applications: iTerm2, WezTerm, kitty, Ghostty, or Alacritty.
2. If characters leak into the composer, try pressing `Escape` to reset the filter state, then continue typing.
3. Avoid terminal multiplexers (tmux, screen) if rendering issues persist, as they may add their own escape sequence handling layer.
4. If the issue is consistent, check `cockpit.log` for any logged errors during rendering.

---

## 9. Checking Logs

### Symptom

Something is wrong but no error is visible in the TUI. You need to find runtime error details.

### Cause

The cockpit redirects all logging output (including Go's standard `log` package) to a log file to prevent stderr output from corrupting the alt-screen TUI. The log file path is determined by the `DataRoot` configuration value.

### Solution

1. **Cockpit log file**: Check `<DataRoot>/cockpit.log` (default: `~/.lango/cockpit.log`). This file receives all structured logs and stdlib log output during cockpit operation.

   ```bash
   tail -f ~/.lango/cockpit.log
   ```

2. **Chat log file**: If using `lango chat` instead of cockpit, the log file is `<DataRoot>/chat.log` (default: `~/.lango/chat.log`).

3. **Run diagnostics**: Use `lango doctor` to perform automated checks on your configuration, credentials, database, and environment.

   ```bash
   lango doctor
   ```

4. **Log level**: The log level and format are controlled by your configuration (`logging.level`, `logging.format`). Set the level to `debug` for more verbose output when troubleshooting:

   ```bash
   lango config set logging.level debug
   ```

5. **Channel start errors**: When using `--with-channels`, channel startup errors are printed to stderr before the TUI takes over the screen. These may scroll off quickly. Check `cockpit.log` for the persisted error messages.
