package p2p

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

func newProvenanceCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provenance",
		Short: "Exchange signed provenance bundles with peers",
		Long:  "Push or fetch signed provenance bundles through the running gateway using existing authenticated P2P sessions.",
	}

	cmd.AddCommand(newProvenancePushCmd(bootLoader))
	cmd.AddCommand(newProvenanceFetchCmd(bootLoader))
	return cmd
}

func newProvenancePushCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		addr      string
		redaction string
	)

	cmd := &cobra.Command{
		Use:   "push <peer-did> <session-key>",
		Short: "Push a signed provenance bundle to a peer",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()
			if !boot.Config.P2P.Enabled {
				return errP2PDisabled
			}
			if addr == "" {
				addr = fmt.Sprintf("http://%s:%d", boot.Config.Server.Host, boot.Config.Server.Port)
			}
			body := map[string]string{
				"peerDid":    args[0],
				"sessionKey": args[1],
				"redaction":  redaction,
			}
			var out map[string]any
			if err := postJSON(addr, "/api/p2p/provenance/push", body, &out); err != nil {
				return err
			}
			fmt.Printf("Pushed provenance bundle to %s (redaction=%s)\n", args[0], redaction)
			return nil
		},
	}
	cmd.Flags().StringVar(&addr, "addr", "", "Gateway address (default: config server host/port)")
	cmd.Flags().StringVar(&redaction, "redaction", "content", "Redaction level (none, content, full)")
	return cmd
}

func newProvenanceFetchCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		addr      string
		redaction string
	)

	cmd := &cobra.Command{
		Use:   "fetch <peer-did> <session-key>",
		Short: "Fetch and import a signed provenance bundle from a peer",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()
			if !boot.Config.P2P.Enabled {
				return errP2PDisabled
			}
			if addr == "" {
				addr = fmt.Sprintf("http://%s:%d", boot.Config.Server.Host, boot.Config.Server.Port)
			}
			body := map[string]string{
				"peerDid":    args[0],
				"sessionKey": args[1],
				"redaction":  redaction,
			}
			var out map[string]any
			if err := postJSON(addr, "/api/p2p/provenance/fetch", body, &out); err != nil {
				return err
			}
			fmt.Printf("Fetched provenance bundle from %s (redaction=%v)\n", args[0], out["redaction"])
			return nil
		},
	}
	cmd.Flags().StringVar(&addr, "addr", "", "Gateway address (default: config server host/port)")
	cmd.Flags().StringVar(&redaction, "redaction", "content", "Redaction level (none, content, full)")
	return cmd
}

func postJSON(addr, path string, body any, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(addr+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("connect to gateway: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var payload map[string]any
		if json.Unmarshal(body, &payload) == nil {
			if msg, ok := payload["error"].(string); ok && msg != "" {
				return errors.New(msg)
			}
			if msg, ok := payload["message"].(string); ok && msg != "" {
				return errors.New(msg)
			}
		}
		if trimmed := strings.TrimSpace(string(body)); trimmed != "" {
			return errors.New(trimmed)
		}
		return fmt.Errorf("gateway returned status %d", resp.StatusCode)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
