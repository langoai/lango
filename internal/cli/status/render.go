package status

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
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
