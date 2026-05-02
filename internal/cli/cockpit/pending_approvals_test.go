package cockpit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/cli/chat"
)

func TestPendingApprovalRegistryLatestReplaces(t *testing.T) {
	registry := NewPendingApprovalRegistry()

	firstRespCh := make(chan approval.ApprovalResponse, 1)
	first := chat.ApprovalRequestMsg{
		Request:  approval.ApprovalRequest{ID: "apr-1", ToolName: "fs_read"},
		Response: firstRespCh,
	}
	second := chat.ApprovalRequestMsg{
		Request:  approval.ApprovalRequest{ID: "apr-2", ToolName: "exec"},
		Response: make(chan approval.ApprovalResponse, 1),
	}

	registry.Register(first)
	registry.Register(second)

	require.True(t, registry.HasPending())
	latest := registry.Latest()
	require.NotNil(t, latest)
	assert.Equal(t, "apr-2", latest.Request.ID)
	assert.Equal(t, "exec", latest.Request.ToolName)

	superseded := <-firstRespCh
	assert.False(t, superseded.Approved)
	assert.False(t, superseded.AlwaysAllow)
	assert.Equal(t, "tui", superseded.Provider)
}

func TestPendingApprovalRegistryResolveWritesExactlyOnce(t *testing.T) {
	registry := NewPendingApprovalRegistry()
	respCh := make(chan approval.ApprovalResponse, 1)
	registry.Register(chat.ApprovalRequestMsg{
		Request:  approval.ApprovalRequest{ID: "apr-1", ToolName: "exec"},
		Response: respCh,
	})

	ok := registry.Resolve("apr-1", approval.ApprovalResponse{
		Approved:    true,
		AlwaysAllow: true,
		Provider:    "tui",
	})
	require.True(t, ok)

	resp := <-respCh
	assert.True(t, resp.Approved)
	assert.True(t, resp.AlwaysAllow)
	assert.Equal(t, "tui", resp.Provider)

	assert.False(t, registry.Resolve("apr-1", approval.ApprovalResponse{Approved: false}))
	assert.False(t, registry.HasPending())
	assert.Nil(t, registry.Latest())

	select {
	case extra := <-respCh:
		t.Fatalf("unexpected extra response written: %+v", extra)
	default:
	}
}

func TestPendingApprovalRegistryCockpitOwnsPendingWhenChatMounted(t *testing.T) {
	mock := &mockChild{}
	registry := NewPendingApprovalRegistry()
	m := newTestModel(mock)
	m.pendingApprovals = registry
	toolsPage := &mockPage{title: "Tools"}
	m.RegisterPage(PageTools, toolsPage)
	m.switchPage(PageTools)

	msg := chat.ApprovalRequestMsg{
		Request:  approval.ApprovalRequest{ID: "apr-1", ToolName: "exec"},
		Response: make(chan approval.ApprovalResponse, 1),
	}

	m.Update(msg)

	require.True(t, registry.HasPending())
	latest := registry.Latest()
	require.NotNil(t, latest)
	assert.Equal(t, "apr-1", latest.Request.ID)
	assert.Equal(t, PageChat, m.activePage)
	require.Len(t, mock.updates, 1)
	assert.Equal(t, msg, mock.updates[0])
}
