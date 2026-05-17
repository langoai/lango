// Package clihttp provides shared HTTP/JSON helpers for gateway-backed CLI commands.
package clihttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/gatewayaddr"
)

// DefaultGatewayAddr is the CLI fallback when no configured gateway is available.
const DefaultGatewayAddr = "http://localhost:18789"

// FetchJSON reads JSON from the given gateway path with a bounded timeout.
func FetchJSON(addr, path string, out interface{}) error {
	return FetchJSONContext(context.Background(), addr, path, out)
}

// FetchJSONContext reads JSON from the given gateway path with a bounded timeout.
func FetchJSONContext(ctx context.Context, addr, path string, out interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr+path, nil)
	if err != nil {
		return fmt.Errorf("build gateway request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connect to gateway: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return decodeGatewayError(resp)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// PostJSON posts a JSON request body to the given gateway path and decodes JSON.
func PostJSON(addr, path string, body interface{}, out interface{}) error {
	return PostJSONContext(context.Background(), addr, path, body, out)
}

// PostJSONContext posts a JSON request body to the given gateway path and decodes JSON.
func PostJSONContext(ctx context.Context, addr, path string, body interface{}, out interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, addr+path, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("build gateway request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connect to gateway: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return decodeGatewayError(resp)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// ResolveGatewayAddr returns an explicit CLI address or derives one from server config.
func ResolveGatewayAddr(explicit string, cfg *config.Config) string {
	if addr := strings.TrimSpace(explicit); addr != "" {
		return strings.TrimRight(addr, "/")
	}
	host := gatewayaddr.DefaultHost
	port := gatewayaddr.DefaultPort
	if cfg != nil {
		if configuredHost := strings.TrimSpace(cfg.Server.Host); configuredHost != "" {
			host = configuredHost
		}
		if cfg.Server.Port > 0 {
			port = cfg.Server.Port
		}
	}
	return gatewayaddr.HTTPURL(host, port)
}

func decodeGatewayError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var payload map[string]interface{}
	if json.Unmarshal(body, &payload) == nil {
		for _, key := range []string{"error", "message"} {
			if msg, ok := payload[key].(string); ok && strings.TrimSpace(msg) != "" {
				return errors.New(msg)
			}
		}
	}
	return fmt.Errorf("gateway returned status %d", resp.StatusCode)
}

// ResolveTableOrJSONOutput validates a Cobra --output flag that only accepts table or json.
func ResolveTableOrJSONOutput(cmd *cobra.Command) (string, error) {
	flag, _ := cmd.Flags().GetString("output")
	switch normalized := strings.ToLower(strings.TrimSpace(flag)); normalized {
	case "", "table":
		return "table", nil
	case "json":
		return "json", nil
	default:
		return "", fmt.Errorf("unknown output format %q (expected: table or json)", strings.TrimSpace(flag))
	}
}

// PrintJSON writes indented JSON to the given writer.
func PrintJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
