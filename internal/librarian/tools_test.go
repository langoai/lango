package librarian

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTools_RequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	tools := BuildTools(nil)

	got, err := findLibrarianTool(t, tools, "librarian_dismiss_inquiry").Handler(
		context.Background(),
		map[string]interface{}{},
	)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.EqualError(t, err, "missing inquiry_id parameter")
}

func TestBuildTools_PendingInquiriesUseCurrentSessionWhenSessionKeyOmitted(t *testing.T) {
	t.Parallel()

	client := testutil.TestEntClient(t)
	store := NewInquiryStore(client, testutil.NopLogger())
	require.NoError(t, store.SaveInquiry(context.Background(), Inquiry{
		ID:         uuid.New(),
		SessionKey: "session-1",
		Topic:      "runtime",
		Question:   "first session question",
		Priority:   "high",
	}))
	require.NoError(t, store.SaveInquiry(context.Background(), Inquiry{
		ID:         uuid.New(),
		SessionKey: "session-2",
		Topic:      "runtime",
		Question:   "second session question",
		Priority:   "high",
	}))

	tools := BuildTools(store)
	got, err := findLibrarianTool(t, tools, "librarian_pending_inquiries").Handler(
		session.WithSessionKey(context.Background(), "session-1"),
		map[string]interface{}{},
	)
	require.NoError(t, err)

	payload := got.(map[string]interface{})
	inquiries := payload["inquiries"].([]Inquiry)
	require.Len(t, inquiries, 1)
	assert.Equal(t, "session-1", inquiries[0].SessionKey)
	assert.Equal(t, "first session question", inquiries[0].Question)
}

func findLibrarianTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("tool %q not found", name)
	return nil
}
