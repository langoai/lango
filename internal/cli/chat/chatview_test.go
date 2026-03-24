package chat

import (
	"strings"
	"testing"
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

	// Capture content at width 80.
	content80 := cv.viewport.View()

	// Resize to a different width.
	cv.setSize(40, 20)
	content40 := cv.viewport.View()

	// Content should be re-rendered (may differ due to word wrap).
	// At minimum, both should contain the text.
	if !strings.Contains(content80, "short sentence") {
		t.Error("content at width 80 should contain the text")
	}
	if !strings.Contains(content40, "short sentence") {
		t.Error("content at width 40 should contain the text")
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
