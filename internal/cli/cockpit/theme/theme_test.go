package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		wantColor lipgloss.Color
	}{
		// Surface colors
		{give: "Surface0", wantColor: Surface0},
		{give: "Surface1", wantColor: Surface1},
		{give: "Surface2", wantColor: Surface2},
		{give: "Surface3", wantColor: Surface3},
		// Text colors
		{give: "TextPrimary", wantColor: TextPrimary},
		{give: "TextSecondary", wantColor: TextSecondary},
		{give: "TextTertiary", wantColor: TextTertiary},
		// Border colors
		{give: "BorderFocused", wantColor: BorderFocused},
		{give: "BorderDefault", wantColor: BorderDefault},
		{give: "BorderSubtle", wantColor: BorderSubtle},
		// Brand colors
		{give: "Primary", wantColor: Primary},
		{give: "Success", wantColor: Success},
		{give: "Warning", wantColor: Warning},
		{give: "Error", wantColor: Error},
		{give: "Accent", wantColor: Accent},
		{give: "Muted", wantColor: Muted},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			if string(tt.wantColor) == "" {
				t.Errorf("%s: color is empty", tt.give)
			}
		})
	}
}

func TestIcons(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give     string
		wantIcon string
	}{
		{give: "IconChat", wantIcon: IconChat},
		{give: "IconSettings", wantIcon: IconSettings},
		{give: "IconTools", wantIcon: IconTools},
		{give: "IconStatus", wantIcon: IconStatus},
		{give: "IconSessions", wantIcon: IconSessions},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			if tt.wantIcon == "" {
				t.Errorf("%s: icon is empty", tt.give)
			}
		})
	}
}

func TestIndicators(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give          string
		wantIndicator string
	}{
		{give: "IndicatorActive", wantIndicator: IndicatorActive},
		{give: "IndicatorInactive", wantIndicator: IndicatorInactive},
		{give: "IndicatorPass", wantIndicator: IndicatorPass},
		{give: "IndicatorFail", wantIndicator: IndicatorFail},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			if tt.wantIndicator == "" {
				t.Errorf("%s: indicator is empty", tt.give)
			}
		})
	}
}

func TestRenderLogo(t *testing.T) {
	t.Parallel()

	got := RenderLogo()
	if got == "" {
		t.Error("RenderLogo returned empty string")
	}
}

func TestSidebarWidthConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		wantWidth int
	}{
		{give: "SidebarFullWidth", wantWidth: SidebarFullWidth},
		{give: "SidebarCollapsedWidth", wantWidth: SidebarCollapsedWidth},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			if tt.wantWidth <= 0 {
				t.Errorf("%s: width must be positive, got %d", tt.give, tt.wantWidth)
			}
		})
	}
}
