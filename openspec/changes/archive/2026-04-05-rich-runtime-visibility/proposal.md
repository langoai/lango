## Why

Cockpit v2 (channel-aware) shows start/finish status for turns, but lacks visibility into what happens *during* a turn — thinking, delegation, recovery, token consumption. Users cannot interpret long-running turns, multi-agent delegation flows, or recovery attempts without this intermediate state.

## What Changes

- Display thinking summary text preview in active thinking blocks
- Wire `OnDelegation` and `OnBudgetWarning` turnrunner callbacks to the TUI
- Create `RuntimeTracker` to aggregate runtime events from EventBus (token usage, recovery decisions) with session-key filtering
- Add `itemDelegation` and `itemRecovery` transcript items with dedicated renderers
- Show per-turn token usage summary after assistant response via DoneMsg flush
- Add "Runtime" section to context panel (active agent, delegation count, live token counter)
- Cockpit intercepts DelegationMsg, BudgetWarningMsg, RecoveryMsg, DoneMsg for cross-page routing and runtime state management

## Capabilities

### New Capabilities
- `tui-runtime-visibility`: Per-turn runtime event rendering in cockpit transcript — delegation flow, recovery attempts, budget warnings, token usage summary, and context panel runtime section

### Modified Capabilities
- `tui-thinking-indicator`: Thinking block now displays summary text preview in active state using `ansi.Truncate`
- `tui-runtime-bridge`: Extended with RuntimeTracker (token accumulation, delegation tracking, recovery forwarding) and session-key filtering
- `cockpit-context-panel`: Added "Runtime" section with active agent, delegation count, and per-turn token display

## Impact

- `internal/cli/chat/`: messages.go (4 new msg types), bridge.go (2 new callbacks), chatview.go (2 new item kinds + 3 append methods), chat.go (4 new handlers), render_delegation.go (NEW), render_recovery.go (NEW), render_thinking.go (summary preview)
- `internal/cli/cockpit/`: runtimebridge.go (NEW — RuntimeTracker), cockpit.go (5 msg intercepts + runtime state management), contextpanel.go (Runtime section + SetRuntimeStatus)
- `cmd/lango/main.go`: RuntimeTracker creation and wiring in runCockpit()
- No breaking changes. All new features degrade gracefully when structured orchestration or channels are disabled.
