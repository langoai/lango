package smartaccount

import (
	"database/sql"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent/enttest"
	sa "github.com/langoai/lango/internal/smartaccount"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/testutil"

	_ "github.com/mattn/go-sqlite3"
)

func TestSessionExecutionHelpersWrapBootstrapErrors(t *testing.T) {
	bootErr := errors.New("boot failed")
	loader := func() (*bootstrap.Result, error) {
		return nil, bootErr
	}

	createResult, createCleanup, createErr := executeSessionCreate(loader, nil, nil, "0", "1h")
	require.Error(t, createErr)
	assert.Contains(t, createErr.Error(), "bootstrap: boot failed")
	assert.Equal(t, sessionCreateResult{}, createResult)
	assert.Nil(t, createCleanup)

	listResult, listCleanup, listErr := loadSessionList(loader)
	require.Error(t, listErr)
	assert.Contains(t, listErr.Error(), "bootstrap: boot failed")
	assert.Nil(t, listResult)
	assert.Nil(t, listCleanup)

	message, revokeCleanup, revokeErr := executeSessionRevoke(loader, false, "session-1")
	require.Error(t, revokeErr)
	assert.Contains(t, revokeErr.Error(), "bootstrap: boot failed")
	assert.Empty(t, message)
	assert.Nil(t, revokeCleanup)
}

func TestSessionExecutionHelpersCloseBootOnInitFailure(t *testing.T) {
	tests := []struct {
		name string
		run  func(BootLoader) error
	}{
		{
			name: "create",
			run: func(loader BootLoader) error {
				_, _, err := executeSessionCreate(loader, nil, nil, "0", "1h")
				return err
			},
		},
		{
			name: "list",
			run: func(loader BootLoader) error {
				_, _, err := loadSessionList(loader)
				return err
			},
		},
		{
			name: "revoke",
			run: func(loader BootLoader) error {
				_, _, err := executeSessionRevoke(loader, true, "")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := sql.Open("sqlite3", ":memory:")
			require.NoError(t, err)

			boot := &bootstrap.Result{
				Config:  config.DefaultConfig(),
				Storage: storage.NewFacade(nil, nil, storage.WithRawDB(db)),
			}
			err = tt.run(func() (*bootstrap.Result, error) {
				return boot, nil
			})

			require.Error(t, err)
			assert.Contains(t, err.Error(), "smart account not enabled")
			assert.Error(t, db.Ping(), "bootstrap result storage should be closed on init failure")
		})
	}
}

func TestExecuteSessionCreateValidatesInputsAfterInitializingDeps(t *testing.T) {
	tests := []struct {
		name      string
		targets   []string
		limit     string
		duration  string
		wantError string
	}{
		{
			name:      "invalid duration",
			limit:     "0",
			duration:  "not-a-duration",
			wantError: `parse duration "not-a-duration"`,
		},
		{
			name:      "invalid spend limit",
			limit:     "ten",
			duration:  "1h",
			wantError: `parse spend limit "ten"`,
		},
		{
			name:      "invalid target",
			targets:   []string{"not-an-address"},
			limit:     "0",
			duration:  "1h",
			wantError: "invalid target address: not-an-address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, cleanup, err := executeSessionCreate(
				newSessionExecutionHelpersWrapBootstrapErrorsSessionBootLoader(t),
				tt.targets,
				nil,
				tt.limit,
				tt.duration,
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantError)
			assert.Equal(t, sessionCreateResult{}, result)
			assert.Nil(t, cleanup)
		})
	}
}

func TestExecuteSessionCreateReturnsResultAndCleanup(t *testing.T) {
	target := "0x000000000000000000000000000000000000aaaa"
	result, cleanup, err := executeSessionCreate(
		newSessionExecutionHelpersWrapBootstrapErrorsSessionBootLoader(t),
		[]string{target},
		[]string{"0xa9059cbb"},
		"42",
		"30m",
	)
	require.NoError(t, err)
	require.NotNil(t, cleanup)
	defer cleanup()

	assert.NotEmpty(t, result.ID)
	assert.True(t, common.IsHexAddress(result.Address))
	assert.Equal(t, []string{common.HexToAddress(target).Hex()}, result.Targets)
	assert.Equal(t, []string{"0xa9059cbb"}, result.Functions)
	assert.Equal(t, "42", result.Limit)

	createdAt, err := time.Parse(time.RFC3339, result.CreatedAt)
	require.NoError(t, err)
	expiresAt, err := time.Parse(time.RFC3339, result.ExpiresAt)
	require.NoError(t, err)
	assert.WithinDuration(t, createdAt.Add(30*time.Minute), expiresAt, time.Second)
}

func TestExecuteSessionRevokeBranchesWithoutLiveChainRPC(t *testing.T) {
	t.Run("all succeeds with empty store", func(t *testing.T) {
		message, cleanup, err := executeSessionRevoke(newSessionExecutionHelpersWrapBootstrapErrorsSessionBootLoader(t), true, "")
		require.NoError(t, err)
		require.NotNil(t, cleanup)
		defer cleanup()
		assert.Equal(t, "All active session keys revoked.", message)
	})

	t.Run("single missing session wraps store error", func(t *testing.T) {
		message, cleanup, err := executeSessionRevoke(newSessionExecutionHelpersWrapBootstrapErrorsSessionBootLoader(t), false, "missing-session")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "revoke session missing-session")
		assert.Empty(t, message)
		assert.Nil(t, cleanup)
	})
}

func TestBuildSessionListEntriesMapsRealSessionStatus(t *testing.T) {
	now := time.Now()
	activeAddress := common.HexToAddress("0x000000000000000000000000000000000000a001")
	revokedAddress := common.HexToAddress("0x000000000000000000000000000000000000a002")
	expiredAddress := common.HexToAddress("0x000000000000000000000000000000000000a003")

	entries := buildSessionListEntries([]*sa.SessionKey{
		{
			ID:        "active-session",
			Address:   activeAddress,
			CreatedAt: now.Add(-3 * time.Hour),
			ExpiresAt: now.Add(time.Hour),
			Policy:    sa.SessionPolicy{SpendLimit: nil},
		},
		{
			ID:        "revoked-session",
			Address:   revokedAddress,
			ParentID:  "parent-session",
			CreatedAt: now.Add(-2 * time.Hour),
			ExpiresAt: now.Add(time.Hour),
			Revoked:   true,
			Policy:    sa.SessionPolicy{SpendLimit: big.NewInt(99)},
		},
		{
			ID:        "expired-session",
			Address:   expiredAddress,
			CreatedAt: now.Add(-time.Hour),
			ExpiresAt: now.Add(-time.Minute),
			Policy:    sa.SessionPolicy{SpendLimit: big.NewInt(7)},
		},
	})

	require.Len(t, entries, 3)
	assert.Equal(t, sessionListEntry{
		ID:        "active-session",
		Address:   activeAddress.Hex(),
		ExpiresAt: now.Add(time.Hour).Format(time.RFC3339),
		Limit:     "unlimited",
		Status:    "active",
	}, entries[0])
	assert.Equal(t, "revoked", entries[1].Status)
	assert.Equal(t, "parent-session", entries[1].ParentID)
	assert.Equal(t, "99", entries[1].Limit)
	assert.Equal(t, "expired", entries[2].Status)
	assert.Equal(t, "7", entries[2].Limit)
}

func TestSessionListStatusMappingAndTableFormatting(t *testing.T) {
	original := loadSessionList
	loadSessionList = func(_ BootLoader) ([]sessionListEntry, func(), error) {
		return []sessionListEntry{
			{
				ID:        "active-session-123456",
				Address:   "0x1234567890abcdef1234567890abcdef12345678",
				ExpiresAt: "2026-05-15T00:00:00Z",
				Limit:     "unlimited",
				Status:    "active",
			},
			{
				ID:        "revoked-session-123456",
				Address:   "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				ParentID:  "parent-session-abcdef",
				ExpiresAt: "2026-05-16T00:00:00Z",
				Limit:     "99",
				Status:    "revoked",
			},
			{
				ID:        "expired-session-123456",
				Address:   "0xfedcbafedcbafedcbafedcbafedcbafedcbafedc",
				ExpiresAt: "2026-05-14T00:00:00Z",
				Limit:     "7",
				Status:    "expired",
			},
		}, func() {}, nil
	}
	t.Cleanup(func() { loadSessionList = original })

	cmd := sessionListCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "ADDRESS")
	assert.Contains(t, out, "PARENT")
	assert.Contains(t, out, "SPEND_LIMIT")
	assert.Contains(t, out, "active-s...")
	assert.Contains(t, out, "revoked-...")
	assert.Contains(t, out, "expired-...")
	assert.Contains(t, out, "0x12345678...")
	assert.Contains(t, out, "parent-s...")
	assert.Contains(t, out, "unlimited")
	assert.Contains(t, out, "99")
	assert.Contains(t, out, "active")
	assert.Contains(t, out, "revoked")
	assert.Contains(t, out, "expired")
}

func TestSessionRevokeRejectsTooManyArgsBeforeExecute(t *testing.T) {
	original := executeSessionRevoke
	called := false
	executeSessionRevoke = func(_ BootLoader, _ bool, _ string) (string, func(), error) {
		called = true
		return "", nil, nil
	}
	t.Cleanup(func() { executeSessionRevoke = original })

	cmd := sessionRevokeCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "session-1", "session-2")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts at most 1 arg")
	assert.Empty(t, out)
	assert.False(t, called)
}

func TestSessionCreateRunsCleanupWhenTableFlushFails(t *testing.T) {
	original := executeSessionCreate
	cleanupCalled := false
	executeSessionCreate = func(_ BootLoader, _, _ []string, _, _ string) (sessionCreateResult, func(), error) {
		return sessionCreateResult{
			ID:        "session-1",
			Address:   "0x1234abcd5678ef901234abcdef567890abcdef12",
			Limit:     "0",
			ExpiresAt: "2026-05-15T00:00:00Z",
			CreatedAt: "2026-05-14T00:00:00Z",
		}, func() { cleanupCalled = true }, nil
	}
	t.Cleanup(func() { executeSessionCreate = original })

	cmd := sessionCreateCmd(nil)
	cmd.SetOut(errorWriter{err: errors.New("writer failed")})
	cmd.SetErr(io.Discard)
	cmd.SetArgs(nil)

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "writer failed")
	assert.True(t, cleanupCalled)
}

func TestTablePreviewBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		prefixLen int
		want      string
	}{
		{name: "short", value: "abc", prefixLen: 8, want: "abc"},
		{name: "exact", value: "abcdefgh", prefixLen: 8, want: "abcdefgh"},
		{name: "long", value: "abcdefghi", prefixLen: 8, want: "abcdefgh..."},
		{name: "zero prefix", value: "abc", prefixLen: 0, want: "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tablePreview(tt.value, tt.prefixLen))
		})
	}
}

type errorWriter struct {
	err error
}

func (w errorWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

func newSessionExecutionHelpersWrapBootstrapErrorsSessionBootLoader(t *testing.T) BootLoader {
	t.Helper()

	rpc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"jsonrpc":"2.0","id":1,"result":"0x0"}`)
	}))
	t.Cleanup(rpc.Close)

	client := enttest.Open(t, "sqlite3", "file:"+sessionExecutionHelpersWrapBootstrapErrorsSQLiteName(t.Name())+"?mode=memory&_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = true
	cfg.Payment.Network.RPCURL = rpc.URL
	cfg.Payment.Network.ChainID = 84532
	cfg.Payment.WalletProvider = "local"
	cfg.SmartAccount.Enabled = true
	cfg.SmartAccount.EntryPointAddress = "0x0000000071727De22E5E9d8BAf0edAc6f37da032"
	cfg.SmartAccount.FactoryAddress = "0x000000000000000000000000000000000000fa01"
	cfg.SmartAccount.Safe7579Address = "0x0000000000000000000000000000000000007579"
	cfg.SmartAccount.FallbackHandler = "0x000000000000000000000000000000000000fa11"
	cfg.SmartAccount.BundlerURL = rpc.URL
	cfg.SmartAccount.Session.MaxDuration = time.Hour
	cfg.SmartAccount.Session.MaxActiveKeys = 10

	boot := &bootstrap.Result{
		Config:  cfg,
		Crypto:  testutil.NewMockCryptoProvider(),
		Storage: storage.NewFacade(nil, nil, storage.WithEntClient(client)),
	}

	return func() (*bootstrap.Result, error) {
		return boot, nil
	}
}

func sessionExecutionHelpersWrapBootstrapErrorsSQLiteName(name string) string {
	replacer := strings.NewReplacer("/", "-", " ", "-", ":", "-")
	return replacer.Replace(name)
}
