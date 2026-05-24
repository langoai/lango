package testutil_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/testutil"
)

func TestMockSessionStoreCRUDAndCopySemantics(t *testing.T) {
	t.Parallel()

	store := testutil.NewMockSessionStore()
	createdAt := time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	giveSession := &session.Session{
		Key:         "session-1",
		AgentID:     "agent-a",
		ChannelType: "cli",
		ChannelID:   "channel-a",
		History: []session.Message{
			{
				Role:      "user",
				Content:   "hello",
				Timestamp: createdAt,
				ToolCalls: []session.ToolCall{
					{ID: "call-1", Name: "tool", ThoughtSignature: []byte{1, 2, 3}},
				},
			},
		},
		Metadata:  map[string]string{"mode": "chat"},
		Model:     "model-a",
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	require.NoError(t, store.Create(giveSession))
	giveSession.AgentID = "mutated-agent"
	giveSession.History[0].Content = "caller mutation"
	giveSession.History[0].ToolCalls[0].ThoughtSignature[0] = 9
	giveSession.History = append(giveSession.History, session.Message{
		Role:    "assistant",
		Content: "caller mutation",
	})
	giveSession.Metadata["mode"] = "mutated"
	giveSession.Model = "mutated-model"

	gotSession, err := store.Get("session-1")
	require.NoError(t, err)
	assert.Equal(t, "agent-a", gotSession.AgentID)
	assert.Equal(t, "model-a", gotSession.Model)
	require.Len(t, gotSession.History, 1)
	assert.Equal(t, "hello", gotSession.History[0].Content)
	require.Len(t, gotSession.History[0].ToolCalls, 1)
	assert.Equal(t, []byte{1, 2, 3}, gotSession.History[0].ToolCalls[0].ThoughtSignature)
	assert.Equal(t, map[string]string{"mode": "chat"}, gotSession.Metadata)

	gotSession.AgentID = "mutated-result"
	gotSession.History[0].Content = "result mutation"
	gotSession.History[0].ToolCalls[0].ThoughtSignature[1] = 8
	gotSession.History = append(gotSession.History, session.Message{
		Role:    "assistant",
		Content: "result mutation",
	})
	gotSession.Metadata["mode"] = "result-mutated"

	gotAgain, err := store.Get("session-1")
	require.NoError(t, err)
	assert.Equal(t, "agent-a", gotAgain.AgentID)
	require.Len(t, gotAgain.History, 1)
	assert.Equal(t, "hello", gotAgain.History[0].Content)
	assert.Equal(t, []byte{1, 2, 3}, gotAgain.History[0].ToolCalls[0].ThoughtSignature)
	assert.Equal(t, map[string]string{"mode": "chat"}, gotAgain.Metadata)

	update := &session.Session{
		Key:     "session-1",
		AgentID: "agent-b",
		History: []session.Message{{
			Role:      "assistant",
			Content:   "updated",
			ToolCalls: []session.ToolCall{{ID: "call-2", ThoughtSignature: []byte{4, 5, 6}}},
		}},
		Metadata:  map[string]string{"mode": "review"},
		Model:     "model-b",
		CreatedAt: createdAt,
		UpdatedAt: updatedAt.Add(time.Minute),
	}
	require.NoError(t, store.Update(update))
	update.AgentID = "mutated-after-update"
	update.History[0].Content = "leak"
	update.History[0].ToolCalls[0].ThoughtSignature[2] = 7
	update.History = append(update.History, session.Message{Role: "user", Content: "leak"})
	update.Metadata["mode"] = "leak"

	gotUpdated, err := store.Get("session-1")
	require.NoError(t, err)
	assert.Equal(t, "agent-b", gotUpdated.AgentID)
	assert.Equal(t, "model-b", gotUpdated.Model)
	require.Len(t, gotUpdated.History, 1)
	assert.Equal(t, "updated", gotUpdated.History[0].Content)
	assert.Equal(t, []byte{4, 5, 6}, gotUpdated.History[0].ToolCalls[0].ThoughtSignature)
	assert.Equal(t, map[string]string{"mode": "review"}, gotUpdated.Metadata)

	assert.True(t, store.HasSession("session-1"))
	assert.Equal(t, 1, store.SessionCount())

	require.NoError(t, store.Delete("session-1"))
	assert.False(t, store.HasSession("session-1"))
	assert.Equal(t, 0, store.SessionCount())
	_, err = store.Get("session-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `session "session-1" not found`)

	assert.Equal(t, 1, store.CreateCalls())
	assert.Equal(t, 4, store.GetCalls())
	assert.Equal(t, 1, store.UpdateCalls())
	assert.Equal(t, 1, store.DeleteCalls())
}

func TestMockSessionStoreMessageTimeoutEndAndList(t *testing.T) {
	t.Parallel()

	store := testutil.NewMockSessionStore()
	firstCreated := time.Date(2026, 5, 19, 11, 0, 0, 0, time.UTC)
	firstUpdated := firstCreated.Add(time.Minute)
	secondCreated := firstCreated.Add(time.Hour)
	secondUpdated := secondCreated.Add(time.Minute)

	require.NoError(t, store.Create(&session.Session{
		Key:       "session-1",
		CreatedAt: firstCreated,
		UpdatedAt: firstUpdated,
	}))
	require.NoError(t, store.Create(&session.Session{
		Key:       "session-2",
		Metadata:  map[string]string{"existing": "value"},
		CreatedAt: secondCreated,
		UpdatedAt: secondUpdated,
	}))

	appended := session.Message{
		Role:    "user",
		Content: "hello",
		ToolCalls: []session.ToolCall{
			{ID: "append-call", ThoughtSignature: []byte{7, 8, 9}},
		},
	}
	require.NoError(t, store.AppendMessage("session-1", appended))
	appended.Content = "caller mutation"
	appended.ToolCalls[0].ThoughtSignature[0] = 0
	require.NoError(t, store.AnnotateTimeout("session-1", "partial response"))
	require.NoError(t, store.AnnotateTimeout("session-2", ""))
	require.NoError(t, store.End("session-1"))
	require.NoError(t, store.End("session-2"))

	gotFirst, err := store.Get("session-1")
	require.NoError(t, err)
	require.Len(t, gotFirst.History, 2)
	assert.Equal(t, "hello", gotFirst.History[0].Content)
	require.Len(t, gotFirst.History[0].ToolCalls, 1)
	assert.Equal(t, []byte{7, 8, 9}, gotFirst.History[0].ToolCalls[0].ThoughtSignature)
	assert.Equal(t, "assistant", string(gotFirst.History[1].Role))
	assert.Equal(t, "partial response\n\n---\n[This response was interrupted due to a timeout]", gotFirst.History[1].Content)
	assert.Equal(t, session.MetadataValueTrue, gotFirst.Metadata[session.MetadataKeyEndPending])

	gotSecond, err := store.Get("session-2")
	require.NoError(t, err)
	require.Len(t, gotSecond.History, 1)
	assert.Equal(t, "[This response was interrupted due to a timeout]", gotSecond.History[0].Content)
	assert.Equal(t, "value", gotSecond.Metadata["existing"])
	assert.Equal(t, session.MetadataValueTrue, gotSecond.Metadata[session.MetadataKeyEndPending])

	summaries, err := store.ListSessions(context.Background())
	require.NoError(t, err)
	require.Len(t, summaries, 2)
	summaryByKey := map[string]session.SessionSummary{
		summaries[0].Key: summaries[0],
		summaries[1].Key: summaries[1],
	}
	assert.Equal(t, firstCreated, summaryByKey["session-1"].CreatedAt)
	assert.Equal(t, firstUpdated, summaryByKey["session-1"].UpdatedAt)
	assert.Equal(t, secondCreated, summaryByKey["session-2"].CreatedAt)
	assert.Equal(t, secondUpdated, summaryByKey["session-2"].UpdatedAt)

	err = store.AppendMessage("missing", session.Message{Role: "user", Content: "lost"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `session "missing" not found`)

	err = store.AnnotateTimeout("missing", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `session "missing" not found`)

	err = store.End("missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `session "missing" not found`)

	assert.Equal(t, 2, store.AppendMessageCalls())
	assert.Equal(t, 3, store.AnnotateTimeoutCalls())
}

func TestMockSessionStoreSalts(t *testing.T) {
	t.Parallel()

	store := testutil.NewMockSessionStore()
	salt := []byte{1, 2, 3}

	require.NoError(t, store.SetSalt("local", salt))
	salt[0] = 9
	gotSalt, err := store.GetSalt("local")
	require.NoError(t, err)
	assert.Equal(t, []byte{1, 2, 3}, gotSalt)
	gotSalt[1] = 8
	gotSaltAgain, err := store.GetSalt("local")
	require.NoError(t, err)
	assert.Equal(t, []byte{1, 2, 3}, gotSaltAgain)

	_, err = store.GetSalt("missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `salt "missing" not found`)

	getErr := errors.New("get salt")
	store.GetSaltErr = getErr
	_, err = store.GetSalt("local")
	require.ErrorIs(t, err, getErr)

	setErr := errors.New("set salt")
	store.SetSaltErr = setErr
	err = store.SetSalt("other", []byte{4})
	require.ErrorIs(t, err, setErr)
}

func TestMockSessionStoreCountersAndErrorInjection(t *testing.T) {
	t.Parallel()

	store := testutil.NewMockSessionStore()
	createErr := errors.New("create")
	getErr := errors.New("get")
	updateErr := errors.New("update")
	deleteErr := errors.New("delete")
	appendErr := errors.New("append")
	timeoutErr := errors.New("timeout")
	closeErr := errors.New("close")

	store.CreateErr = createErr
	store.GetErr = getErr
	store.UpdateErr = updateErr
	store.DeleteErr = deleteErr
	store.AppendMessageErr = appendErr
	store.AnnotateTimeoutErr = timeoutErr
	store.CloseErr = closeErr

	require.ErrorIs(t, store.Create(&session.Session{Key: "session-1"}), createErr)
	_, err := store.Get("session-1")
	require.ErrorIs(t, err, getErr)
	require.ErrorIs(t, store.Update(&session.Session{Key: "session-1"}), updateErr)
	require.ErrorIs(t, store.Delete("session-1"), deleteErr)
	require.ErrorIs(t, store.AppendMessage("session-1", session.Message{}), appendErr)
	require.ErrorIs(t, store.AnnotateTimeout("session-1", ""), timeoutErr)
	require.ErrorIs(t, store.Close(), closeErr)

	assert.Equal(t, 1, store.CreateCalls())
	assert.Equal(t, 1, store.GetCalls())
	assert.Equal(t, 1, store.UpdateCalls())
	assert.Equal(t, 1, store.DeleteCalls())
	assert.Equal(t, 1, store.AppendMessageCalls())
	assert.Equal(t, 1, store.AnnotateTimeoutCalls())
	assert.Equal(t, 1, store.CloseCalls())
	assert.Equal(t, 0, store.SessionCount())
	assert.False(t, store.HasSession("session-1"))
}
