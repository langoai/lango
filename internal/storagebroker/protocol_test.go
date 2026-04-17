package storagebroker

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsBrokerMode(t *testing.T) {
	assert.False(t, IsBrokerMode())
}

func TestRequestResponseJSON(t *testing.T) {
	req := Request{ID: 1, Method: methodHealth}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded Request
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, req.ID, decoded.ID)
	assert.Equal(t, req.Method, decoded.Method)
}

func TestServerHealthAndShutdown(t *testing.T) {
	srv := NewServer()

	health, err := srv.dispatch(context.Background(), Request{Method: methodHealth})
	require.NoError(t, err)
	assert.Equal(t, HealthResult{Opened: false}, health)

	shutdown, err := srv.dispatch(context.Background(), Request{Method: methodShutdown})
	require.NoError(t, err)
	assert.Equal(t, ShutdownResult{ShuttingDown: true}, shutdown)
}

func TestServerDBStatusSummary_RequiresPath(t *testing.T) {
	srv := NewServer()
	_, err := srv.dispatch(context.Background(), Request{
		Method: methodDBStatus,
	})
	require.Error(t, err)
}

func TestServerSecurityStateRequiresOpenDB(t *testing.T) {
	srv := NewServer()
	_, err := srv.dispatch(context.Background(), Request{Method: methodLoadSecurityState})
	require.Error(t, err)
}

func TestServerRunRoundTrip(t *testing.T) {
	srv := NewServer()

	var in bytes.Buffer
	require.NoError(t, json.NewEncoder(&in).Encode(Request{ID: 1, Method: methodHealth}))
	require.NoError(t, json.NewEncoder(&in).Encode(Request{ID: 2, Method: methodShutdown}))

	var out bytes.Buffer
	require.NoError(t, srv.Run(&in, &out))

	dec := json.NewDecoder(&out)
	var resp1, resp2 Response
	require.NoError(t, dec.Decode(&resp1))
	require.NoError(t, dec.Decode(&resp2))

	assert.True(t, resp1.OK)
	assert.True(t, resp2.OK)
}

func TestServerOpenDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/broker.db"

	srv := NewServer()
	result, err := srv.openDB(context.Background(), OpenDBRequest{DBPath: dbPath})
	require.NoError(t, err)
	assert.True(t, result.Opened)

	_, statErr := os.Stat(dbPath)
	require.NoError(t, statErr)

	_ = srv.shutdown()
}
