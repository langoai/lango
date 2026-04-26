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
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/postadjudicationstatus"
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/langoai/lango/internal/types"
)

const defaultAddr = "http://localhost:18789"

const (
	defaultDeadLetterSummaryTopN       = 5
	defaultDeadLetterTrendWindow       = 24 * time.Hour
	defaultDeadLetterTrendBucket       = 6 * time.Hour
	defaultDeadLetterRetryWaitInterval = 2 * time.Second
	defaultDeadLetterRetryWaitTimeout  = 30 * time.Second
)

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
	AnyMatchFamily            string
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
	TransactionReceiptID string                   `json:"transaction_receipt_id"`
	Result               string                   `json:"result"`
	Message              string                   `json:"message"`
	FollowUp             *deadLetterRetryFollowUp `json:"follow_up,omitempty"`
	FollowUpError        string                   `json:"follow_up_error,omitempty"`
	PollCount            int                      `json:"poll_count,omitempty"`
	TimedOut             bool                     `json:"timed_out,omitempty"`
}

type deadLetterRetryFollowUp struct {
	ObservedAt                string                                       `json:"observed_at,omitempty"`
	IsDeadLettered            bool                                         `json:"is_dead_lettered"`
	CanRetry                  bool                                         `json:"can_retry"`
	LatestStatusSubtype       string                                       `json:"latest_status_subtype,omitempty"`
	LatestStatusSubtypeFamily string                                       `json:"latest_status_subtype_family,omitempty"`
	LatestDeadLetterReason    string                                       `json:"latest_dead_letter_reason,omitempty"`
	LatestRetryAttempt        int                                          `json:"latest_retry_attempt,omitempty"`
	LatestDispatchReference   string                                       `json:"latest_dispatch_reference,omitempty"`
	BackgroundTask            *postadjudicationstatus.BackgroundTaskBridge `json:"background_task,omitempty"`
}

type deadLetterSummaryBucket struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type deadLetterReasonSummaryItem struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

type deadLetterActorSummaryItem struct {
	Actor string `json:"actor"`
	Count int    `json:"count"`
}

type deadLetterDispatchSummaryItem struct {
	DispatchReference string `json:"dispatch_reference"`
	Count             int    `json:"count"`
}

type deadLetterSummaryResult struct {
	TotalDeadLetters            int                             `json:"total_dead_letters"`
	RetryableCount              int                             `json:"retryable_count"`
	TopLimit                    int                             `json:"top_limit"`
	ByAdjudication              []deadLetterSummaryBucket       `json:"by_adjudication"`
	ByLatestFamily              []deadLetterSummaryBucket       `json:"by_latest_family"`
	ByReasonFamily              []deadLetterSummaryBucket       `json:"by_reason_family"`
	ByActorFamily               []deadLetterSummaryBucket       `json:"by_actor_family"`
	ByDispatchFamily            []deadLetterSummaryBucket       `json:"by_dispatch_family"`
	TopLatestDeadLetterReasons  []deadLetterReasonSummaryItem   `json:"top_latest_dead_letter_reasons"`
	TopLatestManualReplayActors []deadLetterActorSummaryItem    `json:"top_latest_manual_replay_actors"`
	TopLatestDispatchReferences []deadLetterDispatchSummaryItem `json:"top_latest_dispatch_references"`
	RecentDeadLetterTrend       deadLetterTrendWindow           `json:"recent_dead_letter_trend"`
}

type deadLetterTrendBucket struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type deadLetterTrendWindow struct {
	Window        string                  `json:"window,omitempty"`
	Bucket        string                  `json:"bucket,omitempty"`
	WindowedCount int                     `json:"windowed_count,omitempty"`
	Buckets       []deadLetterTrendBucket `json:"buckets,omitempty"`
}

type deadLetterSummaryOptions struct {
	TopN        int
	TrendWindow time.Duration
	TrendBucket time.Duration
	Now         time.Time
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
	cmd.AddCommand(newDeadLetterSummaryCmd(deadLetterLoaderFromBoot(bootLoader)))
	cmd.AddCommand(newDeadLettersCmd(deadLetterLoaderFromBoot(bootLoader)))
	cmd.AddCommand(newDeadLetterCmd(deadLetterLoaderFromBoot(bootLoader)))
	return cmd
}

func newDeadLetterSummaryCmd(loader deadLetterBridgeLoader) *cobra.Command {
	var (
		outputFmt   string
		topN        int
		trendWindow time.Duration
		trendBucket time.Duration
	)

	cmd := &cobra.Command{
		Use:   "dead-letter-summary",
		Short: "Show a dead-letter backlog overview summary",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			bridge, cleanup, err := loader()
			if err != nil {
				return err
			}
			defer cleanup()

			page, err := bridge.List(cmd.Context(), deadLetterListOptions{})
			if err != nil {
				return err
			}

			summary := aggregateDeadLetterSummaryWithOptions(page, deadLetterSummaryOptions{
				TopN:        topN,
				TrendWindow: trendWindow,
				TrendBucket: trendBucket,
			})
			if outputFmt == "json" {
				return printJSONTo(cmd.OutOrStdout(), summary)
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), renderDeadLetterSummaryTable(summary))
			return err
		},
	}
	cmd.Flags().StringVar(&outputFmt, "output", "table", "Output format: table or json")
	cmd.Flags().IntVar(&topN, "top", defaultDeadLetterSummaryTopN, "Top-N size for raw latest reason/actor/dispatch sections")
	cmd.Flags().DurationVar(&trendWindow, "trend-window", defaultDeadLetterTrendWindow, "Time window for recent dead-letter trend output")
	cmd.Flags().DurationVar(&trendBucket, "trend-bucket", defaultDeadLetterTrendBucket, "Bucket size for recent dead-letter trend output")
	return cmd
}

func newDeadLettersCmd(loader deadLetterBridgeLoader) *cobra.Command {
	var (
		outputFmt                 string
		query                     string
		adjudication              string
		latestStatusSubtype       string
		latestStatusSubtypeFamily string
		anyMatchFamily            string
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
			anyMatchFamilyValue, err := normalizeAnyMatchFamily(anyMatchFamily)
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
				AnyMatchFamily:            anyMatchFamilyValue,
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
	cmd.Flags().StringVar(&anyMatchFamily, "any-match-family", "", "Any-match family filter: retry, manual-retry, or dead-letter")
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
		outputFmt    string
		yes          bool
		wait         bool
		waitInterval time.Duration
		waitTimeout  time.Duration
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

			result := deadLetterRetryResult{
				TransactionReceiptID: transactionReceiptID,
				Result:               "accepted",
				Message:              fmt.Sprintf("Retry request accepted for transaction %s.", transactionReceiptID),
			}
			if outputFmt != "json" && wait {
				if _, err := fmt.Fprintf(
					cmd.OutOrStdout(),
					"Polling follow-up status every %s for up to %s...\n",
					waitInterval,
					waitTimeout,
				); err != nil {
					return err
				}
			}
			followUp, pollCount, timedOut, followUpErr := collectDeadLetterRetryFollowUp(
				cmd.Context(),
				bridge,
				transactionReceiptID,
				status,
				wait,
				waitInterval,
				waitTimeout,
			)
			result.FollowUp = followUp
			result.PollCount = pollCount
			result.TimedOut = timedOut
			if followUpErr != nil {
				result.FollowUpError = followUpErr.Error()
			}
			if outputFmt == "json" {
				return printJSONTo(cmd.OutOrStdout(), result)
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), renderDeadLetterRetryResult(result))
			return err
		},
	}
	cmd.Flags().StringVar(&outputFmt, "output", "table", "Output format: table or json")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&wait, "wait", false, "Poll follow-up status after retry request acceptance")
	cmd.Flags().DurationVar(&waitInterval, "wait-interval", defaultDeadLetterRetryWaitInterval, "Polling interval for retry follow-up status")
	cmd.Flags().DurationVar(&waitTimeout, "wait-timeout", defaultDeadLetterRetryWaitTimeout, "Polling timeout for retry follow-up status")
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
	if anyMatchFamily := strings.TrimSpace(opts.AnyMatchFamily); anyMatchFamily != "" {
		params["any_match_family"] = anyMatchFamily
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
	return normalizeDeadLetterFamilyFlag("latest-status-subtype-family", value)
}

func normalizeAnyMatchFamily(value string) (string, error) {
	return normalizeDeadLetterFamilyFlag("any-match-family", value)
}

func normalizeDeadLetterFamilyFlag(name, value string) (string, error) {
	switch strings.TrimSpace(value) {
	case "", "retry", "manual-retry", "dead-letter":
		return strings.TrimSpace(value), nil
	default:
		return "", fmt.Errorf("invalid --%s %q: must be one of retry, manual-retry, dead-letter", name, value)
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

func aggregateDeadLetterSummary(page deadLetterListPage) deadLetterSummaryResult {
	return aggregateDeadLetterSummaryWithOptions(page, deadLetterSummaryOptions{})
}

func aggregateDeadLetterSummaryWithOptions(
	page deadLetterListPage,
	options deadLetterSummaryOptions,
) deadLetterSummaryResult {
	topN := options.TopN
	if topN <= 0 {
		topN = defaultDeadLetterSummaryTopN
	}
	trendWindow := options.TrendWindow
	if trendWindow <= 0 {
		trendWindow = defaultDeadLetterTrendWindow
	}
	trendBucket := options.TrendBucket
	if trendBucket <= 0 {
		trendBucket = defaultDeadLetterTrendBucket
	}
	if trendBucket > trendWindow {
		trendBucket = trendWindow
	}
	now := options.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}

	retryableCount := 0
	adjudicationCounts := map[string]int{}
	latestFamilyCounts := map[string]int{}
	reasonFamilyCounts := map[string]int{}
	actorFamilyCounts := map[string]int{}
	dispatchFamilyCounts := map[string]int{}
	reasonCounts := map[string]int{}
	actorCounts := map[string]int{}
	dispatchCounts := map[string]int{}
	deadLetterTimes := make([]time.Time, 0, len(page.Entries))

	for _, entry := range page.Entries {
		if entry.CanRetry {
			retryableCount++
		}
		adjudicationCounts[summaryBucketLabel(entry.Adjudication)]++
		latestFamilyCounts[summaryBucketLabel(entry.LatestStatusSubtypeFamily)]++
		reasonFamily := postadjudicationstatus.ClassifyDeadLetterReasonFamily(
			entry.LatestDeadLetterReason,
		)
		reasonFamilyCounts[reasonFamily]++
		actorFamily := postadjudicationstatus.ClassifyManualReplayActorFamily(
			entry.LatestManualReplayActor,
		)
		actorFamilyCounts[actorFamily]++
		if reason := strings.TrimSpace(entry.LatestDeadLetterReason); reason != "" {
			reasonCounts[reason]++
		}
		if actor := strings.TrimSpace(entry.LatestManualReplayActor); actor != "" {
			actorCounts[actor]++
		}
		if dispatchReference := strings.TrimSpace(entry.LatestDispatchReference); dispatchReference != "" {
			dispatchFamilyCounts[postadjudicationstatus.ClassifyDispatchReferenceFamily(dispatchReference)]++
			dispatchCounts[dispatchReference]++
		}
		if deadLetteredAt, ok := parseSummaryTimestamp(entry.LatestDeadLetteredAt); ok {
			deadLetterTimes = append(deadLetterTimes, deadLetteredAt)
		}
	}

	return deadLetterSummaryResult{
		TotalDeadLetters:            summaryTotal(page),
		RetryableCount:              retryableCount,
		TopLimit:                    topN,
		ByAdjudication:              orderedSummaryBuckets(adjudicationCounts, []string{"release", "refund"}),
		ByLatestFamily:              orderedSummaryBuckets(latestFamilyCounts, []string{"retry", "manual-retry", "dead-letter"}),
		ByReasonFamily:              orderedSummaryBuckets(reasonFamilyCounts, deadLetterReasonFamilyOrder()),
		ByActorFamily:               orderedSummaryBuckets(actorFamilyCounts, manualReplayActorFamilyOrder()),
		ByDispatchFamily:            orderedSummaryBuckets(dispatchFamilyCounts, postadjudicationstatus.DispatchReferenceFamilyOrder()),
		TopLatestDeadLetterReasons:  topDeadLetterReasons(reasonCounts, topN),
		TopLatestManualReplayActors: topManualReplayActors(actorCounts, topN),
		TopLatestDispatchReferences: topDispatchReferences(dispatchCounts, topN),
		RecentDeadLetterTrend:       summarizeDeadLetterTrend(deadLetterTimes, now, trendWindow, trendBucket),
	}
}

func parseSummaryTimestamp(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func summarizeDeadLetterTrend(
	timestamps []time.Time,
	now time.Time,
	window time.Duration,
	bucket time.Duration,
) deadLetterTrendWindow {
	if window <= 0 || bucket <= 0 {
		return deadLetterTrendWindow{}
	}

	windowStart := now.Add(-window)
	bucketCount := int(window / bucket)
	if window%bucket != 0 {
		bucketCount++
	}
	if bucketCount <= 0 {
		bucketCount = 1
	}

	buckets := make([]deadLetterTrendBucket, 0, bucketCount)
	counts := make([]int, bucketCount)
	windowedCount := 0
	for _, timestamp := range timestamps {
		if timestamp.Before(windowStart) || timestamp.After(now) {
			continue
		}
		windowedCount++
		index := int(timestamp.Sub(windowStart) / bucket)
		if index >= bucketCount {
			index = bucketCount - 1
		}
		counts[index]++
	}

	for index := 0; index < bucketCount; index++ {
		start := windowStart.Add(time.Duration(index) * bucket).UTC()
		end := start.Add(bucket).UTC()
		if end.After(now) {
			end = now
		}
		buckets = append(buckets, deadLetterTrendBucket{
			Label: start.Format(time.RFC3339) + " -> " + end.Format(time.RFC3339),
			Count: counts[index],
		})
	}

	return deadLetterTrendWindow{
		Window:        window.String(),
		Bucket:        bucket.String(),
		WindowedCount: windowedCount,
		Buckets:       buckets,
	}
}

func collectDeadLetterRetryFollowUp(
	ctx context.Context,
	bridge deadLetterBridge,
	transactionReceiptID string,
	baseline postadjudicationstatus.TransactionStatus,
	wait bool,
	waitInterval time.Duration,
	waitTimeout time.Duration,
) (*deadLetterRetryFollowUp, int, bool, error) {
	observe := func() (*deadLetterRetryFollowUp, error) {
		status, err := bridge.Detail(ctx, transactionReceiptID)
		if err != nil {
			return nil, err
		}
		return deadLetterRetryFollowUpFromStatus(status), nil
	}

	followUp, err := observe()
	if err != nil || !wait || deadLetterRetryFollowUpChanged(baseline, followUp) {
		return followUp, boolToPollCount(followUp != nil), false, err
	}

	if waitInterval <= 0 {
		waitInterval = defaultDeadLetterRetryWaitInterval
	}
	if waitTimeout <= 0 {
		waitTimeout = defaultDeadLetterRetryWaitTimeout
	}

	pollCount := 1
	ticker := time.NewTicker(waitInterval)
	defer ticker.Stop()
	timer := time.NewTimer(waitTimeout)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return followUp, pollCount, false, ctx.Err()
		case <-timer.C:
			return followUp, pollCount, true, nil
		case <-ticker.C:
			nextFollowUp, nextErr := observe()
			pollCount++
			if nextErr != nil {
				return followUp, pollCount, false, nextErr
			}
			followUp = nextFollowUp
			if deadLetterRetryFollowUpChanged(baseline, followUp) {
				return followUp, pollCount, false, nil
			}
		}
	}
}

func boolToPollCount(hasFollowUp bool) int {
	if hasFollowUp {
		return 1
	}
	return 0
}

func deadLetterRetryFollowUpFromStatus(
	status postadjudicationstatus.TransactionStatus,
) *deadLetterRetryFollowUp {
	followUp := &deadLetterRetryFollowUp{
		ObservedAt:                time.Now().UTC().Format(time.RFC3339),
		IsDeadLettered:            status.IsDeadLettered,
		CanRetry:                  status.CanRetry,
		LatestStatusSubtype:       status.RetryDeadLetterSummary.LatestStatusSubtype,
		LatestStatusSubtypeFamily: status.RetryDeadLetterSummary.LatestStatusSubtypeFamily,
		LatestDeadLetterReason:    status.RetryDeadLetterSummary.LatestDeadLetterReason,
		LatestRetryAttempt:        status.RetryDeadLetterSummary.LatestRetryAttempt,
		LatestDispatchReference:   status.RetryDeadLetterSummary.LatestDispatchReference,
	}
	if status.LatestBackgroundTask != nil {
		backgroundTask := *status.LatestBackgroundTask
		followUp.BackgroundTask = &backgroundTask
	}
	return followUp
}

func deadLetterRetryFollowUpChanged(
	baseline postadjudicationstatus.TransactionStatus,
	followUp *deadLetterRetryFollowUp,
) bool {
	if followUp == nil {
		return false
	}
	if baseline.IsDeadLettered != followUp.IsDeadLettered || baseline.CanRetry != followUp.CanRetry {
		return true
	}
	if baseline.RetryDeadLetterSummary.LatestStatusSubtype != followUp.LatestStatusSubtype {
		return true
	}
	if baseline.RetryDeadLetterSummary.LatestStatusSubtypeFamily != followUp.LatestStatusSubtypeFamily {
		return true
	}
	if baseline.RetryDeadLetterSummary.LatestRetryAttempt != followUp.LatestRetryAttempt {
		return true
	}
	if baseline.RetryDeadLetterSummary.LatestDispatchReference != followUp.LatestDispatchReference {
		return true
	}
	if backgroundTaskChanged(baseline.LatestBackgroundTask, followUp.BackgroundTask) {
		return true
	}
	return false
}

func backgroundTaskChanged(
	baseline *postadjudicationstatus.BackgroundTaskBridge,
	followUp *postadjudicationstatus.BackgroundTaskBridge,
) bool {
	if baseline == nil || followUp == nil {
		return baseline != followUp
	}
	return baseline.TaskID != followUp.TaskID ||
		baseline.Status != followUp.Status ||
		baseline.AttemptCount != followUp.AttemptCount ||
		baseline.NextRetryAt != followUp.NextRetryAt
}

func deadLetterReasonFamilyOrder() []string {
	return []string{
		postadjudicationstatus.DeadLetterReasonFamilyRetryExhausted,
		postadjudicationstatus.DeadLetterReasonFamilyPolicyBlocked,
		postadjudicationstatus.DeadLetterReasonFamilyReceiptInvalid,
		postadjudicationstatus.DeadLetterReasonFamilyBackgroundFailed,
		postadjudicationstatus.DeadLetterReasonFamilyUnknown,
	}
}

func manualReplayActorFamilyOrder() []string {
	return []string{
		postadjudicationstatus.ManualReplayActorFamilyOperator,
		postadjudicationstatus.ManualReplayActorFamilySystem,
		postadjudicationstatus.ManualReplayActorFamilyService,
		postadjudicationstatus.ManualReplayActorFamilyUnknown,
	}
}

func summaryTotal(page deadLetterListPage) int {
	if page.Total > 0 {
		return page.Total
	}
	return len(page.Entries)
}

func summaryBucketLabel(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "unknown"
	}
	return trimmed
}

func orderedSummaryBuckets(counts map[string]int, preferredOrder []string) []deadLetterSummaryBucket {
	buckets := make([]deadLetterSummaryBucket, 0, len(counts))
	used := map[string]struct{}{}
	for _, label := range preferredOrder {
		if count, ok := counts[label]; ok {
			buckets = append(buckets, deadLetterSummaryBucket{Label: label, Count: count})
			used[label] = struct{}{}
		}
	}

	extras := make([]string, 0, len(counts))
	for label := range counts {
		if _, ok := used[label]; ok {
			continue
		}
		extras = append(extras, label)
	}
	sort.Strings(extras)
	for _, label := range extras {
		buckets = append(buckets, deadLetterSummaryBucket{Label: label, Count: counts[label]})
	}
	return buckets
}

func topDeadLetterReasons(counts map[string]int, limit int) []deadLetterReasonSummaryItem {
	if limit <= 0 || len(counts) == 0 {
		return nil
	}

	items := make([]deadLetterReasonSummaryItem, 0, len(counts))
	for reason, count := range counts {
		items = append(items, deadLetterReasonSummaryItem{Reason: reason, Count: count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count != items[j].Count {
			return items[i].Count > items[j].Count
		}
		return items[i].Reason < items[j].Reason
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}

func topManualReplayActors(counts map[string]int, limit int) []deadLetterActorSummaryItem {
	if limit <= 0 || len(counts) == 0 {
		return nil
	}

	items := make([]deadLetterActorSummaryItem, 0, len(counts))
	for actor, count := range counts {
		items = append(items, deadLetterActorSummaryItem{Actor: actor, Count: count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count != items[j].Count {
			return items[i].Count > items[j].Count
		}
		return items[i].Actor < items[j].Actor
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}

func topDispatchReferences(counts map[string]int, limit int) []deadLetterDispatchSummaryItem {
	if limit <= 0 || len(counts) == 0 {
		return nil
	}

	items := make([]deadLetterDispatchSummaryItem, 0, len(counts))
	for dispatchReference, count := range counts {
		items = append(items, deadLetterDispatchSummaryItem{
			DispatchReference: dispatchReference,
			Count:             count,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count != items[j].Count {
			return items[i].Count > items[j].Count
		}
		return items[i].DispatchReference < items[j].DispatchReference
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items
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
	ctx = ctxkeys.WithDefaultPrincipal(ctx, "system:cli")
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
