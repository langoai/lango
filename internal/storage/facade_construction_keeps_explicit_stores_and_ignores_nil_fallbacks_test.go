package storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	langoent "github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/session"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

type facadeConstructionKeepsExplicitStoresAndIgnoresNilFallbacksConfigProfiles struct {
	ConfigProfileStore
}

type facadeConstructionKeepsExplicitStoresAndIgnoresNilFallbacksSecurityState struct {
	SecurityStateStore
}

func TestFacadeConstructionKeepsExplicitStoresAndIgnoresNilFallbacks(t *testing.T) {
	t.Parallel()

	profiles := &facadeConstructionKeepsExplicitStoresAndIgnoresNilFallbacksConfigProfiles{}
	securityState := &facadeConstructionKeepsExplicitStoresAndIgnoresNilFallbacksSecurityState{}

	f := NewFacade(
		profiles,
		securityState,
		nil,
		WithEntClient(nil),
		WithRawDB(nil),
		WithSessionClient(nil),
		WithSessionDBPath(""),
		WithBrokerSessionStore(nil),
	)

	require.Same(t, profiles, f.ConfigProfiles())
	require.Same(t, securityState, f.SecurityState())
	require.Nil(t, f.rawDB)

	_, err := f.OpenSessionStore()
	require.ErrorContains(t, err, "session storage unavailable")
}

func TestFacadeSessionFactoryWinsOverDBPathFallback(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("factory selected")
	calls := 0
	f := NewFacade(nil, nil,
		WithSessionStoreFactory(func(opts ...session.StoreOption) (session.Store, error) {
			calls++
			require.Len(t, opts, 1)
			return nil, wantErr
		}),
		WithSessionDBPath(t.TempDir()+"/ignored.db"),
	)

	_, err := f.OpenSessionStore(func(*session.EntStore) {})
	require.ErrorIs(t, err, wantErr)
	require.Equal(t, 1, calls)
}

func TestFacadeBrokerSessionStoreResolution(t *testing.T) {
	t.Parallel()

	f := NewFacade(nil, nil, WithBrokerSessionStore(&runtimeReaderStubBroker{}))

	store, err := f.OpenSessionStore()
	require.NoError(t, err)
	require.IsType(t, &brokerSessionStore{}, store)
}

func TestFacadeRawDBCloseDoesNotOverrideExistingCloser(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	wantErr := errors.New("custom close")
	f := NewFacade(nil, nil,
		func(f *Facade) {
			f.closeFn = func() error {
				return wantErr
			}
		},
		WithRawDB(db),
	)

	require.Same(t, db, f.rawDB)
	require.ErrorIs(t, f.Close(), wantErr)
	require.NoError(t, db.Ping())
}

func TestFacadeResolveEntBackedRequiresClientAndBuilder(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:facade-buildMetaToolsWithEscrowEngineAddsDerivedEscrowTools5-resolve?mode=memory&_fk=1")
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	f := NewFacade(nil, nil, WithEntClient(client))

	resolved, ok := ResolveEntBacked(f, func(c *langoent.Client) *langoent.Client {
		return c
	})
	require.True(t, ok)
	require.Same(t, client, resolved)

	withoutClient, ok := ResolveEntBacked(NewFacade(nil, nil), func(*langoent.Client) int {
		return 1
	})
	require.False(t, ok)
	require.Zero(t, withoutClient)

	missingBuilder, ok := ResolveEntBacked[int](f, nil)
	require.False(t, ok)
	require.Zero(t, missingBuilder)
}

func TestFacadeEntBackedReadersPropagateClosedClientErrors(t *testing.T) {
	ctx := context.Background()
	client := enttest.Open(t, "sqlite3", "file:facade-buildMetaToolsWithEscrowEngineAddsDerivedEscrowTools5-closed?mode=memory&_fk=1")
	f := NewFacade(nil, nil, WithEntClient(client))
	require.NoError(t, client.Close())

	_, err := f.SecuritySummary(ctx)
	require.Error(t, err)
	_, err = f.RecentSandboxDecisions(ctx, "", 1)
	require.Error(t, err)
	_, err = f.LearningHistory(ctx, 1)
	require.Error(t, err)
	_, err = f.PendingInquiries(ctx, 1)
	require.Error(t, err)
	_, err = f.Alerts(ctx, time.Time{})
	require.Error(t, err)
	_, err = f.PaymentHistory(ctx, 1)
	require.Error(t, err)
	_, err = f.PaymentUsage(ctx)
	require.Error(t, err)

	reader := f.WorkflowStateStore(nil)
	require.NotNil(t, reader)
	_, err = reader.ListRuns(ctx, 1)
	require.Error(t, err)
}
