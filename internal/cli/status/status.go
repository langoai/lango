// Package status implements the lango status command — a unified status dashboard.
package status

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/postadjudicationstatus"
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/langoai/lango/internal/types"
)

const defaultAddr = "http://localhost:18789"

type deadLetterBridge interface {
	List(context.Context, deadLetterListOptions) (deadLetterListPage, error)
	Detail(context.Context, string) (postadjudicationstatus.TransactionStatus, error)
	Retry(context.Context, string) error
}

type deadLetterBridgeLoader func() (deadLetterBridge, func(), error)

type deadLetterListOptions struct {
	Query                     string
	Adjudication              string
	LatestStatusSubtype       string
	LatestStatusSubtypeFamily string
	ManualReplayActor         string
	DeadLetteredAfter         string
	DeadLetteredBefore        string
	DeadLetterReasonQuery     string
	LatestDispatchReference   string
}

type deadLetterListPage struct {
	Entries []postadjudicationstatus.DeadLetterBacklogEntry `json:"entries"`
	Count   int                                             `json:"count"`
	Total   int                                             `json:"total"`
	Offset  int                                             `json:"offset"`
	Limit   int                                             `json:"limit"`
}

type deadLetterRetryResult struct {
	TransactionReceiptID string `json:"transaction_receipt_id"`
	Result               string `json:"result"`
	Message              string `json:"message"`
}

type toolCatalogDeadLetterBridge struct {
	catalog *toolcatalog.Catalog
}

// NewStatusCmd creates the status command.
func NewStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		outputFmt string
		addr      string
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show unified system status dashboard",
		Long: `Show a unified status dashboard combining health, config, and metrics.

When the server is running, fetches live data from the gateway.
When the server is not running, shows configuration-based status only.

Examples:
  lango status              # Full status dashboard
  lango status --output json  # Machine-readable JSON output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			info := collectStatus(boot.Config, boot.ProfileName, addr)
			info.Version = tui.GetVersion()

			if outputFmt == "json" {
				return printJSON(info)
			}
			fmt.Print(renderDashboard(info))
			return nil
		},
	}

	cmd.Flags().StringVar(&outputFmt, "output", "table", "Output format: table or json")
	cmd.Flags().StringVar(&addr, "addr", defaultAddr, "Gateway address")
	cmd.AddCommand(newDeadLettersCmd(deadLetterLoaderFromBoot(bootLoader)))
	cmd.AddCommand(newDeadLetterCmd(deadLetterLoaderFromBoot(bootLoader)))
	return cmd
}

func newDeadLettersCmd(loader deadLetterBridgeLoader) *cobra.Command {
	var (
		outputFmt                 string
		query                     string
		adjudication              string
		latestStatusSubtype       string
		latestStatusSubtypeFamily string
		manualReplayActor         string
		deadLetteredAfter         string
		deadLetteredBefore        string
		deadLetterReasonQuery     string
		latestDispatchReference   string
	)

	cmd := &cobra.Command{
		Use:   "dead-letters",
		Short: "List dead-lettered post-adjudication executions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			subtype, err := normalizeLatestStatusSubtype(latestStatusSubtype)
			if err != nil {
				return err
			}
			family, err := normalizeLatestStatusSubtypeFamily(latestStatusSubtypeFamily)
			if err != nil {
				return err
			}
			after, err := normalizeRFC3339Flag("dead-lettered-after", deadLetteredAfter)
			if err != nil {
				return err
			}
			before, err := normalizeRFC3339Flag("dead-lettered-before", deadLetteredBefore)
			if err != nil {
				return err
			}

			bridge, cleanup, err := loader()
			if err != nil {
				return err
			}
			defer cleanup()

			page, err := bridge.List(cmd.Context(), deadLetterListOptions{
				Query:                     query,
				Adjudication:              adjudication,
				LatestStatusSubtype:       subtype,
				LatestStatusSubtypeFamily: family,
				ManualReplayActor:         strings.TrimSpace(manualReplayActor),
				DeadLetteredAfter:         after,
				DeadLetteredBefore:        before,
				DeadLetterReasonQuery:     strings.TrimSpace(deadLetterReasonQuery),
				LatestDispatchReference:   strings.TrimSpace(latestDispatchReference),
			})
			if err != nil {
				return err
			}
			if outputFmt == "json" {
				return printJSONTo(cmd.OutOrStdout(), page)
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), renderDeadLetterBacklogTable(page))
			return err
		},
	}
	cmd.Flags().StringVar(&outputFmt, "output", "table", "Output format: table or json")
	cmd.Flags().StringVar(&query, "query", "", "Substring filter for transaction or submission receipt IDs")
	cmd.Flags().StringVar(&adjudication, "adjudication", "", "Adjudication outcome filter: release or refund")
	cmd.Flags().StringVar(&latestStatusSubtype, "latest-status-subtype", "", "Latest status subtype filter: retry-scheduled, manual-retry-requested, or dead-lettered")
	cmd.Flags().StringVar(&latestStatusSubtypeFamily, "latest-status-subtype-family", "", "Latest status subtype family filter: retry, manual-retry, or dead-letter")
	cmd.Flags().StringVar(&manualReplayActor, "manual-replay-actor", "", "Latest manual replay actor filter")
	cmd.Flags().StringVar(&deadLetteredAfter, "dead-lettered-after", "", "Latest dead-letter lower-bound timestamp filter (RFC3339)")
	cmd.Flags().StringVar(&deadLetteredBefore, "dead-lettered-before", "", "Latest dead-letter upper-bound timestamp filter (RFC3339)")
	cmd.Flags().StringVar(&deadLetterReasonQuery, "dead-letter-reason-query", "", "Latest dead-letter reason substring filter")
	cmd.Flags().StringVar(&latestDispatchReference, "latest-dispatch-reference", "", "Latest dispatch reference exact-match filter")
	return cmd
}

func newDeadLetterCmd(loader deadLetterBridgeLoader) *cobra.Command {
	var outputFmt string

	cmd := &cobra.Command{
		Use:   "dead-letter <transaction-receipt-id>",
		Short: "Show dead-letter status for a transaction",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bridge, cleanup, err := loader()
			if err != nil {
				return err
			}
			defer cleanup()

			status, err := bridge.Detail(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if outputFmt == "json" {
				return printJSONTo(cmd.OutOrStdout(), status)
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), renderDeadLetterDetail(status))
			return err
		},
	}
	cmd.Flags().StringVar(&outputFmt, "output", "table", "Output format: table or json")
	cmd.AddCommand(newDeadLetterRetryCmd(loader))
	return cmd
}

func newDeadLetterRetryCmd(loader deadLetterBridgeLoader) *cobra.Command {
	var (
		outputFmt string
		yes       bool
	)

	cmd := &cobra.Command{
		Use:   "retry <transaction-receipt-id>",
		Short: "Retry a dead-lettered post-adjudication execution",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			transactionReceiptID := strings.TrimSpace(args[0])

			bridge, cleanup, err := loader()
			if err != nil {
				return err
			}
			defer cleanup()

			status, err := bridge.Detail(cmd.Context(), transactionReceiptID)
			if err != nil {
				return fmt.Errorf("read dead-letter status for transaction %q: %w", transactionReceiptID, err)
			}
			if !status.CanRetry {
				return fmt.Errorf("retry precheck rejected for transaction %q: can_retry=false", transactionReceiptID)
			}

			if !yes {
				confirmed, err := confirmDeadLetterRetry(cmd, transactionReceiptID)
				if err != nil {
					return err
				}
				if !confirmed {
					_, err := fmt.Fprintln(cmd.OutOrStdout(), "Retry aborted.")
					return err
				}
			}

			if err := bridge.Retry(cmd.Context(), transactionReceiptID); err != nil {
				return fmt.Errorf("retry request failed for transaction %q: %w", transactionReceiptID, err)
			}

			message := fmt.Sprintf("Retry request accepted for transaction %s.", transactionReceiptID)
			result := deadLetterRetryResult{
				TransactionReceiptID: transactionReceiptID,
				Result:               "accepted",
				Message:              message,
			}
			if outputFmt == "json" {
				return printJSONTo(cmd.OutOrStdout(), result)
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), message)
			return err
		},
	}
	cmd.Flags().StringVar(&outputFmt, "output", "table", "Output format: table or json")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	return cmd
}

func deadLetterLoaderFromBoot(bootLoader func() (*bootstrap.Result, error)) deadLetterBridgeLoader {
	return func() (deadLetterBridge, func(), error) {
		boot, err := bootLoader()
		if err != nil {
			return nil, nil, fmt.Errorf("bootstrap: %w", err)
		}
		application, err := app.New(boot, app.WithLocalChat())
		if err != nil {
			_ = boot.Close()
			return nil, nil, fmt.Errorf("build app: %w", err)
		}
		cleanup := func() {
			_ = application.Stop(context.Background())
			_ = boot.Close()
		}
		bridge := &toolCatalogDeadLetterBridge{catalog: application.ToolCatalog}
		if !bridge.ready() {
			cleanup()
			return nil, nil, fmt.Errorf("dead-letter status tools are not available")
		}
		return bridge, cleanup, nil
	}
}

func (b *toolCatalogDeadLetterBridge) ready() bool {
	if b == nil || b.catalog == nil {
		return false
	}
	_, hasList := b.catalog.Get("list_dead_lettered_post_adjudication_executions")
	_, hasDetail := b.catalog.Get("get_post_adjudication_execution_status")
	return hasList && hasDetail
}

func (b *toolCatalogDeadLetterBridge) List(ctx context.Context, opts deadLetterListOptions) (deadLetterListPage, error) {
	if b == nil || b.catalog == nil {
		return deadLetterListPage{}, fmt.Errorf("dead-letter tool catalog is not configured")
	}
	entry, ok := b.catalog.Get("list_dead_lettered_post_adjudication_executions")
	if !ok || entry.Tool == nil || entry.Tool.Handler == nil {
		return deadLetterListPage{}, fmt.Errorf("dead-letter backlog tool is not available")
	}
	params := map[string]interface{}{}
	if query := strings.TrimSpace(opts.Query); query != "" {
		params["query"] = query
	}
	switch strings.TrimSpace(opts.Adjudication) {
	case "release", "refund":
		params["adjudication"] = strings.TrimSpace(opts.Adjudication)
	}
	switch strings.TrimSpace(opts.LatestStatusSubtype) {
	case "retry-scheduled", "manual-retry-requested", "dead-lettered":
		params["latest_status_subtype"] = strings.TrimSpace(opts.LatestStatusSubtype)
	}
	switch strings.TrimSpace(opts.LatestStatusSubtypeFamily) {
	case "retry", "manual-retry", "dead-letter":
		params["latest_status_subtype_family"] = strings.TrimSpace(opts.LatestStatusSubtypeFamily)
	}
	if actor := strings.TrimSpace(opts.ManualReplayActor); actor != "" {
		params["manual_replay_actor"] = actor
	}
	if after := strings.TrimSpace(opts.DeadLetteredAfter); after != "" {
		params["dead_lettered_after"] = after
	}
	if before := strings.TrimSpace(opts.DeadLetteredBefore); before != "" {
		params["dead_lettered_before"] = before
	}
	if reasonQuery := strings.TrimSpace(opts.DeadLetterReasonQuery); reasonQuery != "" {
		params["dead_letter_reason_query"] = reasonQuery
	}
	if dispatchReference := strings.TrimSpace(opts.LatestDispatchReference); dispatchReference != "" {
		params["latest_dispatch_reference"] = dispatchReference
	}
	raw, err := entry.Tool.Handler(ctx, params)
	if err != nil {
		return deadLetterListPage{}, err
	}
	payload, ok := raw.(map[string]interface{})
	if !ok {
		return deadLetterListPage{}, fmt.Errorf("dead-letter backlog tool returned invalid payload")
	}
	entriesRaw, ok := payload["entries"]
	if !ok {
		return deadLetterListPage{}, fmt.Errorf("dead-letter backlog tool returned no entries")
	}
	data, err := json.Marshal(entriesRaw)
	if err != nil {
		return deadLetterListPage{}, err
	}
	var entries []postadjudicationstatus.DeadLetterBacklogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return deadLetterListPage{}, err
	}
	return deadLetterListPage{
		Entries: entries,
		Count:   optionalInt(payload, "count"),
		Total:   optionalInt(payload, "total"),
		Offset:  optionalInt(payload, "offset"),
		Limit:   optionalInt(payload, "limit"),
	}, nil
}

func normalizeLatestStatusSubtype(value string) (string, error) {
	switch strings.TrimSpace(value) {
	case "", "retry-scheduled", "manual-retry-requested", "dead-lettered":
		return strings.TrimSpace(value), nil
	default:
		return "", fmt.Errorf("invalid --latest-status-subtype %q: must be one of retry-scheduled, manual-retry-requested, dead-lettered", value)
	}
}

func normalizeLatestStatusSubtypeFamily(value string) (string, error) {
	switch strings.TrimSpace(value) {
	case "", "retry", "manual-retry", "dead-letter":
		return strings.TrimSpace(value), nil
	default:
		return "", fmt.Errorf("invalid --latest-status-subtype-family %q: must be one of retry, manual-retry, dead-letter", value)
	}
}

func normalizeRFC3339Flag(name, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}
	if _, err := time.Parse(time.RFC3339, trimmed); err != nil {
		return "", fmt.Errorf("invalid --%s %q: must be RFC3339", name, value)
	}
	return trimmed, nil
}

func (b *toolCatalogDeadLetterBridge) Detail(ctx context.Context, transactionReceiptID string) (postadjudicationstatus.TransactionStatus, error) {
	if b == nil || b.catalog == nil {
		return postadjudicationstatus.TransactionStatus{}, fmt.Errorf("dead-letter tool catalog is not configured")
	}
	entry, ok := b.catalog.Get("get_post_adjudication_execution_status")
	if !ok || entry.Tool == nil || entry.Tool.Handler == nil {
		return postadjudicationstatus.TransactionStatus{}, fmt.Errorf("dead-letter detail tool is not available")
	}
	raw, err := entry.Tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": strings.TrimSpace(transactionReceiptID),
	})
	if err != nil {
		return postadjudicationstatus.TransactionStatus{}, err
	}
	if status, ok := raw.(postadjudicationstatus.TransactionStatus); ok {
		return status, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return postadjudicationstatus.TransactionStatus{}, err
	}
	var status postadjudicationstatus.TransactionStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return postadjudicationstatus.TransactionStatus{}, err
	}
	return status, nil
}

func (b *toolCatalogDeadLetterBridge) Retry(ctx context.Context, transactionReceiptID string) error {
	if b == nil || b.catalog == nil {
		return fmt.Errorf("dead-letter tool catalog is not configured")
	}
	entry, ok := b.catalog.Get("retry_post_adjudication_execution")
	if !ok || entry.Tool == nil || entry.Tool.Handler == nil {
		return fmt.Errorf("dead-letter retry tool is not available")
	}
	_, err := entry.Tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": strings.TrimSpace(transactionReceiptID),
	})
	return err
}

func confirmDeadLetterRetry(cmd *cobra.Command, transactionReceiptID string) (bool, error) {
	if _, err := fmt.Fprintf(
		cmd.OutOrStdout(),
		"Retry dead-lettered execution for transaction %s? [y/N]: ",
		transactionReceiptID,
	); err != nil {
		return false, err
	}
	reader := bufio.NewReader(cmd.InOrStdin())
	answer, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}

func optionalInt(payload map[string]interface{}, key string) int {
	value, ok := payload[key]
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

// StatusInfo holds all collected status data.
type StatusInfo struct {
	Version        string        `json:"version"`
	Profile        string        `json:"profile"`
	ContextProfile string        `json:"contextProfile,omitempty"`
	ServerUp       bool          `json:"serverUp"`
	Gateway        string        `json:"gateway"`
	Provider       string        `json:"provider"`
	Model          string        `json:"model"`
	Features       []FeatureInfo `json:"features"`
	Channels       []string      `json:"channels"`
	ServerInfo     *LiveInfo     `json:"serverInfo,omitempty"`
}

// FeatureInfo describes a feature's status.
type FeatureInfo struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Detail  string `json:"detail,omitempty"`
}

// LiveInfo holds data fetched from a running server.
type LiveInfo struct {
	Healthy  bool                  `json:"healthy"`
	Uptime   string                `json:"uptime,omitempty"`
	Features []types.FeatureStatus `json:"features,omitempty"`
}

func collectStatus(cfg *config.Config, profile, addr string) StatusInfo {
	info := StatusInfo{
		Profile:        profile,
		ContextProfile: string(cfg.ContextProfile),
		Gateway:        fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port),
		Provider:       cfg.Agent.Provider,
		Model:          cfg.Agent.Model,
	}

	// Check server health.
	info.ServerUp, info.ServerInfo = probeServer(addr)

	// Collect channels.
	if cfg.Channels.Telegram.Enabled {
		info.Channels = append(info.Channels, "telegram")
	}
	if cfg.Channels.Discord.Enabled {
		info.Channels = append(info.Channels, "discord")
	}
	if cfg.Channels.Slack.Enabled {
		info.Channels = append(info.Channels, "slack")
	}

	// Collect features.
	info.Features = collectFeatures(cfg)

	// Enrich feature details with live runtime statuses when available.
	if info.ServerInfo != nil && len(info.ServerInfo.Features) > 0 {
		liveByName := make(map[string]types.FeatureStatus, len(info.ServerInfo.Features))
		for _, fs := range info.ServerInfo.Features {
			liveByName[fs.Name] = fs
		}
		for i := range info.Features {
			if live, ok := liveByName[info.Features[i].Name]; ok && info.Features[i].Detail == "" {
				if live.Reason != "" {
					info.Features[i].Detail = live.Reason
				}
			}
		}
	}

	return info
}

func profileDetail(cfg *config.Config) string {
	if cfg.ContextProfile != "" {
		return "profile: " + string(cfg.ContextProfile)
	}
	return ""
}

func collectFeatures(cfg *config.Config) []FeatureInfo {
	pd := profileDetail(cfg)
	return []FeatureInfo{
		{"Knowledge", cfg.Knowledge.Enabled, pd},
		{"Embedding & RAG", cfg.Embedding.Provider != "", cfg.Embedding.Provider},
		{"Graph", cfg.Graph.Enabled, pd},
		{"Obs. Memory", cfg.ObservationalMemory.Enabled, pd},
		{"Librarian", cfg.Librarian.Enabled, pd},
		{"Multi-Agent", cfg.Agent.MultiAgent, ""},
		{"Cron", cfg.Cron.Enabled, ""},
		{"Background", cfg.Background.Enabled, ""},
		{"Workflow", cfg.Workflow.Enabled, ""},
		{"MCP", cfg.MCP.Enabled, mcpDetail(cfg)},
		{"P2P", cfg.P2P.Enabled, ""},
		{"Payment", cfg.Payment.Enabled, ""},
		{"Economy", cfg.Economy.Enabled, ""},
		{"A2A", cfg.A2A.Enabled, ""},
		{"RunLedger", cfg.RunLedger.Enabled, ""},
		{"Provenance", cfg.Provenance.Enabled, ""},
	}
}

func mcpDetail(cfg *config.Config) string {
	if !cfg.MCP.Enabled {
		return ""
	}
	n := len(cfg.MCP.Servers)
	if n == 0 {
		return "no servers"
	}
	return fmt.Sprintf("%d server(s)", n)
}

func probeServer(addr string) (bool, *LiveInfo) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(addr + "/health")
	if err != nil {
		return false, nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	live := &LiveInfo{Healthy: true}

	// Try to parse feature statuses from health response.
	var healthResp struct {
		Features []types.FeatureStatus `json:"features"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err == nil {
		live.Features = healthResp.Features
	}
	return true, live
}

func printJSON(v interface{}) error {
	return printJSONTo(os.Stdout, v)
}

func printJSONTo(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
