package storagebroker

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	entinquiry "github.com/langoai/lango/internal/ent/inquiry"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	sec "github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
	"github.com/stretchr/testify/require"
)

func TestServerRunSurfacesDecodeAndEncodeFailures(t *testing.T) {
	t.Parallel()

	srv := NewServer()
	err := srv.Run(bytes.NewBufferString(`{"id":`), &bytes.Buffer{})
	require.ErrorContains(t, err, "decode broker request")

	err = srv.Run(bytes.NewBufferString(`{"id":1,"method":"health"}`+"\n"), errWriter{})
	require.ErrorContains(t, err, "encode broker response")
}

func TestServerHealthDBStatusAndShutdownStateTransitions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	srv := NewServer()
	dbPath := t.TempDir() + "/broker.db"

	health := srv.health()
	require.False(t, health.Opened)

	_, err := srv.dispatch(ctx, Request{
		Method:  methodDBStatus,
		Payload: mustPayload(t, DBStatusSummaryRequest{DBPath: dbPath}),
	})
	require.Error(t, err)

	opened, err := srv.dispatch(ctx, Request{
		Method:  methodOpenDB,
		Payload: mustPayload(t, OpenDBRequest{DBPath: dbPath}),
	})
	require.NoError(t, err)
	require.Equal(t, OpenDBResult{Opened: true}, opened)
	require.True(t, srv.health().Opened)

	statusAny, err := srv.dispatch(ctx, Request{
		Method:  methodDBStatus,
		Payload: mustPayload(t, DBStatusSummaryRequest{DBPath: dbPath}),
	})
	require.NoError(t, err)
	require.Equal(t, DBStatusSummaryResult{Available: true}, statusAny)

	shutdown, err := srv.dispatch(ctx, Request{Method: methodShutdown})
	require.NoError(t, err)
	require.Equal(t, ShutdownResult{ShuttingDown: true}, shutdown)
	require.False(t, srv.health().Opened)
}

func TestServerPayloadProtectionCopiesKeyAndDefaultsVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	key := bytes.Repeat([]byte{0x61}, 32)
	originalKey := append([]byte(nil), key...)
	srv := NewServer()
	t.Cleanup(func() {
		_ = srv.shutdown()
	})

	_, err := srv.dispatch(ctx, Request{
		Method: methodOpenDB,
		Payload: mustPayload(t, OpenDBRequest{
			DBPath:     t.TempDir() + "/broker.db",
			PayloadKey: key,
		}),
	})
	require.NoError(t, err)
	for i := range key {
		key[i] = 0
	}

	encAny, err := srv.dispatch(ctx, Request{
		Method:  methodEncryptPayload,
		Payload: mustPayload(t, EncryptPayloadRequest{Plaintext: []byte("copied key payload")}),
	})
	require.NoError(t, err)
	enc := encAny.(EncryptPayloadResult)
	require.Equal(t, sec.PayloadKeyVersionV1, enc.KeyVersion)

	plaintext, err := sec.DecryptPayloadWithKey(originalKey, enc.Ciphertext, enc.Nonce)
	require.NoError(t, err)
	require.Equal(t, []byte("copied key payload"), plaintext)

	_, err = sec.DecryptPayloadWithKey(key, enc.Ciphertext, enc.Nonce)
	require.Error(t, err)
}

func TestServerConfigSessionAndRecallMethodsRequireInitializedStores(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	srv := NewServer()

	tests := []struct {
		name    string
		method  string
		payload any
		wantErr string
	}{
		{
			name:    "config load",
			method:  methodConfigLoad,
			payload: ConfigLoadRequest{Name: "default"},
			wantErr: "database not opened",
		},
		{
			name:    "config load active",
			method:  methodConfigLoadActive,
			wantErr: "database not opened",
		},
		{
			name:   methodConfigSave,
			method: methodConfigSave,
			payload: ConfigSaveRequest{
				Name:   "default",
				Config: []byte(`{}`),
			},
			wantErr: "database not opened",
		},
		{
			name:    methodConfigSetActive,
			method:  methodConfigSetActive,
			payload: ConfigSetActiveRequest{Name: "default"},
			wantErr: "database not opened",
		},
		{
			name:    methodConfigList,
			method:  methodConfigList,
			wantErr: "database not opened",
		},
		{
			name:    methodConfigDelete,
			method:  methodConfigDelete,
			payload: ConfigDeleteRequest{Name: "default"},
			wantErr: "database not opened",
		},
		{
			name:    methodConfigExists,
			method:  methodConfigExists,
			payload: ConfigExistsRequest{Name: "default"},
			wantErr: "database not opened",
		},
		{
			name:   methodSessionCreate,
			method: methodSessionCreate,
			payload: SessionCreateRequest{Session: session.Session{
				Key: "s1",
			}},
			wantErr: "session store not initialized",
		},
		{
			name:    methodSessionGet,
			method:  methodSessionGet,
			payload: SessionGetRequest{Key: "s1"},
			wantErr: "session store not initialized",
		},
		{
			name:   methodSessionUpdate,
			method: methodSessionUpdate,
			payload: SessionUpdateRequest{Session: session.Session{
				Key: "s1",
			}},
			wantErr: "session store not initialized",
		},
		{
			name:    methodSessionDelete,
			method:  methodSessionDelete,
			payload: SessionDeleteRequest{Key: "s1"},
			wantErr: "session store not initialized",
		},
		{
			name:   methodSessionAppend,
			method: methodSessionAppend,
			payload: SessionAppendMessageRequest{
				Key: "s1",
				Message: session.Message{
					Role:    types.RoleUser,
					Content: "hello",
				},
			},
			wantErr: "session store not initialized",
		},
		{
			name:    methodSessionEnd,
			method:  methodSessionEnd,
			payload: SessionEndRequest{Key: "s1"},
			wantErr: "session store not initialized",
		},
		{
			name:    methodSessionList,
			method:  methodSessionList,
			wantErr: "session store not initialized",
		},
		{
			name:    methodSessionGetSalt,
			method:  methodSessionGetSalt,
			payload: SessionGetSaltRequest{Name: "agent"},
			wantErr: "session store not initialized",
		},
		{
			name:    methodSessionSetSalt,
			method:  methodSessionSetSalt,
			payload: SessionSetSaltRequest{Name: "agent", Salt: []byte("salt")},
			wantErr: "session store not initialized",
		},
		{
			name:    methodRecallIndex,
			method:  methodRecallIndex,
			payload: RecallIndexRequest{Key: "s1"},
			wantErr: "recall index not initialized",
		},
		{
			name:    methodRecallProcess,
			method:  methodRecallProcess,
			wantErr: "recall index not initialized",
		},
		{
			name:    methodRecallSearch,
			method:  methodRecallSearch,
			payload: RecallSearchRequest{Query: "hello", Limit: 3},
			wantErr: "recall index not initialized",
		},
		{
			name:    methodRecallSummary,
			method:  methodRecallSummary,
			payload: RecallSummaryRequest{Key: "s1"},
			wantErr: "recall index not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := Request{Method: tt.method}
			if tt.payload != nil {
				req.Payload = mustPayload(t, tt.payload)
			}
			_, err := srv.dispatch(ctx, req)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestServerLearningAndInquiryDispatchReturnsOrderedLimitedRows(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	srv := NewServer()
	t.Cleanup(func() {
		_ = srv.shutdown()
	})
	_, err := srv.dispatch(ctx, Request{
		Method:  methodOpenDB,
		Payload: mustPayload(t, OpenDBRequest{DBPath: t.TempDir() + "/broker.db"}),
	})
	require.NoError(t, err)

	oldTime := time.Now().Add(-2 * time.Hour).Truncate(time.Second)
	newTime := time.Now().Add(-1 * time.Hour).Truncate(time.Second)
	require.NoError(t, srv.client.Learning.Create().
		SetTrigger("old trigger").
		SetCategory(entlearning.CategoryTimeout).
		SetDiagnosis("old diagnosis").
		SetFix("old fix").
		SetConfidence(0.25).
		SetCreatedAt(oldTime).
		Exec(ctx))
	require.NoError(t, srv.client.Learning.Create().
		SetTrigger("new trigger").
		SetCategory(entlearning.CategoryToolError).
		SetDiagnosis("new diagnosis").
		SetFix("new fix").
		SetConfidence(0.75).
		SetCreatedAt(newTime).
		Exec(ctx))
	require.NoError(t, srv.client.Inquiry.Create().
		SetSessionKey("s1").
		SetTopic("pending topic").
		SetQuestion("what should happen next?").
		SetPriority(entinquiry.PriorityHigh).
		SetStatus(entinquiry.StatusPending).
		SetCreatedAt(oldTime).
		Exec(ctx))
	require.NoError(t, srv.client.Inquiry.Create().
		SetSessionKey("s1").
		SetTopic("resolved topic").
		SetQuestion("already answered?").
		SetStatus(entinquiry.StatusResolved).
		Exec(ctx))

	historyAny, err := srv.dispatch(ctx, Request{
		Method:  methodLearningHistory,
		Payload: mustPayload(t, LearningHistoryRequest{Limit: 1}),
	})
	require.NoError(t, err)
	history := historyAny.(LearningHistoryResult)
	require.Len(t, history.Entries, 1)
	require.Equal(t, "new trigger", history.Entries[0].Trigger)
	require.Equal(t, string(entlearning.CategoryToolError), history.Entries[0].Category)
	require.Equal(t, "new diagnosis", history.Entries[0].Diagnosis)
	require.Equal(t, "new fix", history.Entries[0].Fix)
	require.Equal(t, 0.75, history.Entries[0].Confidence)

	inquiriesAny, err := srv.dispatch(ctx, Request{
		Method:  methodPendingInquiries,
		Payload: mustPayload(t, PendingInquiriesRequest{Limit: 5}),
	})
	require.NoError(t, err)
	inquiries := inquiriesAny.(PendingInquiriesResult)
	require.Len(t, inquiries.Entries, 1)
	require.Equal(t, "pending topic", inquiries.Entries[0].Topic)
	require.Equal(t, "what should happen next?", inquiries.Entries[0].Question)
	require.Equal(t, string(entinquiry.PriorityHigh), inquiries.Entries[0].Priority)
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}
