package workspace

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockTripleAdder struct {
	triples []Triple
	err     error
}

func (m *mockTripleAdder) add(ctx context.Context, triples []Triple) error {
	if m.err != nil {
		return m.err
	}
	m.triples = append(m.triples, triples...)
	return nil
}

func TestChronicler_Record(t *testing.T) {
	mock := &mockTripleAdder{}
	logger := zap.NewNop().Sugar()
	chronicler := NewChronicler(mock.add, logger)

	msg := Message{
		ID:          "msg-001",
		Type:        MessageTypeTaskProposal,
		WorkspaceID: "ws-1",
		SenderDID:   "did:lango:agent-a",
		Content:     "propose task X",
		Timestamp:   time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
	}

	err := chronicler.Record(context.Background(), msg)
	require.NoError(t, err)

	// Verify base triples were generated.
	assert.GreaterOrEqual(t, len(mock.triples), 5)

	subject := "workspace:message:msg-001"
	tripleMap := make(map[string]string, len(mock.triples))
	for _, tr := range mock.triples {
		assert.Equal(t, subject, tr.Subject)
		tripleMap[tr.Predicate] = tr.Object
	}

	assert.Equal(t, string(MessageTypeTaskProposal), tripleMap["type"])
	assert.Equal(t, "ws-1", tripleMap["workspace"])
	assert.Equal(t, "did:lango:agent-a", tripleMap["sender"])
	assert.Equal(t, "propose task X", tripleMap["content"])
	assert.Equal(t, "2026-03-10T12:00:00Z", tripleMap["timestamp"])
}

func TestChronicler_Record_WithParent(t *testing.T) {
	mock := &mockTripleAdder{}
	logger := zap.NewNop().Sugar()
	chronicler := NewChronicler(mock.add, logger)

	msg := Message{
		ID:          "msg-002",
		Type:        MessageTypeLogStream,
		WorkspaceID: "ws-1",
		SenderDID:   "did:lango:agent-a",
		Content:     "reply message",
		ParentID:    "msg-001",
		Timestamp:   time.Now(),
	}

	err := chronicler.Record(context.Background(), msg)
	require.NoError(t, err)

	// Find the replyTo triple.
	var found bool
	for _, tr := range mock.triples {
		if tr.Predicate == "replyTo" {
			assert.Equal(t, "workspace:message:msg-001", tr.Object)
			found = true
			break
		}
	}
	assert.True(t, found, "replyTo triple should be present")
}

func TestChronicler_Record_WithMetadata(t *testing.T) {
	mock := &mockTripleAdder{}
	logger := zap.NewNop().Sugar()
	chronicler := NewChronicler(mock.add, logger)

	msg := Message{
		ID:          "msg-003",
		Type:        MessageTypeKnowledgeShare,
		WorkspaceID: "ws-1",
		SenderDID:   "did:lango:agent-b",
		Content:     "sharing knowledge",
		Metadata:    map[string]string{"priority": "high", "category": "architecture"},
		Timestamp:   time.Now(),
	}

	err := chronicler.Record(context.Background(), msg)
	require.NoError(t, err)

	// Verify meta: triples.
	metaTriples := make(map[string]string)
	for _, tr := range mock.triples {
		if len(tr.Predicate) > 5 && tr.Predicate[:5] == "meta:" {
			metaTriples[tr.Predicate] = tr.Object
		}
	}

	assert.Equal(t, "high", metaTriples["meta:priority"])
	assert.Equal(t, "architecture", metaTriples["meta:category"])
}

func TestChronicler_Record_NilAdder(t *testing.T) {
	logger := zap.NewNop().Sugar()
	chronicler := NewChronicler(nil, logger)

	msg := Message{
		ID:          "msg-004",
		Type:        MessageTypeLogStream,
		WorkspaceID: "ws-1",
		SenderDID:   "did:lango:agent-a",
		Content:     "should not error",
		Timestamp:   time.Now(),
	}

	err := chronicler.Record(context.Background(), msg)
	require.NoError(t, err)
}

func TestChronicler_Record_AdderError(t *testing.T) {
	mock := &mockTripleAdder{err: errors.New("store unavailable")}
	logger := zap.NewNop().Sugar()
	chronicler := NewChronicler(mock.add, logger)

	msg := Message{
		ID:          "msg-005",
		Type:        MessageTypeLogStream,
		WorkspaceID: "ws-1",
		SenderDID:   "did:lango:agent-a",
		Content:     "should error",
		Timestamp:   time.Now(),
	}

	err := chronicler.Record(context.Background(), msg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "persist message triples")
}
