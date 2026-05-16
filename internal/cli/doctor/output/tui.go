// Package output provides output renderers for the doctor command.
package output

import (
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/cli/doctor/checks"
	"github.com/langoai/lango/internal/cli/tui"
)

// TUIRenderer renders check results to the terminal with styling.
type TUIRenderer struct{}

// RenderResult renders a single check result.
func (r *TUIRenderer) RenderResult(result checks.Result) string {
	name := sanitizeDoctorText(result.Name)
	message := sanitizeDoctorText(result.Message)
	details := sanitizeDoctorText(result.Details)

	var indicator string
	switch result.Status {
	case checks.StatusPass:
		indicator = tui.FormatPass(name)
	case checks.StatusWarn:
		indicator = tui.FormatWarn(name)
	case checks.StatusFail:
		indicator = tui.FormatFail(name)
	case checks.StatusSkip:
		indicator = tui.FormatMuted(name)
	}

	var sb strings.Builder
	sb.WriteString(indicator)
	sb.WriteString("\n")
	sb.WriteString("  ")
	sb.WriteString(message)
	sb.WriteString("\n")

	if details != "" {
		sb.WriteString("  ")
		sb.WriteString(tui.FormatMuted("→ " + details))
		sb.WriteString("\n")
	}

	return sb.String()
}

// RenderSummary renders the final summary.
func (r *TUIRenderer) RenderSummary(summary checks.Summary) string {
	var parts []string

	if summary.Passed > 0 {
		parts = append(parts, tui.SuccessStyle.Render(fmt.Sprintf("%d passed", summary.Passed)))
	}
	if summary.Warnings > 0 {
		parts = append(parts, tui.WarningStyle.Render(fmt.Sprintf("%d warnings", summary.Warnings)))
	}
	if summary.Failed > 0 {
		parts = append(parts, tui.ErrorStyle.Render(fmt.Sprintf("%d errors", summary.Failed)))
	}
	if summary.Skipped > 0 {
		parts = append(parts, tui.MutedStyle.Render(fmt.Sprintf("%d skipped", summary.Skipped)))
	}

	return fmt.Sprintf("\n%s %s\n",
		tui.SubtitleStyle.Render("Summary:"),
		strings.Join(parts, ", "))
}

// RenderTitle renders the doctor title.
func (r *TUIRenderer) RenderTitle() string {
	return tui.TitleStyle.Render("🏥 Lango Doctor") + "\n"
}

// RenderFixHint renders a hint about --fix flag.
func (r *TUIRenderer) RenderFixHint(hasFixable bool) string {
	if hasFixable {
		return tui.FormatMuted("\nRun 'lango doctor --fix' to attempt automatic repairs.\n")
	}
	return ""
}
