## 1. Render Style Pre-allocation (Wave 1A)

- [x] 1.1 Pre-allocate styles in chatview.go: renderTranscriptBlock (3 vars), renderSystemBlock (2 vars), renderStatusBlock (2 vars), renderApprovalEventBlock (reuses status styles)
- [x] 1.2 Pre-allocate styles in render_tool.go: toolLabelStyle, toolDetailStyle, toolOutputStyle
- [x] 1.3 Pre-allocate styles in render_thinking.go: 4 module-level vars
- [x] 1.4 Pre-allocate styles in render_channel.go: channelBadgeStyle, channelSenderStyle, channelTextStyle
- [x] 1.5 Pre-allocate styles in render_delegation.go: 4 module-level vars
- [x] 1.6 Pre-allocate styles in render_recovery.go: 2 module-level vars
- [x] 1.7 Verify go build/test/vet pass — 33 NewStyle() calls eliminated

## 2. Terminal Task Cleanup + Grant Lazy Expiry (Wave 1B)

- [x] 2.1 Add maxTerminalTasks = 500 constant to background/manager.go
- [x] 2.2 Implement evictTerminalTasksLocked() with CompletedAt/StartedAt sort
- [x] 2.3 Call eviction on every terminal state transition in execute()
- [x] 2.4 Extract cleanExpiredLocked() from GrantStore.CleanExpired()
- [x] 2.5 Upgrade GrantStore.List() from RLock to Lock with cleanExpiredLocked() call
- [x] 2.6 Add TestTerminalTaskEviction and TestTerminalTaskEviction_PreservesActiveTasks
- [x] 2.7 Add TestGrantLazyCleanup and TestGrantLazyCleanup_NoTTL
- [x] 2.8 Verify go build/test/vet pass

## 3. Approval Diff Styled Line Cache (Wave 1C)

- [x] 3.1 Add diffLineCache struct to approval_state.go (content, width, splitMode keys + lines slice)
- [x] 3.2 Add diffCache field to approvalState struct
- [x] 3.3 Clear diffCache in Reset() and ToggleSplit() and Clear()
- [x] 3.4 Change renderApproval signature to accept *approvalState
- [x] 3.5 Change renderApprovalDialog to read scroll/split/cache from *approvalState
- [x] 3.6 Implement cached diff line rendering with scroll windowing
- [x] 3.7 Update chat.go RenderParts and recalcLayout call sites
- [x] 3.8 Update approval_dialog_test.go and approval_origin_test.go for new signatures
- [x] 3.9 Add diff cache hit/miss/invalidation tests
- [x] 3.10 Verify go build/test/vet pass

## 4. Transcript Block Memoization + Entry Trimming (Wave 2A)

- [x] 4.1 Add cachedBlock string field to transcriptItem struct
- [x] 4.2 Extract renderEntry() helper from render() switch block
- [x] 4.3 Implement memoized render(): skip entries with valid cachedBlock
- [x] 4.4 Add cache invalidation in setSize() (width change → clear all)
- [x] 4.5 Add cache invalidation in finalizeToolResult() and finalizeThinking()
- [x] 4.6 Add maxTranscriptEntries = 2000 constant
- [x] 4.7 Create appendEntry() helper with trimming logic
- [x] 4.8 Implement tombstone-based trimming with accumulated count
- [x] 4.9 Preserve in-flight tool/thinking entries during trim
- [x] 4.10 Use make() for new backing array to enable GC of old array
- [x] 4.11 Refactor all 11 append* methods to use appendEntry()
- [x] 4.12 Add BenchmarkRender_100Entries and BenchmarkRender_1000Entries (12.5x speedup verified)
- [x] 4.13 Add TestTranscriptTrimming and TestTranscriptTrimming_AccumulatedTombstone
- [x] 4.14 Add TestRender_MemoizationCachesBlocks and invalidation tests
- [x] 4.15 Verify go build/test/vet pass

## 5. Context Panel Optimization (Wave 2B)

- [x] 5.1 Implement SetChannelStatuses() slice reuse (cap check before alloc)
- [x] 5.2 Pre-allocate module-level styles for renderRuntimeStatus and renderChannelStatus
- [x] 5.3 Add cachedToolCountSum field, compute during renderToolStats sort
- [x] 5.4 Use cached sum in refreshSnapshot() instead of recomputing
- [x] 5.5 Verify go build/test/vet pass

## 6. Final Verification

- [x] 6.1 Run go build ./... — full project build passes
- [x] 6.2 Run go test ./... — full project test suite passes (0 failures)
- [x] 6.3 Run go vet ./... — no warnings
- [x] 6.4 Benchmark: 12.5x speedup at 1000 entries (cached 408us vs uncached 5116us, 3 vs 36K allocs)
- [x] 6.5 Codex review passes — P2 tombstone off-by-one fixed, P3 diffCache Clear() fixed, P2 in-flight entry protection added, P3 new backing array for GC
