package output

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/langoai/lango/internal/cli/doctor/checks"
)

func sanitizeDoctorText(text string) string {
	return strings.Join(strings.Fields(ansi.Strip(text)), " ")
}

func sanitizeTraceFailures(items []checks.TraceFailure) []checks.TraceFailure {
	if len(items) == 0 {
		return items
	}
	out := make([]checks.TraceFailure, len(items))
	for i, item := range items {
		item.TraceID = sanitizeDoctorText(item.TraceID)
		item.Outcome = sanitizeDoctorText(item.Outcome)
		item.ErrorCode = sanitizeDoctorText(item.ErrorCode)
		item.CauseClass = sanitizeDoctorText(item.CauseClass)
		item.Summary = sanitizeDoctorText(item.Summary)
		out[i] = item
	}
	return out
}
