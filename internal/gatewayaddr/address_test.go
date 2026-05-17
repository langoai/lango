package gatewayaddr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPURLFormatsIPv6Host(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "http://[::1]:18789", HTTPURL("::1", 18789))
}

func TestHTTPURLNormalizesBracketedIPv6Host(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "http://[::1]:18789", HTTPURL("[::1]", 18789))
	assert.Equal(t, "[::1]:18789", ListenAddress("[::1]", 18789))
}

func TestHTTPURLFallsBackWhenHostOrPortUnset(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "http://localhost:18789", HTTPURL("  ", 0))
}

func TestListenAddressFormatsIPv6Host(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "[::1]:18789", ListenAddress("::1", 18789))
}

func TestDialHTTPURLUsesLoopbackForWildcardHosts(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "http://localhost:18789", DialHTTPURL("0.0.0.0", 18789))
	assert.Equal(t, "http://localhost:18789", DialHTTPURL("::", 18789))
}
