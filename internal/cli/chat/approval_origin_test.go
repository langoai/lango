package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/approval"
)

// --- formatChannelOrigin tests ---

func TestFormatChannelOrigin_Telegram(t *testing.T) {
	got := formatChannelOrigin("telegram:123:456")
	assert.Equal(t, "[Telegram] 123", got)
}

func TestFormatChannelOrigin_Discord(t *testing.T) {
	got := formatChannelOrigin("discord:ch123:user456")
	assert.Equal(t, "[Discord] ch123", got)
}

func TestFormatChannelOrigin_Slack(t *testing.T) {
	got := formatChannelOrigin("slack:C123:U456")
	assert.Equal(t, "[Slack] C123", got)
}

func TestFormatChannelOrigin_NonChannel(t *testing.T) {
	got := formatChannelOrigin("tui-12345")
	assert.Equal(t, "", got)
}

func TestFormatChannelOrigin_Empty(t *testing.T) {
	got := formatChannelOrigin("")
	assert.Equal(t, "", got)
}

func TestFormatChannelOrigin_TwoParts(t *testing.T) {
	got := formatChannelOrigin("telegram:123")
	assert.Equal(t, "", got)
}

// --- formatChannelBadge tests ---

func TestFormatChannelBadge_Telegram(t *testing.T) {
	got := formatChannelBadge("telegram:123:456")
	assert.Equal(t, "[TG]", got)
}

func TestFormatChannelBadge_Discord(t *testing.T) {
	got := formatChannelBadge("discord:ch:user")
	assert.Equal(t, "[DC]", got)
}

func TestFormatChannelBadge_Slack(t *testing.T) {
	got := formatChannelBadge("slack:ch:user")
	assert.Equal(t, "[SL]", got)
}

func TestFormatChannelBadge_NonChannel(t *testing.T) {
	got := formatChannelBadge("tui-12345")
	assert.Equal(t, "", got)
}

// --- Renderer tests with channel session keys ---

func TestRenderApprovalBanner_WithChannelOrigin(t *testing.T) {
	req := approval.ApprovalRequest{
		ToolName:   "fs_read",
		Summary:    "Read config file",
		SessionKey: "telegram:123:456",
	}
	output := renderApprovalBanner(req, 80)

	assert.Contains(t, output, "[Telegram]", "banner should show channel origin for telegram session key")
	assert.Contains(t, output, "123", "banner should show chat ID from session key")
}

func TestRenderApprovalBanner_NoChannelOrigin(t *testing.T) {
	req := approval.ApprovalRequest{
		ToolName:   "fs_read",
		Summary:    "Read config file",
		SessionKey: "tui-12345",
	}
	output := renderApprovalBanner(req, 80)

	assert.NotContains(t, output, "[Telegram]", "banner should not show Telegram for non-channel key")
	assert.NotContains(t, output, "[Discord]", "banner should not show Discord for non-channel key")
	assert.NotContains(t, output, "[Slack]", "banner should not show Slack for non-channel key")
	assert.NotContains(t, output, "←", "banner should not show origin arrow for non-channel key")
}

func TestRenderApprovalStrip_WithChannelBadge(t *testing.T) {
	vm := approval.ApprovalViewModel{
		Request: approval.ApprovalRequest{
			ToolName:   "fs_read",
			Summary:    "Read config file",
			SessionKey: "discord:ch123:user456",
		},
		Risk: approval.RiskIndicator{Level: "moderate", Label: "Reads file"},
	}
	output := renderApprovalStrip(vm, 120)

	assert.Contains(t, output, "[DC]", "strip should show channel badge for discord session key")
}

func TestRenderApprovalStrip_NoChannelBadge(t *testing.T) {
	vm := approval.ApprovalViewModel{
		Request: approval.ApprovalRequest{
			ToolName:   "fs_read",
			Summary:    "Read config file",
			SessionKey: "tui-12345",
		},
		Risk: approval.RiskIndicator{Level: "moderate", Label: "Reads file"},
	}
	output := renderApprovalStrip(vm, 120)

	assert.NotContains(t, output, "[TG]", "strip should not show TG badge for non-channel key")
	assert.NotContains(t, output, "[DC]", "strip should not show DC badge for non-channel key")
	assert.NotContains(t, output, "[SL]", "strip should not show SL badge for non-channel key")
}

func TestRenderApprovalDialog_WithChannelOrigin(t *testing.T) {
	vm := approval.ApprovalViewModel{
		Request: approval.ApprovalRequest{
			ToolName:   "fs_edit",
			Summary:    "Edit config.yaml",
			SessionKey: "slack:C123:U456",
			Params:     map[string]interface{}{"path": "/etc/config.yaml"},
		},
		Risk: approval.RiskIndicator{Level: "high", Label: "Modifies filesystem"},
	}
	output := renderApprovalDialog(vm, 80, 40, 0, false)

	assert.Contains(t, output, "[Slack]", "dialog should show channel origin for slack session key")
	assert.Contains(t, output, "C123", "dialog should show chat ID from session key")
	assert.Contains(t, output, "←", "dialog should show origin arrow")
}

func TestRenderApprovalDialog_NoChannelOrigin(t *testing.T) {
	vm := approval.ApprovalViewModel{
		Request: approval.ApprovalRequest{
			ToolName:   "fs_edit",
			Summary:    "Edit config.yaml",
			SessionKey: "tui-12345",
			Params:     map[string]interface{}{"path": "/etc/config.yaml"},
		},
		Risk: approval.RiskIndicator{Level: "high", Label: "Modifies filesystem"},
	}
	output := renderApprovalDialog(vm, 80, 40, 0, false)

	assert.NotContains(t, output, "[Telegram]", "dialog should not show Telegram for non-channel key")
	assert.NotContains(t, output, "[Discord]", "dialog should not show Discord for non-channel key")
	assert.NotContains(t, output, "[Slack]", "dialog should not show Slack for non-channel key")
}
