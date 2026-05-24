package checks

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestCompanionConnectionCheckUsesLoopbackForWildcardBindHost(t *testing.T) {
	t.Parallel()

	seenHost := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/status", r.URL.Path)
		seenHost <- r.Host
		require.NoError(t, json.NewEncoder(w).Encode(map[string]int{"clients": 0}))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	_, portText, err := net.SplitHostPort(serverURL.Host)
	require.NoError(t, err)
	port, err := strconv.Atoi(portText)
	require.NoError(t, err)

	check := &CompanionConnectionCheck{}
	result := check.Run(context.Background(), &config.Config{
		Server: config.ServerConfig{
			Host:             "0.0.0.0",
			Port:             port,
			WebSocketEnabled: true,
		},
	})

	assert.Equal(t, StatusPass, result.Status, result.Message)
	assert.Equal(t, "localhost:"+portText, <-seenHost)
}
