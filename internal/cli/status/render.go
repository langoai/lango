package status

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/postadjudicationstatus"
)

func sanitizeStatusText(text string) string {
	return strings.Join(strings.Fields(ansi.Strip(text)), " ")
}

func renderDashboard(info StatusInfo) string {
	var b strings.Builder

	// Title
	version := info.Version
	if version == "" {
		version = "dev"
	}
	title := lipgloss.NewStyle().Bold(true).Foreground(tui.Primary).Render(
		fmt.Sprintf("Lango Status                              v%s (profile: %s)", sanitizeStatusText(version), sanitizeStatusText(info.Profile)),
	)
	b.WriteString("\n")
	b.WriteString(title)
	b.WriteString("\n")
	sep := lipgloss.NewStyle().Foreground(tui.Separator).Render(strings.Repeat("\u2500", 60))
	b.WriteString(sep)
	b.WriteString("\n\n")

	// System section
	b.WriteString(sectionHeader("System"))
	if info.ServerUp {
		b.WriteString(infoLine("Server", tui.FormatPass("running")))
	} else {
		b.WriteString(infoLine("Server", tui.FormatFail("not running")))
	}
	b.WriteString(infoLine("Gateway", lipgloss.NewStyle().Foreground(tui.Muted).Render(sanitizeStatusText(info.Gateway))))
	providerInfo := sanitizeStatusText(info.Provider)
	if info.Model != "" {
		providerInfo += " (" + sanitizeStatusText(info.Model) + ")"
	}
	b.WriteString(infoLine("Provider", lipgloss.NewStyle().Foreground(tui.Muted).Render(providerInfo)))
	if info.ContextProfile != "" {
		b.WriteString(infoLine("Ctx Profile", lipgloss.NewStyle().Foreground(tui.Muted).Render(sanitizeStatusText(info.ContextProfile))))
	}
	b.WriteString("\n")

	// Channels
	if len(info.Channels) > 0 {
		b.WriteString(sectionHeader("Channels"))
		activeChannels := make([]string, 0, len(info.Channels))
		for _, ch := range info.Channels {
			activeChannels = append(activeChannels, sanitizeStatusText(ch))
		}
		b.WriteString(infoLine("Active", lipgloss.NewStyle().Foreground(tui.Success).Render(strings.Join(activeChannels, ", "))))
		b.WriteString("\n")
	}

	// Features
	b.WriteString(sectionHeader("Features"))
	var enabled []string
	var disabled []string
	for _, f := range info.Features {
		if f.Enabled {
			label := sanitizeStatusText(f.Name)
			if f.Detail != "" {
				label += " (" + sanitizeStatusText(f.Detail) + ")"
			}
			enabled = append(enabled, label)
		} else {
			disabled = append(disabled, sanitizeStatusText(f.Name))
		}
	}

	// Show enabled features.
	for _, name := range enabled {
		b.WriteString("    ")
		b.WriteString(tui.FormatPass(name))
		b.WriteString("\n")
	}

	// Show disabled summary.
	if len(disabled) > 0 {
		b.WriteString("    ")
		b.WriteString(tui.FormatFail("Disabled: " + strings.Join(disabled, ", ")))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	return b.String()
}

func sectionHeader(title string) string {
	return "  " + lipgloss.NewStyle().Bold(true).Foreground(tui.Highlight).Render(title) + "\n"
}

func infoLine(label, value string) string {
	labelStyle := lipgloss.NewStyle().Width(16).PaddingLeft(4)
	return labelStyle.Render(label) + value + "\n"
}

func renderDeadLetterBacklogTable(page DeadLetterListPage) string {
	if len(page.Entries) == 0 {
		return "No current dead-letter backlog.\n"
	}

	var b strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(tui.Primary).Render("Dead-Letter Backlog")
	b.WriteString(title)
	b.WriteString("\n")
	sep := lipgloss.NewStyle().Foreground(tui.Separator).Render(strings.Repeat("\u2500", 72))
	b.WriteString(sep)
	b.WriteString("\n")
	fmt.Fprintf(&b, "%-20s %-24s %-12s %-8s %-8s\n", "Transaction", "Reason", "Adjudication", "Attempt", "Retry")
	b.WriteString(sep)
	b.WriteString("\n")
	for _, entry := range page.Entries {
		fmt.Fprintf(&b,
			"%-20s %-24s %-12s %-8d %-8t\n",
			tui.Truncate(sanitizeStatusText(entry.TransactionReceiptID), 20),
			tui.Truncate(sanitizeStatusText(entry.LatestDeadLetterReason), 24),
			tui.Truncate(sanitizeStatusText(entry.Adjudication), 12),
			entry.LatestRetryAttempt,
			entry.CanRetry,
		)
	}
	if page.Total > 0 {
		b.WriteString("\n")
		fmt.Fprintf(&b, "Count: %d  Total: %d  Offset: %d  Limit: %d\n", page.Count, page.Total, page.Offset, page.Limit)
	}
	return b.String()
}

func renderDeadLetterDetail(status postadjudicationstatus.TransactionStatus) string {
	var b strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(tui.Primary).Render("Dead-Letter Detail")
	b.WriteString(title)
	b.WriteString("\n")
	sep := lipgloss.NewStyle().Foreground(tui.Separator).Render(strings.Repeat("\u2500", 72))
	b.WriteString(sep)
	b.WriteString("\n")
	b.WriteString(infoLine("Transaction", sanitizeStatusText(status.CanonicalSnapshot.TransactionReceipt.TransactionReceiptID)))
	b.WriteString(infoLine("Submission", sanitizeStatusText(status.CanonicalSnapshot.SubmissionReceipt.SubmissionReceiptID)))
	b.WriteString(infoLine("Adjudication", sanitizeStatusText(status.Adjudication)))
	b.WriteString(infoLine("Dead-lettered", fmt.Sprintf("%t", status.IsDeadLettered)))
	b.WriteString(infoLine("Retryable", fmt.Sprintf("%t", status.CanRetry)))
	b.WriteString(infoLine("Latest Reason", sanitizeStatusText(fallbackText(status.RetryDeadLetterSummary.LatestDeadLetterReason))))
	b.WriteString(infoLine("Retry Attempt", fmt.Sprintf("%d", status.RetryDeadLetterSummary.LatestRetryAttempt)))
	b.WriteString(infoLine("Dispatch Ref", sanitizeStatusText(fallbackText(status.RetryDeadLetterSummary.LatestDispatchReference))))
	if task := status.LatestBackgroundTask; task != nil {
		b.WriteString(infoLine("Task ID", sanitizeStatusText(task.TaskID)))
		b.WriteString(infoLine("Task Status", sanitizeStatusText(task.Status)))
		b.WriteString(infoLine("Task Attempts", fmt.Sprintf("%d", task.AttemptCount)))
		b.WriteString(infoLine("Next Retry", sanitizeStatusText(fallbackText(task.NextRetryAt))))
	} else {
		b.WriteString(infoLine("Task ID", "n/a"))
	}
	return b.String()
}

func renderDeadLetterSummaryTable(summary deadLetterSummaryResult) string {
	var b strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(tui.Primary).Render("Dead-Letter Summary")
	b.WriteString(title)
	b.WriteString("\n")
	sep := lipgloss.NewStyle().Foreground(tui.Separator).Render(strings.Repeat("\u2500", 72))
	b.WriteString(sep)
	b.WriteString("\n")
	b.WriteString(infoLine("Total", fmt.Sprintf("%d", summary.TotalDeadLetters)))
	b.WriteString(infoLine("Retryable", fmt.Sprintf("%d", summary.RetryableCount)))

	b.WriteString("\n")
	b.WriteString(sectionHeader("By Adjudication"))
	b.WriteString(renderSummaryBuckets(summary.ByAdjudication))

	b.WriteString("\n")
	b.WriteString(sectionHeader("By Latest Family"))
	b.WriteString(renderSummaryBuckets(summary.ByLatestFamily))

	b.WriteString("\n")
	b.WriteString(sectionHeader("By reason family"))
	b.WriteString(renderSummaryBuckets(summary.ByReasonFamily))

	b.WriteString("\n")
	b.WriteString(sectionHeader("By actor family"))
	b.WriteString(renderSummaryBuckets(summary.ByActorFamily))

	b.WriteString("\n")
	b.WriteString(sectionHeader("By dispatch family"))
	b.WriteString(renderSummaryBuckets(summary.ByDispatchFamily))

	b.WriteString("\n")
	b.WriteString(sectionHeader("Recent dead-letter trend"))
	b.WriteString(renderDeadLetterTrend(summary.RecentDeadLetterTrend))

	b.WriteString("\n")
	b.WriteString(sectionHeader(fmt.Sprintf("Top Latest Dead-Letter Reasons (Top %d)", summary.TopLimit)))
	b.WriteString(renderReasonSummaryItems(summary.TopLatestDeadLetterReasons))

	b.WriteString("\n")
	b.WriteString(sectionHeader(fmt.Sprintf("Top Latest Manual Replay Actors (Top %d)", summary.TopLimit)))
	b.WriteString(renderActorSummaryItems(summary.TopLatestManualReplayActors))

	b.WriteString("\n")
	b.WriteString(sectionHeader(fmt.Sprintf("Top Latest Dispatch References (Top %d)", summary.TopLimit)))
	b.WriteString(renderDispatchSummaryItems(summary.TopLatestDispatchReferences))
	return b.String()
}

func renderSummaryBuckets(buckets []deadLetterSummaryBucket) string {
	if len(buckets) == 0 {
		return infoLine("none", "0")
	}

	var b strings.Builder
	for _, bucket := range buckets {
		fmt.Fprintf(&b, "    %-24s%d\n", sanitizeStatusText(bucket.Label), bucket.Count)
	}
	return b.String()
}

func renderReasonSummaryItems(items []deadLetterReasonSummaryItem) string {
	if len(items) == 0 {
		return infoLine("none", "0")
	}

	var b strings.Builder
	sep := lipgloss.NewStyle().Foreground(tui.Separator).Render(strings.Repeat("\u2500", 72))
	fmt.Fprintf(&b, "%-60s %-8s\n", "Reason", "Count")
	b.WriteString(sep)
	b.WriteString("\n")
	for _, item := range items {
		fmt.Fprintf(&b, "%-60s %-8d\n", tui.Truncate(sanitizeStatusText(item.Reason), 60), item.Count)
	}
	return b.String()
}

func renderActorSummaryItems(items []deadLetterActorSummaryItem) string {
	if len(items) == 0 {
		return infoLine("none", "0")
	}

	var b strings.Builder
	sep := lipgloss.NewStyle().Foreground(tui.Separator).Render(strings.Repeat("\u2500", 72))
	fmt.Fprintf(&b, "%-60s %-8s\n", "Actor", "Count")
	b.WriteString(sep)
	b.WriteString("\n")
	for _, item := range items {
		fmt.Fprintf(&b, "%-60s %-8d\n", tui.Truncate(sanitizeStatusText(item.Actor), 60), item.Count)
	}
	return b.String()
}

func renderDispatchSummaryItems(items []deadLetterDispatchSummaryItem) string {
	if len(items) == 0 {
		return infoLine("none", "0")
	}

	var b strings.Builder
	sep := lipgloss.NewStyle().Foreground(tui.Separator).Render(strings.Repeat("\u2500", 72))
	fmt.Fprintf(&b, "%-60s %-8s\n", "Dispatch Reference", "Count")
	b.WriteString(sep)
	b.WriteString("\n")
	for _, item := range items {
		fmt.Fprintf(&b, "%-60s %-8d\n", tui.Truncate(sanitizeStatusText(item.DispatchReference), 60), item.Count)
	}
	return b.String()
}

func renderDeadLetterTrend(trend deadLetterTrendWindow) string {
	if trend.Window == "" || trend.Bucket == "" {
		return infoLine("none", "0")
	}

	var b strings.Builder
	b.WriteString(infoLine("Window", trend.Window))
	b.WriteString(infoLine("Bucket", trend.Bucket))
	b.WriteString(infoLine("Windowed Count", fmt.Sprintf("%d", trend.WindowedCount)))
	if len(trend.Buckets) == 0 {
		return b.String()
	}
	for _, bucket := range trend.Buckets {
		fmt.Fprintf(&b, "    %-40s%d\n", tui.Truncate(sanitizeStatusText(bucket.Label), 40), bucket.Count)
	}
	return b.String()
}

func renderDeadLetterRetryResult(result deadLetterRetryResult) string {
	var b strings.Builder
	b.WriteString(sanitizeStatusText(result.Message))
	b.WriteString("\n")

	if result.PollCount > 0 {
		b.WriteString(infoLine("Follow-up Polls", fmt.Sprintf("%d", result.PollCount)))
	}
	if result.TimedOut {
		b.WriteString(infoLine("Wait Timed Out", "true"))
	}
	if result.FollowUpError != "" {
		b.WriteString(infoLine("Follow-up Error", sanitizeStatusText(result.FollowUpError)))
	}
	if result.FollowUp == nil {
		return b.String()
	}

	b.WriteString("\n")
	b.WriteString(sectionHeader("Follow-up"))
	b.WriteString(infoLine("Observed At", fallbackText(result.FollowUp.ObservedAt)))
	b.WriteString(infoLine("Dead-lettered", fmt.Sprintf("%t", result.FollowUp.IsDeadLettered)))
	b.WriteString(infoLine("Retryable", fmt.Sprintf("%t", result.FollowUp.CanRetry)))
	b.WriteString(infoLine("Latest Status", sanitizeStatusText(fallbackText(result.FollowUp.LatestStatusSubtype))))
	b.WriteString(infoLine("Latest Family", sanitizeStatusText(fallbackText(result.FollowUp.LatestStatusSubtypeFamily))))
	b.WriteString(infoLine("Latest Reason", sanitizeStatusText(fallbackText(result.FollowUp.LatestDeadLetterReason))))
	b.WriteString(infoLine("Retry Attempt", fmt.Sprintf("%d", result.FollowUp.LatestRetryAttempt)))
	b.WriteString(infoLine("Dispatch Ref", fallbackText(result.FollowUp.LatestDispatchReference)))
	if result.FollowUp.BackgroundTask == nil {
		b.WriteString(infoLine("Task ID", "n/a"))
		return b.String()
	}

	b.WriteString(infoLine("Task ID", result.FollowUp.BackgroundTask.TaskID))
	b.WriteString(infoLine("Task Status", sanitizeStatusText(result.FollowUp.BackgroundTask.Status)))
	b.WriteString(infoLine("Task Attempts", fmt.Sprintf("%d", result.FollowUp.BackgroundTask.AttemptCount)))
	b.WriteString(infoLine("Next Retry", fallbackText(result.FollowUp.BackgroundTask.NextRetryAt)))
	return b.String()
}

func fallbackText(value string) string {
	if strings.TrimSpace(value) == "" {
		return "n/a"
	}
	return value
}
