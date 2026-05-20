package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/search"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/storagebroker"
	"github.com/langoai/lango/internal/types"
	"github.com/stretchr/testify/require"
)

func TestBrokerConfigProfiles_DelegatesSuccessPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	createdAt := "2026-05-20T01:02:03Z"
	updatedAt := "2026-05-20T04:05:06Z"
	explicitKeys := map[string]bool{"dataRoot": true}
	broker := &adapterTestBroker{
		configLoadResult: storagebroker.ConfigLoadResult{
			Config:       []byte(`{"dataRoot":"/tmp/lango"}`),
			ExplicitKeys: explicitKeys,
		},
		configLoadActiveResult: storagebroker.ConfigLoadActiveResult{
			Name:         "active",
			Config:       []byte(`{"dataRoot":"/tmp/active"}`),
			ExplicitKeys: explicitKeys,
		},
		configListResult: storagebroker.ConfigListResult{Profiles: []storagebroker.ConfigProfileInfo{{
			Name:      "active",
			Active:    true,
			Version:   7,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}}},
		configExistsResult: storagebroker.ConfigExistsResult{Exists: true},
	}
	store := NewBrokerConfigProfiles(broker)
	require.NotNil(t, store)

	cfg := &config.Config{DataRoot: "/tmp/save"}
	require.NoError(t, store.Save(ctx, "saved", cfg, explicitKeys))
	require.Equal(t, "saved", broker.configSaveName)
	require.Same(t, cfg, broker.configSaveConfig)
	require.Equal(t, explicitKeys, broker.configSaveExplicitKeys)

	loaded, keys, err := store.Load(ctx, "loaded")
	require.NoError(t, err)
	require.Equal(t, "/tmp/lango", loaded.DataRoot)
	require.Equal(t, explicitKeys, keys)
	require.Equal(t, "loaded", broker.configLoadName)

	activeName, active, activeKeys, err := store.LoadActive(ctx)
	require.NoError(t, err)
	require.Equal(t, "active", activeName)
	require.Equal(t, "/tmp/active", active.DataRoot)
	require.Equal(t, explicitKeys, activeKeys)

	require.NoError(t, store.SetActive(ctx, "active"))
	require.Equal(t, "active", broker.configSetActiveName)

	profiles, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, profiles, 1)
	require.Equal(t, "active", profiles[0].Name)
	require.True(t, profiles[0].Active)
	require.Equal(t, 7, profiles[0].Version)
	require.Equal(t, mustParseRFC3339(t, createdAt), profiles[0].CreatedAt)
	require.Equal(t, mustParseRFC3339(t, updatedAt), profiles[0].UpdatedAt)

	exists, err := store.Exists(ctx, "active")
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, "active", broker.configExistsName)

	require.NoError(t, store.Delete(ctx, "active"))
	require.Equal(t, "active", broker.configDeleteName)
	require.Nil(t, NewBrokerConfigProfiles(nil))
}

func TestBrokerConfigProfiles_PropagatesErrorsAndDecodeFailures(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	wantErr := errors.New("broker failed")

	t.Run("load broker error", func(t *testing.T) {
		t.Parallel()
		store := NewBrokerConfigProfiles(&adapterTestBroker{configLoadErr: wantErr})
		_, _, err := store.Load(ctx, "broken")
		require.ErrorIs(t, err, wantErr)
	})

	t.Run("load decode error", func(t *testing.T) {
		t.Parallel()
		store := NewBrokerConfigProfiles(&adapterTestBroker{
			configLoadResult: storagebroker.ConfigLoadResult{Config: []byte(`{`)},
		})
		_, _, err := store.Load(ctx, "broken")
		require.ErrorContains(t, err, `decode broker profile "broken"`)
	})

	t.Run("load active broker error", func(t *testing.T) {
		t.Parallel()
		store := NewBrokerConfigProfiles(&adapterTestBroker{configLoadActiveErr: wantErr})
		_, _, _, err := store.LoadActive(ctx)
		require.ErrorIs(t, err, wantErr)
	})

	t.Run("load active decode error", func(t *testing.T) {
		t.Parallel()
		store := NewBrokerConfigProfiles(&adapterTestBroker{
			configLoadActiveResult: storagebroker.ConfigLoadActiveResult{Config: []byte(`{`)},
		})
		_, _, _, err := store.LoadActive(ctx)
		require.ErrorContains(t, err, "decode active broker profile")
	})

	t.Run("delegated method errors", func(t *testing.T) {
		t.Parallel()
		broker := &adapterTestBroker{
			configSaveErr:      wantErr,
			configSetActiveErr: wantErr,
			configListErr:      wantErr,
			configDeleteErr:    wantErr,
			configExistsErr:    wantErr,
		}
		store := NewBrokerConfigProfiles(broker)

		require.ErrorIs(t, store.Save(ctx, "x", &config.Config{}, nil), wantErr)
		require.ErrorIs(t, store.SetActive(ctx, "x"), wantErr)
		_, err := store.List(ctx)
		require.ErrorIs(t, err, wantErr)
		require.ErrorIs(t, store.Delete(ctx, "x"), wantErr)
		_, err = store.Exists(ctx, "x")
		require.ErrorIs(t, err, wantErr)
	})
}

func TestBrokerSecurityState_DelegatesAndPropagatesErrors(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("security failed")
	broker := &adapterTestBroker{
		securityState: storagebroker.LoadSecurityStateResult{
			Salt:     []byte("salt"),
			Checksum: []byte("checksum"),
			FirstRun: true,
		},
	}
	store := NewBrokerSecurityState(broker)
	require.NotNil(t, store)

	salt, err := store.LoadSalt()
	require.NoError(t, err)
	require.Equal(t, []byte("salt"), salt)

	checksum, err := store.LoadChecksum()
	require.NoError(t, err)
	require.Equal(t, []byte("checksum"), checksum)

	firstRun, err := store.IsFirstRun()
	require.NoError(t, err)
	require.True(t, firstRun)

	require.NoError(t, store.StoreSalt([]byte("new-salt")))
	require.Equal(t, []byte("new-salt"), broker.storedSalt)
	require.NoError(t, store.StoreChecksum([]byte("new-checksum")))
	require.Equal(t, []byte("new-checksum"), broker.storedChecksum)
	require.Nil(t, NewBrokerSecurityState(nil))

	errorStore := NewBrokerSecurityState(&adapterTestBroker{
		loadSecurityStateErr: wantErr,
		storeSaltErr:         wantErr,
		storeChecksumErr:     wantErr,
	})
	_, err = errorStore.LoadSalt()
	require.ErrorIs(t, err, wantErr)
	_, err = errorStore.LoadChecksum()
	require.ErrorIs(t, err, wantErr)
	_, err = errorStore.IsFirstRun()
	require.ErrorIs(t, err, wantErr)
	require.ErrorContains(t, err, "load broker security state")
	require.ErrorIs(t, errorStore.StoreSalt([]byte("x")), wantErr)
	require.ErrorIs(t, errorStore.StoreChecksum([]byte("x")), wantErr)
}

func TestBrokerSessionStore_DelegatesSuccessPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	createdAt := time.Unix(10, 0)
	updatedAt := time.Unix(20, 0)
	wantSession := &session.Session{Key: "session-1", Model: "gpt-test"}
	broker := &adapterTestBroker{
		sessionGetResult: wantSession,
		sessionListResult: []session.SessionSummary{{
			Key:       "session-1",
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}},
		sessionGetSaltResult: []byte("salt"),
		recallSearchResult: []search.SearchResult{{
			RowID: "42",
			Rank:  0.75,
		}},
		recallSummary: "summary",
	}
	store, ok := NewBrokerSessionStore(broker).(*brokerSessionStore)
	require.True(t, ok)
	require.NotNil(t, store)

	require.NoError(t, store.Create(wantSession))
	require.Same(t, wantSession, broker.sessionCreateValue)

	got, err := store.Get("session-1")
	require.NoError(t, err)
	require.Same(t, wantSession, got)
	require.Equal(t, "session-1", broker.sessionGetKey)

	updated := &session.Session{Key: "session-1", Model: "gpt-next"}
	require.NoError(t, store.Update(updated))
	require.Same(t, updated, broker.sessionUpdateValue)

	require.NoError(t, store.Delete("session-1"))
	require.Equal(t, "session-1", broker.sessionDeleteKey)

	msg := session.Message{Role: types.RoleUser, Content: "hello"}
	require.NoError(t, store.AppendMessage("session-1", msg))
	require.Equal(t, "session-1", broker.sessionAppendKey)
	require.Equal(t, msg, broker.sessionAppendMessage)

	require.NoError(t, store.AnnotateTimeout("session-1", "partial"))
	require.Equal(t, "session-1", broker.sessionAppendKey)
	require.Equal(t, types.RoleAssistant, broker.sessionAppendMessage.Role)
	require.Equal(t, "[This response was interrupted due to a timeout]", broker.sessionAppendMessage.Content)

	require.NoError(t, store.End("session-1"))
	require.Equal(t, "session-1", broker.sessionEndKey)

	summaries, err := store.ListSessions(ctx)
	require.NoError(t, err)
	require.Equal(t, broker.sessionListResult, summaries)

	salt, err := store.GetSalt("session-1")
	require.NoError(t, err)
	require.Equal(t, []byte("salt"), salt)
	require.Equal(t, "session-1", broker.sessionGetSaltName)

	require.NoError(t, store.SetSalt("session-1", []byte("new-salt")))
	require.Equal(t, "session-1", broker.sessionSetSaltName)
	require.Equal(t, []byte("new-salt"), broker.sessionSetSaltValue)

	require.NoError(t, store.IndexSession(ctx, "session-1"))
	require.Equal(t, "session-1", broker.recallIndexKey)
	require.NoError(t, store.ProcessPending(ctx))
	require.True(t, broker.recallProcessPendingCalled)

	results, err := store.Search(ctx, "query", 3)
	require.NoError(t, err)
	require.Equal(t, broker.recallSearchResult, results)
	require.Equal(t, "query", broker.recallSearchQuery)
	require.Equal(t, 3, broker.recallSearchLimit)

	summary, err := store.GetSummary(ctx, "session-1")
	require.NoError(t, err)
	require.Equal(t, "summary", summary)
	require.Equal(t, "session-1", broker.recallSummaryKey)

	require.NoError(t, store.Close())
	require.Nil(t, NewBrokerSessionStore(nil))
}

func TestBrokerSessionStore_PropagatesErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	wantErr := errors.New("session failed")
	broker := &adapterTestBroker{
		sessionCreateErr:        wantErr,
		sessionGetErr:           wantErr,
		sessionUpdateErr:        wantErr,
		sessionDeleteErr:        wantErr,
		sessionAppendErr:        wantErr,
		sessionEndErr:           wantErr,
		sessionListErr:          wantErr,
		sessionGetSaltErr:       wantErr,
		sessionSetSaltErr:       wantErr,
		recallIndexErr:          wantErr,
		recallProcessPendingErr: wantErr,
		recallSearchErr:         wantErr,
		recallSummaryErr:        wantErr,
	}
	store, ok := NewBrokerSessionStore(broker).(*brokerSessionStore)
	require.True(t, ok)

	require.ErrorIs(t, store.Create(&session.Session{Key: "session-1"}), wantErr)
	_, err := store.Get("session-1")
	require.ErrorIs(t, err, wantErr)
	require.ErrorIs(t, store.Update(&session.Session{Key: "session-1"}), wantErr)
	require.ErrorIs(t, store.Delete("session-1"), wantErr)
	require.ErrorIs(t, store.AppendMessage("session-1", session.Message{}), wantErr)
	require.ErrorIs(t, store.AnnotateTimeout("session-1", ""), wantErr)
	require.ErrorIs(t, store.End("session-1"), wantErr)
	_, err = store.ListSessions(ctx)
	require.ErrorIs(t, err, wantErr)
	_, err = store.GetSalt("session-1")
	require.ErrorIs(t, err, wantErr)
	require.ErrorIs(t, store.SetSalt("session-1", []byte("salt")), wantErr)
	require.ErrorIs(t, store.IndexSession(ctx, "session-1"), wantErr)
	require.ErrorIs(t, store.ProcessPending(ctx), wantErr)
	_, err = store.Search(ctx, "query", 3)
	require.ErrorIs(t, err, wantErr)
	_, err = store.GetSummary(ctx, "session-1")
	require.ErrorIs(t, err, wantErr)
}

func TestBrokerSessionStore_EndRunsProcessorAfterBrokerEnd(t *testing.T) {
	t.Parallel()

	broker := &adapterTestBroker{}
	store, ok := NewBrokerSessionStore(broker).(*brokerSessionStore)
	require.True(t, ok)

	var processedKey string
	store.SetSessionEndProcessor(func(_ context.Context, key string) error {
		processedKey = key
		return nil
	})

	require.NoError(t, store.End("session-1"))
	require.Equal(t, "session-1", broker.sessionEndKey)
	require.Equal(t, "session-1", processedKey)

	wantErr := errors.New("processor failed")
	store.SetSessionEndProcessor(func(context.Context, string) error {
		return wantErr
	})
	require.ErrorIs(t, store.End("session-2"), wantErr)
}

type adapterTestBroker struct {
	runtimeReaderStubBroker

	securityState        storagebroker.LoadSecurityStateResult
	loadSecurityStateErr error
	storedSalt           []byte
	storeSaltErr         error
	storedChecksum       []byte
	storeChecksumErr     error

	configLoadName         string
	configLoadResult       storagebroker.ConfigLoadResult
	configLoadErr          error
	configLoadActiveResult storagebroker.ConfigLoadActiveResult
	configLoadActiveErr    error
	configSaveName         string
	configSaveConfig       any
	configSaveExplicitKeys map[string]bool
	configSaveErr          error
	configSetActiveName    string
	configSetActiveErr     error
	configListResult       storagebroker.ConfigListResult
	configListErr          error
	configDeleteName       string
	configDeleteErr        error
	configExistsName       string
	configExistsResult     storagebroker.ConfigExistsResult
	configExistsErr        error

	sessionCreateValue         *session.Session
	sessionCreateErr           error
	sessionGetKey              string
	sessionGetResult           *session.Session
	sessionGetErr              error
	sessionUpdateValue         *session.Session
	sessionUpdateErr           error
	sessionDeleteKey           string
	sessionDeleteErr           error
	sessionAppendKey           string
	sessionAppendMessage       session.Message
	sessionAppendErr           error
	sessionEndKey              string
	sessionEndErr              error
	sessionListResult          []session.SessionSummary
	sessionListErr             error
	sessionGetSaltName         string
	sessionGetSaltResult       []byte
	sessionGetSaltErr          error
	sessionSetSaltName         string
	sessionSetSaltValue        []byte
	sessionSetSaltErr          error
	recallIndexKey             string
	recallIndexErr             error
	recallProcessPendingCalled bool
	recallProcessPendingErr    error
	recallSearchQuery          string
	recallSearchLimit          int
	recallSearchResult         []search.SearchResult
	recallSearchErr            error
	recallSummaryKey           string
	recallSummary              string
	recallSummaryErr           error
}

func (b *adapterTestBroker) LoadSecurityState(context.Context) (storagebroker.LoadSecurityStateResult, error) {
	if b.loadSecurityStateErr != nil {
		return storagebroker.LoadSecurityStateResult{}, b.loadSecurityStateErr
	}
	return b.securityState, nil
}

func (b *adapterTestBroker) StoreSalt(_ context.Context, salt []byte) error {
	b.storedSalt = salt
	return b.storeSaltErr
}

func (b *adapterTestBroker) StoreChecksum(_ context.Context, checksum []byte) error {
	b.storedChecksum = checksum
	return b.storeChecksumErr
}

func (b *adapterTestBroker) ConfigLoad(_ context.Context, name string) (storagebroker.ConfigLoadResult, error) {
	b.configLoadName = name
	if b.configLoadErr != nil {
		return storagebroker.ConfigLoadResult{}, b.configLoadErr
	}
	return b.configLoadResult, nil
}

func (b *adapterTestBroker) ConfigLoadActive(context.Context) (storagebroker.ConfigLoadActiveResult, error) {
	if b.configLoadActiveErr != nil {
		return storagebroker.ConfigLoadActiveResult{}, b.configLoadActiveErr
	}
	return b.configLoadActiveResult, nil
}

func (b *adapterTestBroker) ConfigSave(_ context.Context, name string, cfg any, explicitKeys map[string]bool) error {
	b.configSaveName = name
	b.configSaveConfig = cfg
	b.configSaveExplicitKeys = explicitKeys
	return b.configSaveErr
}

func (b *adapterTestBroker) ConfigSetActive(_ context.Context, name string) error {
	b.configSetActiveName = name
	return b.configSetActiveErr
}

func (b *adapterTestBroker) ConfigList(context.Context) (storagebroker.ConfigListResult, error) {
	if b.configListErr != nil {
		return storagebroker.ConfigListResult{}, b.configListErr
	}
	return b.configListResult, nil
}

func (b *adapterTestBroker) ConfigDelete(_ context.Context, name string) error {
	b.configDeleteName = name
	return b.configDeleteErr
}

func (b *adapterTestBroker) ConfigExists(_ context.Context, name string) (storagebroker.ConfigExistsResult, error) {
	b.configExistsName = name
	if b.configExistsErr != nil {
		return storagebroker.ConfigExistsResult{}, b.configExistsErr
	}
	return b.configExistsResult, nil
}

func (b *adapterTestBroker) SessionCreate(_ context.Context, sess *session.Session) error {
	b.sessionCreateValue = sess
	return b.sessionCreateErr
}

func (b *adapterTestBroker) SessionGet(_ context.Context, key string) (*session.Session, error) {
	b.sessionGetKey = key
	if b.sessionGetErr != nil {
		return nil, b.sessionGetErr
	}
	return b.sessionGetResult, nil
}

func (b *adapterTestBroker) SessionUpdate(_ context.Context, sess *session.Session) error {
	b.sessionUpdateValue = sess
	return b.sessionUpdateErr
}

func (b *adapterTestBroker) SessionDelete(_ context.Context, key string) error {
	b.sessionDeleteKey = key
	return b.sessionDeleteErr
}

func (b *adapterTestBroker) SessionAppendMessage(_ context.Context, key string, msg session.Message) error {
	b.sessionAppendKey = key
	b.sessionAppendMessage = msg
	return b.sessionAppendErr
}

func (b *adapterTestBroker) SessionEnd(_ context.Context, key string) error {
	b.sessionEndKey = key
	return b.sessionEndErr
}

func (b *adapterTestBroker) SessionList(context.Context) ([]session.SessionSummary, error) {
	if b.sessionListErr != nil {
		return nil, b.sessionListErr
	}
	return b.sessionListResult, nil
}

func (b *adapterTestBroker) SessionGetSalt(_ context.Context, name string) ([]byte, error) {
	b.sessionGetSaltName = name
	if b.sessionGetSaltErr != nil {
		return nil, b.sessionGetSaltErr
	}
	return b.sessionGetSaltResult, nil
}

func (b *adapterTestBroker) SessionSetSalt(_ context.Context, name string, salt []byte) error {
	b.sessionSetSaltName = name
	b.sessionSetSaltValue = salt
	return b.sessionSetSaltErr
}

func (b *adapterTestBroker) RecallIndexSession(_ context.Context, key string) error {
	b.recallIndexKey = key
	return b.recallIndexErr
}

func (b *adapterTestBroker) RecallProcessPending(context.Context) error {
	b.recallProcessPendingCalled = true
	return b.recallProcessPendingErr
}

func (b *adapterTestBroker) RecallSearch(_ context.Context, query string, limit int) ([]search.SearchResult, error) {
	b.recallSearchQuery = query
	b.recallSearchLimit = limit
	if b.recallSearchErr != nil {
		return nil, b.recallSearchErr
	}
	return b.recallSearchResult, nil
}

func (b *adapterTestBroker) RecallGetSummary(_ context.Context, key string) (string, error) {
	b.recallSummaryKey = key
	if b.recallSummaryErr != nil {
		return "", b.recallSummaryErr
	}
	return b.recallSummary, nil
}

func mustParseRFC3339(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339, value)
	require.NoError(t, err)
	return parsed
}
