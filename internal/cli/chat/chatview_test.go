package chat

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendAssistant_PreservesRawContent(t *testing.T) {
	cv := newChatViewModel(80, 20)
	raw := "Hello **world**"
	cv.appendAssistant(raw)

	if len(cv.entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(cv.entries))
	}
	entry := cv.entries[0]
	if entry.kind != itemAssistant {
		t.Errorf("want assistant kind, got %q", entry.kind)
	}
	if entry.rawContent != raw {
		t.Errorf("want rawContent=%q, got %q", raw, entry.rawContent)
	}
	if entry.content == "" {
		t.Error("want non-empty rendered content")
	}
}

func TestAppendAssistant_SkipsEmpty(t *testing.T) {
	cv := newChatViewModel(80, 20)
	cv.appendAssistant("")
	cv.appendAssistant("   ")
	if len(cv.entries) != 0 {
		t.Errorf("want 0 entries for blank input, got %d", len(cv.entries))
	}
}

func TestFinalizeStream_CommitsBufferAsAssistant(t *testing.T) {
	cv := newChatViewModel(80, 20)
	cv.streamBuf.WriteString("streaming content")
	cv.finalizeStream()

	if cv.streamBuf.Len() != 0 {
		t.Error("want streamBuf cleared after finalize")
	}
	if len(cv.entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(cv.entries))
	}
	if cv.entries[0].rawContent != "streaming content" {
		t.Errorf("want rawContent=%q, got %q", "streaming content", cv.entries[0].rawContent)
	}
}

func TestFinalizeStream_EmptyBufferNoEntry(t *testing.T) {
	cv := newChatViewModel(80, 20)
	cv.finalizeStream()
	if len(cv.entries) != 0 {
		t.Errorf("want 0 entries for empty stream, got %d", len(cv.entries))
	}
}

func TestRender_BlockJoinNoLeadingBlanks(t *testing.T) {
	cv := newChatViewModel(80, 20)
	cv.appendUser("hello")
	cv.appendAssistant("world")

	content := cv.viewport.View()
	// Block-join uses \n\n between blocks, not leading \n on each.
	// The rendered output should NOT start with blank lines.
	if strings.HasPrefix(content, "\n\n") {
		t.Error("rendered content should not start with double blank lines")
	}
}

func TestRender_ResizeReflowsAssistantMarkdown(t *testing.T) {
	cv := newChatViewModel(80, 20)
	cv.appendAssistant("A short sentence.")

	if len(cv.entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(cv.entries))
	}
	if cv.entries[0].rawContent != "A short sentence." {
		t.Fatalf("rawContent should be preserved, got %q", cv.entries[0].rawContent)
	}

	content80 := cv.viewport.View()

	cv.setSize(40, 20)
	content40 := cv.viewport.View()

	if strings.TrimSpace(content80) == "" {
		t.Fatal("content at width 80 should not be empty")
	}
	if strings.TrimSpace(content40) == "" {
		t.Fatal("content at width 40 should not be empty")
	}
}

func TestContentWidth_MinimumClamp(t *testing.T) {
	cv := newChatViewModel(5, 20) // very narrow
	w := cv.contentWidth()
	if w < 10 {
		t.Errorf("contentWidth should be at least 10, got %d", w)
	}
}

func TestContentWidth_Normal(t *testing.T) {
	cv := newChatViewModel(80, 20)
	w := cv.contentWidth()
	if w != 78 {
		t.Errorf("contentWidth for width=80 should be 78, got %d", w)
	}
}

func TestRender_InFlightStreamingBlock(t *testing.T) {
	cv := newChatViewModel(80, 20)
	cv.appendUser("question")
	cv.appendChunk("partial ")
	cv.appendChunk("response")

	content := cv.viewport.View()
	if !strings.Contains(content, "partial response") {
		t.Error("in-flight streaming content should be visible")
	}
}

func TestAppendStatusAndApprovalEventKinds(t *testing.T) {
	cv := newChatViewModel(80, 20)
	cv.appendStatus("working", "info")
	cv.appendApprovalEvent("approval requested", "requested")

	if len(cv.entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(cv.entries))
	}
	if cv.entries[0].kind != itemStatus {
		t.Fatalf("first item should be status, got %q", cv.entries[0].kind)
	}
	if cv.entries[1].kind != itemApproval {
		t.Fatalf("second item should be approval, got %q", cv.entries[1].kind)
	}
}

// ---------------------------------------------------------------------------
// Tool lifecycle tests
// ---------------------------------------------------------------------------

func TestAppendToolStart(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendToolStart("call1", "fs_read", nil)

	require.Len(t, cv.entries, 1)
	e := cv.entries[0]
	assert.Equal(t, itemTool, e.kind)
	assert.Equal(t, "fs_read", e.content)
	assert.Equal(t, "call1", e.meta["callID"])
	assert.Equal(t, string(toolStateRunning), e.meta["state"])
}

func TestAppendToolStart_Multiple(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendToolStart("call1", "fs_read", nil)
	cv.appendToolStart("call2", "web_search", nil)

	require.Len(t, cv.entries, 2)
	assert.Equal(t, "fs_read", cv.entries[0].content)
	assert.Equal(t, "call1", cv.entries[0].meta["callID"])
	assert.Equal(t, "web_search", cv.entries[1].content)
	assert.Equal(t, "call2", cv.entries[1].meta["callID"])
}

func TestFinalizeToolResult_Success(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendToolStart("call1", "fs_read", nil)
	cv.finalizeToolResult("call1", true, 500*time.Millisecond, "")

	require.Len(t, cv.entries, 1)
	e := cv.entries[0]
	assert.Equal(t, string(toolStateSuccess), e.meta["state"])
	assert.NotEmpty(t, e.meta["duration"])
}

func TestFinalizeToolResult_Error(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendToolStart("call1", "fs_read", nil)
	cv.finalizeToolResult("call1", false, 1*time.Second, "")

	require.Len(t, cv.entries, 1)
	e := cv.entries[0]
	assert.Equal(t, string(toolStateError), e.meta["state"])
}

func TestFinalizeToolResult_WithOutput(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendToolStart("call1", "fs_read", nil)
	cv.finalizeToolResult("call1", true, 200*time.Millisecond, "file contents here")

	require.Len(t, cv.entries, 1)
	e := cv.entries[0]
	assert.Equal(t, "file contents here", e.meta["output"])
	assert.Equal(t, string(toolStateSuccess), e.meta["state"])
}

func TestFinalizeToolResult_NoMatch(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendToolStart("call1", "fs_read", nil)
	cv.finalizeToolResult("nonexistent", true, 100*time.Millisecond, "output")

	require.Len(t, cv.entries, 1)
	e := cv.entries[0]
	// Original entry should remain unchanged — still running, no output.
	assert.Equal(t, string(toolStateRunning), e.meta["state"])
	assert.Empty(t, e.meta["output"])
	assert.Empty(t, e.meta["duration"])
}

// ---------------------------------------------------------------------------
// Thinking lifecycle tests
// ---------------------------------------------------------------------------

func TestAppendThinking(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendThinking("analyzing request")

	require.Len(t, cv.entries, 1)
	e := cv.entries[0]
	assert.Equal(t, itemThinking, e.kind)
	assert.Equal(t, "analyzing request", e.content)
	assert.Equal(t, "active", e.meta["state"])
}

func TestFinalizeThinking_Done(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendThinking("analyzing request")
	cv.finalizeThinking("done analyzing", 2*time.Second)

	require.Len(t, cv.entries, 1)
	e := cv.entries[0]
	assert.Equal(t, "done", e.meta["state"])
	assert.NotEmpty(t, e.meta["duration"])
	assert.Equal(t, "done analyzing", e.content)
}

func TestFinalizeThinking_Summary(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendThinking("initial summary")
	cv.finalizeThinking("replaced summary", 3*time.Second)

	require.Len(t, cv.entries, 1)
	assert.Equal(t, "replaced summary", cv.entries[0].content)
}

func TestFinalizeThinking_NoActive(t *testing.T) {
	cv := newChatViewModel(80, 24)
	// No prior appendThinking — should not panic or add entries.
	cv.finalizeThinking("orphan summary", 1*time.Second)
	assert.Empty(t, cv.entries)
}

// ---------------------------------------------------------------------------
// Other transcript tests
// ---------------------------------------------------------------------------

func TestClear(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendUser("hello")
	cv.appendSystem("system note")
	cv.appendToolStart("c1", "tool", nil)
	require.Len(t, cv.entries, 3)

	cv.clear()
	assert.Empty(t, cv.entries)
	assert.Equal(t, 0, cv.streamBuf.Len())
}

func TestAppendSystem(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendSystem("system message")

	require.Len(t, cv.entries, 1)
	e := cv.entries[0]
	assert.Equal(t, itemSystem, e.kind)
	assert.Equal(t, "system message", e.content)
}

func TestAppendSystem_Empty(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendSystem("")
	cv.appendSystem("   ")
	assert.Empty(t, cv.entries)
}

// ---------------------------------------------------------------------------
// Delegation / Recovery / Token Summary tests
// ---------------------------------------------------------------------------

func TestAppendDelegation(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendDelegation("operator", "librarian", "search needed")

	require.Len(t, cv.entries, 1)
	e := cv.entries[0]
	assert.Equal(t, itemDelegation, e.kind)
	assert.Equal(t, "operator", e.meta["from"])
	assert.Equal(t, "librarian", e.meta["to"])
	assert.Equal(t, "search needed", e.meta["reason"])
}

func TestAppendDelegation_EmptyReason(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendDelegation("operator", "librarian", "")

	require.Len(t, cv.entries, 1)
	assert.Equal(t, "", cv.entries[0].meta["reason"])
}

func TestAppendRecovery(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendRecovery("retry", "rate_limit", 2, 3*time.Second)

	require.Len(t, cv.entries, 1)
	e := cv.entries[0]
	assert.Equal(t, itemRecovery, e.kind)
	assert.Equal(t, "retry", e.meta["action"])
	assert.Equal(t, "rate_limit", e.meta["causeClass"])
	assert.Equal(t, "2", e.meta["attempt"])
	assert.Equal(t, "3s", e.meta["backoff"])
}

func TestAppendTokenSummary(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendTokenSummary(100, 200, 300, 50)

	require.Len(t, cv.entries, 1)
	e := cv.entries[0]
	assert.Equal(t, itemStatus, e.kind) // reuses appendStatus
	assert.Contains(t, e.content, "100")
	assert.Contains(t, e.content, "200")
	assert.Contains(t, e.content, "300")
	assert.Contains(t, e.content, "50 cached")
}

func TestAppendTokenSummary_NoCache(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendTokenSummary(1500, 2000, 3500, 0)

	require.Len(t, cv.entries, 1)
	assert.Contains(t, cv.entries[0].content, "1.5k")
	assert.NotContains(t, cv.entries[0].content, "cached")
}

func TestFormatTokenCount(t *testing.T) {
	tests := []struct {
		give int64
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1.0k"},
		{1500, "1.5k"},
		{12345, "12.3k"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, formatTokenCount(tt.give))
	}
}

func TestRender_ToolAndThinking(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendToolStart("c1", "fs_read", nil)
	cv.finalizeToolResult("c1", true, 100*time.Millisecond, "output")
	cv.appendThinking("thinking hard")
	cv.finalizeThinking("thought complete", 1*time.Second)

	// The key assertion: render produces a non-empty viewport without panicking.
	content := cv.viewport.View()
	assert.NotEmpty(t, content)
}

// ---------------------------------------------------------------------------
// Block memoization tests
// ---------------------------------------------------------------------------

func TestRender_MemoizationCachesBlocks(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendUser("hello")
	cv.appendSystem("note")

	// After render, cachedBlock should be populated.
	for i, e := range cv.entries {
		assert.NotEmpty(t, e.cachedBlock, "entry %d should have cachedBlock", i)
	}
}

func TestRender_MemoizationInvalidatedOnResize(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendUser("hello")
	cv.appendAssistant("world")

	// Verify cache is populated.
	for _, e := range cv.entries {
		assert.NotEmpty(t, e.cachedBlock)
	}

	// Resize clears all caches.
	cv.setSize(60, 24)
	// After setSize triggers render(), caches are re-populated at new width.
	for i, e := range cv.entries {
		assert.NotEmpty(t, e.cachedBlock, "entry %d should be re-cached after resize", i)
	}
}

func TestRender_FinalizeToolInvalidatesCache(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendToolStart("c1", "fs_read", nil)

	cached1 := cv.entries[0].cachedBlock
	assert.NotEmpty(t, cached1)

	cv.finalizeToolResult("c1", true, 500*time.Millisecond, "output")

	// Cache should be re-populated with updated content.
	cached2 := cv.entries[0].cachedBlock
	assert.NotEmpty(t, cached2)
	assert.NotEqual(t, cached1, cached2, "cache should change after finalization")
}

func TestRender_FinalizeThinkingInvalidatesCache(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendThinking("analyzing")

	cached1 := cv.entries[0].cachedBlock
	assert.NotEmpty(t, cached1)

	cv.finalizeThinking("done analyzing", 2*time.Second)

	cached2 := cv.entries[0].cachedBlock
	assert.NotEmpty(t, cached2)
	assert.NotEqual(t, cached1, cached2, "cache should change after finalization")
}

// ---------------------------------------------------------------------------
// Transcript trimming tests
// ---------------------------------------------------------------------------

func TestTranscriptTrimming(t *testing.T) {
	cv := newChatViewModel(80, 24)

	// Add 2500 entries to trigger trimming (cap is 2000).
	for i := 0; i < 2500; i++ {
		cv.appendUser(fmt.Sprintf("message %d", i))
	}

	// After trimming 500, should have ~2001 entries (2500 - 500 + tombstone replaces [0]).
	// Actually: append triggers trim when len > 2000. At 2001, trim 500 -> 1501 entries.
	// Then further appends grow back. Let's just check it's capped.
	assert.LessOrEqual(t, len(cv.entries), maxTranscriptEntries+1,
		"entries should not exceed max + 1 (trim fires after append)")

	// First entry should be the tombstone.
	assert.Equal(t, itemSystem, cv.entries[0].kind)
	assert.Contains(t, cv.entries[0].content, "older messages trimmed")
}

func TestTranscriptTrimming_AccumulatedTombstone(t *testing.T) {
	cv := newChatViewModel(80, 24)

	// Fill to trigger first trim.
	for i := 0; i < maxTranscriptEntries+1; i++ {
		cv.appendUser(fmt.Sprintf("msg %d", i))
	}

	// Verify first tombstone.
	require.Equal(t, itemSystem, cv.entries[0].kind)
	assert.Contains(t, cv.entries[0].content, "500 older messages trimmed")

	// Fill to trigger second trim.
	remaining := maxTranscriptEntries - len(cv.entries) + 2
	for i := 0; i < remaining; i++ {
		cv.appendUser(fmt.Sprintf("more %d", i))
	}

	// Second tombstone should accumulate both trim counts.
	require.Equal(t, itemSystem, cv.entries[0].kind)
	assert.Contains(t, cv.entries[0].content, "1000 older messages trimmed",
		"tombstone should accumulate across repeated trims")
}

func TestTranscriptTrimming_PreservesRecentEntries(t *testing.T) {
	cv := newChatViewModel(80, 24)

	// Add exactly enough to trigger one trim.
	for i := 0; i < maxTranscriptEntries+1; i++ {
		cv.appendUser(fmt.Sprintf("msg-%04d", i))
	}

	// The last entry should be the most recent message.
	last := cv.entries[len(cv.entries)-1]
	assert.Equal(t, itemUser, last.kind)
	assert.Equal(t, fmt.Sprintf("msg-%04d", maxTranscriptEntries), last.content)
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkRender_100Entries(b *testing.B) {
	cv := newChatViewModel(80, 40)
	for i := 0; i < 100; i++ {
		cv.appendUser(fmt.Sprintf("message %d", i))
	}
	// Clear caches to benchmark cold render.
	for i := range cv.entries {
		cv.entries[i].cachedBlock = ""
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear caches each iteration to measure full render cost.
		for j := range cv.entries {
			cv.entries[j].cachedBlock = ""
		}
		cv.render()
	}
}

func BenchmarkRender_1000Entries(b *testing.B) {
	cv := newChatViewModel(80, 40)
	for i := 0; i < 1000; i++ {
		cv.appendUser(fmt.Sprintf("message %d", i))
	}
	b.ResetTimer()

	// First render: cold (no cache).
	for i := range cv.entries {
		cv.entries[i].cachedBlock = ""
	}
	cv.render()

	// Subsequent renders: hot (all cached). This should be significantly faster.
	b.Run("cached", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cv.render()
		}
	})

	b.Run("uncached", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := range cv.entries {
				cv.entries[j].cachedBlock = ""
			}
			cv.render()
		}
	})
}
