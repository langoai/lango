package alerts

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func executeAlertsCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestAlertsList_TableWritesToCommandOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/alerts", r.URL.Path)
		_, _ = w.Write([]byte(`{
			"alerts": [
				{
					"id": "a-1",
					"type": "policy_block_rate",
					"actor": "system",
					"details": {"severity":"warning","message":"threshold crossed"},
					"timestamp": "2026-05-14T10:00:00Z"
				}
			],
			"total": 1,
			"days": 7
		}`))
	}))
	defer server.Close()

	cmd := NewAlertsCmd()
	out, err := executeAlertsCommand(t, cmd, "list", "--addr", server.URL)

	require.NoError(t, err)
	assert.Contains(t, out, "Alerts (last 7 day(s))")
	assert.Contains(t, out, "policy_block_rate")
	assert.Contains(t, out, "threshold crossed")
}

func TestAlertsSummary_JSONWritesToCommandOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/alerts", r.URL.Path)
		_, _ = w.Write([]byte(`{
			"alerts": [
				{"id":"a-1","type":"policy_block_rate","actor":"system","details":{},"timestamp":"2026-05-14T10:00:00Z"},
				{"id":"a-2","type":"policy_block_rate","actor":"system","details":{},"timestamp":"2026-05-14T11:00:00Z"}
			],
			"total": 2,
			"days": 30
		}`))
	}))
	defer server.Close()

	cmd := NewAlertsCmd()
	out, err := executeAlertsCommand(t, cmd, "summary", "--addr", server.URL, "--output", "json")

	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.EqualValues(t, 2, decoded["totalAlerts"])
	summary := decoded["summary"].(map[string]any)
	assert.EqualValues(t, 2, summary["policy_block_rate"])
}

func TestAlertsList_InvalidOutputRejectsBeforeGatewayCall(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cmd := NewAlertsCmd()
	out, err := executeAlertsCommand(t, cmd, "list", "--addr", server.URL, "--output", "yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
	assert.Empty(t, out)
	assert.Equal(t, int32(0), hits.Load())
}
