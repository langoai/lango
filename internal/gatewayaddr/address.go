// Package gatewayaddr formats configured gateway host/port pairs consistently.
package gatewayaddr

import (
	"net"
	"strconv"
	"strings"
)

const (
	DefaultHost = "localhost"
	DefaultPort = 18789
)

// HTTPURL formats a gateway client URL from configured host and port.
func HTTPURL(host string, port int) string {
	return "http://" + net.JoinHostPort(resolveHost(host), strconv.Itoa(resolvePort(port)))
}

// DialHTTPURL formats a gateway URL for local reachability checks.
func DialHTTPURL(host string, port int) string {
	resolvedHost := resolveHost(host)
	if isWildcardHost(resolvedHost) {
		resolvedHost = DefaultHost
	}
	return "http://" + net.JoinHostPort(resolvedHost, strconv.Itoa(resolvePort(port)))
}

// ListenAddress formats the gateway server listen address.
func ListenAddress(host string, port int) string {
	return net.JoinHostPort(normalizeHost(strings.TrimSpace(host)), strconv.Itoa(resolvePort(port)))
}

func resolveHost(host string) string {
	if trimmed := strings.TrimSpace(host); trimmed != "" {
		return normalizeHost(trimmed)
	}
	return DefaultHost
}

func normalizeHost(host string) string {
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		unbracketed := strings.TrimSuffix(strings.TrimPrefix(host, "["), "]")
		if net.ParseIP(unbracketed) != nil {
			return unbracketed
		}
	}
	return host
}

func resolvePort(port int) int {
	if port > 0 {
		return port
	}
	return DefaultPort
}

func isWildcardHost(host string) bool {
	parsed := net.ParseIP(strings.Trim(host, "[]"))
	if parsed == nil {
		return false
	}
	return parsed.IsUnspecified()
}
