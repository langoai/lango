package settings

import (
	"fmt"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

// NewAlertingForm creates the Alerting configuration form.
func NewAlertingForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Alerting Configuration")

	enabled := &tuicore.Field{
		Key: "alerting_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Alerting.Enabled,
		Description: "Enable operational alerting (requires Observability enabled)",
	}
	form.AddField(enabled)
	isEnabled := func() bool { return enabled.Checked }

	form.AddField(&tuicore.Field{
		Key: "alerting_policy_block_rate", Label: "  Policy Block Rate Threshold", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%d", cfg.Alerting.PolicyBlockRate),
		Placeholder: "10",
		Description: "Policy block events per 5-minute window before triggering alert",
		VisibleWhen: isEnabled,
	})

	form.AddField(&tuicore.Field{
		Key: "alerting_recovery_retries", Label: "  Recovery Retry Threshold", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%d", cfg.Alerting.RecoveryRetries),
		Placeholder: "5",
		Description: "Recovery retry events per session before triggering alert",
		VisibleWhen: isEnabled,
	})

	// Delivery webhook URL (first channel).
	webhookURL := ""
	if len(cfg.Alerting.Delivery) > 0 {
		webhookURL = cfg.Alerting.Delivery[0].WebhookURL
	}
	form.AddField(&tuicore.Field{
		Key: "alerting_webhook_url", Label: "  Webhook URL", Type: tuicore.InputText,
		Value:       webhookURL,
		Placeholder: "https://hooks.slack.com/services/...",
		Description: "Webhook URL for external alert delivery (leave empty to disable)",
		VisibleWhen: isEnabled,
	})

	return &form
}
