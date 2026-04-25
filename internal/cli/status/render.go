package status

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/postadjudicationstatus"
)

func renderDashboard(info StatusInfo) string {
	var b strings.Builder

	// Title
	version := info.Version
	if version == "" {
		version = "dev"
	}
	title := lipgloss.NewStyle().Bold(true).Foreground(tui.Primary).Render(
		fmt.Sprintf("Lango Status                              v%s (profile: %s)", version, info.Profile),
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
	b.WriteString(infoLine("Gateway", lipgloss.NewStyle().Foreground(tui.Muted).Render(info.Gateway)))
	providerInfo := info.Provider
	if info.Model != "" {
		providerInfo += " (" + info.Model + ")"
	}
	b.WriteString(infoLine("Provider", lipgloss.NewStyle().Foreground(tui.Muted).Render(providerInfo)))
	if info.ContextProfile != "" {
		b.WriteString(infoLine("Ctx Profile", lipgloss.NewStyle().Foreground(tui.Muted).Render(info.ContextProfile)))
	}
	b.WriteString("\n")

	// Channels
	if len(info.Channels) > 0 {
		b.WriteString(sectionHeader("Channels"))
		b.WriteString(infoLine("Active", lipgloss.NewStyle().Foreground(tui.Success).Render(strings.Join(info.Channels, ", "))))
		b.WriteString("\n")
	}

	// Features
	b.WriteString(sectionHeader("Features"))
	var enabled []string
	var disabled []string
	for _, f := range info.Features {
		if f.Enabled {
			label := f.Name
			if f.Detail != "" {
				label += " (" + f.Detail + ")"
			}
			enabled = append(enabled, label)
		} else {
			disabled = append(disabled, f.Name)
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

func renderDeadLetterBacklogTable(page deadLetterListPage) string {
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
	b.WriteString(fmt.Sprintf("%-20s %-24s %-12s %-8s %-8s\n", "Transaction", "Reason", "Adjudication", "Attempt", "Retry"))
	b.WriteString(sep)
	b.WriteString("\n")
	for _, entry := range page.Entries {
		b.WriteString(fmt.Sprintf(
			"%-20s %-24s %-12s %-8d %-8t\n",
			tui.Truncate(entry.TransactionReceiptID, 20),
			tui.Truncate(entry.LatestDeadLetterReason, 24),
			tui.Truncate(entry.Adjudication, 12),
			entry.LatestRetryAttempt,
			entry.CanRetry,
		))
	}
	if page.Total > 0 {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Count: %d  Total: %d  Offset: %d  Limit: %d\n", page.Count, page.Total, page.Offset, page.Limit))
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
	b.WriteString(infoLine("Transaction", status.CanonicalSnapshot.TransactionReceipt.TransactionReceiptID))
	b.WriteString(infoLine("Submission", status.CanonicalSnapshot.SubmissionReceipt.SubmissionReceiptID))
	b.WriteString(infoLine("Adjudication", status.Adjudication))
	b.WriteString(infoLine("Dead-lettered", fmt.Sprintf("%t", status.IsDeadLettered)))
	b.WriteString(infoLine("Retryable", fmt.Sprintf("%t", status.CanRetry)))
	b.WriteString(infoLine("Latest Reason", fallbackText(status.RetryDeadLetterSummary.LatestDeadLetterReason)))
	b.WriteString(infoLine("Retry Attempt", fmt.Sprintf("%d", status.RetryDeadLetterSummary.LatestRetryAttempt)))
	b.WriteString(infoLine("Dispatch Ref", fallbackText(status.RetryDeadLetterSummary.LatestDispatchReference)))
	if task := status.LatestBackgroundTask; task != nil {
		b.WriteString(infoLine("Task ID", task.TaskID))
		b.WriteString(infoLine("Task Status", task.Status))
		b.WriteString(infoLine("Task Attempts", fmt.Sprintf("%d", task.AttemptCount)))
		b.WriteString(infoLine("Next Retry", fallbackText(task.NextRetryAt)))
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
	b.WriteString(sectionHeader("Top Latest Dead-Letter Reasons"))
	b.WriteString(renderReasonSummaryItems(summary.TopLatestDeadLetterReasons))
	return b.String()
}

func renderSummaryBuckets(buckets []deadLetterSummaryBucket) string {
	if len(buckets) == 0 {
		return infoLine("none", "0")
	}

	var b strings.Builder
	for _, bucket := range buckets {
		b.WriteString(infoLine(bucket.Label, fmt.Sprintf("%d", bucket.Count)))
	}
	return b.String()
}

func renderReasonSummaryItems(items []deadLetterReasonSummaryItem) string {
	if len(items) == 0 {
		return infoLine("none", "0")
	}

	var b strings.Builder
	sep := lipgloss.NewStyle().Foreground(tui.Separator).Render(strings.Repeat("\u2500", 72))
	b.WriteString(fmt.Sprintf("%-60s %-8s\n", "Reason", "Count"))
	b.WriteString(sep)
	b.WriteString("\n")
	for _, item := range items {
		b.WriteString(fmt.Sprintf("%-60s %-8d\n", tui.Truncate(item.Reason, 60), item.Count))
	}
	return b.String()
}

func fallbackText(value string) string {
	if strings.TrimSpace(value) == "" {
		return "n/a"
	}
	return value
}
