## Why

Phase 1-5에서 cockpit TUI가 완성되었지만, 긴 세션과 많은 이벤트에서 성능이 저하되는 구조적 문제가 있다. render()가 O(n)으로 모든 entries를 매번 re-render하고, 매 render마다 2-4K의 lipgloss.NewStyle() 할당이 발생하며, transcript/task/grant가 무한 성장한다.

## What Changes

- Module-level style pre-allocation으로 per-render style 할당 제거 (render_tool, render_thinking, render_channel, render_delegation, render_recovery, chatview render helpers)
- Transcript entry별 cachedBlock memoization — 변경되지 않은 entries는 캐시된 block 재사용
- Transcript max entries cap (2000) + tombstone trimming — in-flight tool/thinking entries 보존, 새 backing array로 메모리 해제
- Background task FIFO cap (maxTerminalTasks=500) — CompletedAt 기준 oldest eviction
- GrantStore lazy cleanup — List() 내 Lock 승격 + cleanExpiredLocked()
- Approval diff styled line cache — content+width+splitMode 기준 캐시, scroll은 windowing만
- Context panel slice 재사용 + style pre-allocation + toolCountSum 캐시

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
