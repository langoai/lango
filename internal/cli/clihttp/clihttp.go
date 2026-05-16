// Package clihttp provides shared HTTP/JSON helpers for gateway-backed CLI commands.
package clihttp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// FetchJSON reads JSON from the given gateway path with a bounded timeout.
func FetchJSON(addr, path string, out interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(addr + path)
	if err != nil {
		return fmt.Errorf("connect to gateway: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gateway returned status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
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
