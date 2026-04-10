## Why

In Phase 1-5, the cockpit TUI was completed, but there are structural issues causing performance degradation in long sessions with many events. render() runs O(n) re-rendering all entries every time, 2-4K lipgloss.NewStyle() allocations occur per render, and transcript/task/grant grow without bounds.

## What Changes

- Module-level style pre-allocation to eliminate per-render style allocation (render_tool, render_thinking, render_channel, render_delegation, render_recovery, chatview render helpers)
- Per-transcript-entry cachedBlock memoization — unchanged entries reuse cached blocks
- Transcript max entries cap (2000) + tombstone trimming — in-flight tool/thinking entries preserved, memory released via new backing array
- Background task FIFO cap (maxTerminalTasks=500) — oldest eviction based on CompletedAt
- GrantStore lazy cleanup — Lock upgrade + cleanExpiredLocked() in List()
- Approval diff styled line cache — cached by content+width+splitMode, scroll uses windowing only
- Context panel slice reuse + style pre-allocation + toolCountSum cache

## Capabilities

### New Capabilities
- `tui-perf-transcript-memo`: Transcript block memoization and entry trimming with in-flight entry preservation
- `tui-perf-style-prealloc`: Module-level style pre-allocation pattern across all render files

### Modified Capabilities
- `tui-approval-tiers`: Approval dialog diff rendering cached with diffLineCache; renderApproval signature changed to accept *approvalState
- `cockpit-context-panel`: SetChannelStatuses slice reuse, style pre-allocation, toolCountSum caching

## Impact

- `internal/cli/chat/chatview.go`: cachedBlock field, appendEntry helper, memoized render(), entry trimming
- `internal/cli/chat/render_*.go` (5 files): module-level style vars
- `internal/cli/chat/approval_state.go`: diffLineCache struct and field, Clear() cache cleanup
- `internal/cli/chat/approval_dialog.go`: cached diff line rendering, signature change
- `internal/cli/chat/approval.go`: renderApproval signature change
- `internal/cli/chat/chat.go`: renderApproval call site updates
- `internal/background/manager.go`: maxTerminalTasks cap, FIFO eviction
- `internal/approval/grant.go`: cleanExpiredLocked, List() Lock upgrade
- `internal/cli/cockpit/contextpanel.go`: slice reuse, style pre-alloc, sum caching
- Benchmark: 12.5x render speedup at 1000 entries (cached vs uncached)
