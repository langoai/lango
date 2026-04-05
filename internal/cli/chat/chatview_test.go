package chat

import (
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
