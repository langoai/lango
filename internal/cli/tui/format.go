package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Truncate shortens s to maxLen, appending "..." if needed.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// WordWrap wraps text to the given width, breaking on spaces.
func WordWrap(text string, width int) string {
	if width <= 0 || text == "" {
		return text
	}
	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		var cur strings.Builder
		for _, w := range words {
			if cur.Len() == 0 {
				cur.WriteString(w)
			} else if cur.Len()+1+len(w) > width {
				lines = append(lines, cur.String())
				cur.Reset()
				cur.WriteString(w)
			} else {
				cur.WriteByte(' ')
				cur.WriteString(w)
			}
		}
		if cur.Len() > 0 {
			lines = append(lines, cur.String())
		}
	}
	return strings.Join(lines, "\n")
}

// FormatNumber renders an integer with comma-separated thousands
// (e.g., 12345 -> "12,345").
func FormatNumber(n int64) string {
	if n < 0 {
		return "-" + FormatNumber(-n)
	}
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var buf strings.Builder
	remainder := len(s) % 3
	if remainder > 0 {
		buf.WriteString(s[:remainder])
	}
	for i := remainder; i < len(s); i += 3 {
		if buf.Len() > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(s[i : i+3])
	}
	return buf.String()
}

// FormatTokens returns a human-readable token count with comma separators.
func FormatTokens(n int) string {
	return FormatNumber(int64(n))
}

// FormatDuration renders a duration as a human-friendly string
// (e.g., "2h 15m", "3m 42s", "150ms").
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	totalSec := int(d.Seconds())
	h := totalSec / 3600
	m := (totalSec % 3600) / 60
	s := totalSec % 60

	switch {
	case h > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	default:
		return fmt.Sprintf("%dm %ds", m, s)
	}
}

// RelativeTime formats a timestamp as a human-readable relative duration.
// It returns precise values for sub-minute durations (e.g., "5s ago").
func RelativeTime(now, t time.Time) string {
	d := now.Sub(t)
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

// RelativeTimeHuman formats a timestamp as a friendly relative duration.
// For sub-minute durations it returns "just now" instead of precise seconds.
func RelativeTimeHuman(now, t time.Time) string {
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		return fmt.Sprintf("%dh ago", h)
	default:
		days := int(d.Hours()) / 24
		return fmt.Sprintf("%dd ago", days)
	}
}
