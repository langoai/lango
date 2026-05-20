package storagebroker

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServerRunStopsAfterShutdownResponse(t *testing.T) {
	t.Parallel()

	var input bytes.Buffer
	enc := json.NewEncoder(&input)
	require.NoError(t, enc.Encode(Request{ID: 5701, Method: methodHealth}))
	require.NoError(t, enc.Encode(Request{ID: 5702, Method: methodShutdown}))
	require.NoError(t, enc.Encode(Request{ID: 5703, Method: methodHealth}))

	var output bytes.Buffer
	require.NoError(t, NewServer().Run(&input, &output))

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	require.Len(t, lines, 2)

	var health Response
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &health))
	require.Equal(t, uint64(5701), health.ID)
	require.True(t, health.OK)

	var shutdown Response
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &shutdown))
	require.Equal(t, uint64(5702), shutdown.ID)
	require.True(t, shutdown.OK)
	require.JSONEq(t, `{"shutting_down":true}`, string(shutdown.Result))
}

func TestServerSecurityStateAndOpenDBIdempotency(t *testing.T) {
	t.Parallel()

	srv := NewServer()
	dbPath := t.TempDir() + "/broker.db"
	openReq := OpenDBRequest{
		DBPath:         dbPath,
		PayloadKey:     bytes.Repeat([]byte{0x57}, 32),
		PayloadVersion: 57,
	}
	ctx := context.Background()
	first, err := srv.dispatch(ctx, Request{
		Method:  methodOpenDB,
		Payload: mustPayload(t, openReq),
	})
	require.NoError(t, err)
	require.True(t, first.(OpenDBResult).Opened)
	firstClient := srv.client
	firstDB := srv.rawDB

	second, err := srv.dispatch(ctx, Request{
		Method:  methodOpenDB,
		Payload: mustPayload(t, openReq),
	})
	require.NoError(t, err)
	require.True(t, second.(OpenDBResult).Opened)
	require.Same(t, firstClient, srv.client)
	require.Same(t, firstDB, srv.rawDB)

	require.NoError(t, srv.sessionSetSalt(SessionSetSaltRequest{Name: "profile", Salt: []byte("profile-salt")}))
	salt, err := srv.sessionGetSalt(SessionGetSaltRequest{Name: "profile"})
	require.NoError(t, err)
	require.Equal(t, []byte("profile-salt"), salt.Salt)

	_, err = srv.dispatch(ctx, Request{
		Method:  methodStoreSalt,
		Payload: mustPayload(t, StoreSaltRequest{Salt: []byte("security-salt")}),
	})
	require.NoError(t, err)
	_, err = srv.dispatch(ctx, Request{
		Method:  methodStoreChecksum,
		Payload: mustPayload(t, StoreChecksumRequest{Checksum: []byte("checksum")}),
	})
	require.NoError(t, err)
	state, err := srv.dispatch(ctx, Request{Method: methodLoadSecurityState})
	require.NoError(t, err)
	require.Equal(t, []byte("security-salt"), state.(LoadSecurityStateResult).Salt)
	require.Equal(t, []byte("checksum"), state.(LoadSecurityStateResult).Checksum)

	t.Cleanup(func() {
		_ = srv.shutdown()
	})
}

func TestServerConfigDeleteAndExistsBranches(t *testing.T) {
	t.Parallel()

	srv := openServerSecurityAndConfigWrappersServer(t, true)
	cfg := serverSecurityAndConfigWrappersConfig(t, "config-branch-model")
	ctx := context.Background()

	_, err := srv.dispatch(ctx, Request{
		Method: methodConfigSave,
		Payload: mustPayload(t, ConfigSaveRequest{
			Name:         "config-branch",
			Config:       mustRawJSON(t, cfg),
			ExplicitKeys: map[string]bool{"agent.model": true},
		}),
	})
	require.NoError(t, err)

	exists, err := srv.configExists(ctx, ConfigExistsRequest{Name: "config-branch"})
	require.NoError(t, err)
	require.True(t, exists.Exists)
	require.NoError(t, srv.configDelete(ctx, ConfigDeleteRequest{Name: "config-branch"}))

	exists, err = srv.configExists(ctx, ConfigExistsRequest{Name: "config-branch"})
	require.NoError(t, err)
	require.False(t, exists.Exists)
}
