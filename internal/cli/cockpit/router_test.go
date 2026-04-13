package cockpit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPageIDString(t *testing.T) {
	tests := []struct {
		give PageID
		want string
	}{
		{give: PageChat, want: "chat"},
		{give: PageSettings, want: "settings"},
		{give: PageTools, want: "tools"},
		{give: PageStatus, want: "status"},
		{give: PageID(99), want: "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.give.String(); got != tt.want {
				t.Errorf("PageID(%d).String() = %q, want %q",
					tt.give, got, tt.want)
			}
		})
	}
}

func TestPageIDFromString(t *testing.T) {
	tests := []struct {
		give string
		want PageID
	}{
		{give: "chat", want: PageChat},
		{give: "settings", want: PageSettings},
		{give: "tools", want: PageTools},
		{give: "status", want: PageStatus},
		{give: "unknown-value", want: PageChat},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			if got := PageIDFromString(tt.give); got != tt.want {
				t.Errorf("PageIDFromString(%q) = %d, want %d",
					tt.give, got, tt.want)
			}
		})
	}
}

func TestAllPageMetas_Count(t *testing.T) {
	metas := AllPageMetas()
	assert.Len(t, metas, 7, "AllPageMetas should return exactly 7 items (Chat + 6 pages)")
}

func TestAllPageMetas_AllPageIDsCovered(t *testing.T) {
	metas := AllPageMetas()
	metaIDs := make(map[string]bool, len(metas))
	for _, m := range metas {
		metaIDs[m.ID] = true
	}

	// Every non-Chat PageID must have an entry.
	nonChatPages := []PageID{
		PageSettings, PageTools, PageStatus,
		PageSessions, PageTasks, PageApprovals,
	}
	for _, pid := range nonChatPages {
		assert.True(t, metaIDs[pid.String()],
			"AllPageMetas should contain entry for %s", pid.String())
	}
	// Chat must also be present.
	assert.True(t, metaIDs[PageChat.String()], "AllPageMetas should contain entry for chat")
}

func TestAllPageMetas_RoundTrip(t *testing.T) {
	metas := AllPageMetas()
	for _, m := range metas {
		pid := PageIDFromString(m.ID)
		require.NotEqual(t, PageID(99), pid,
			"ID %q should round-trip through PageIDFromString", m.ID)
		assert.Equal(t, m.ID, pid.String(),
			"PageIDFromString(%q).String() should return original ID", m.ID)
	}
}
