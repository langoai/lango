package app

import (
	"context"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/economy/escrow/sentinel"
	"github.com/langoai/lango/internal/toolparam"
)

// buildSentinelTools creates agent tools for the Security Sentinel engine.
func buildSentinelTools(se *sentinel.Engine) []*agent.Tool {
	return []*agent.Tool{
		sentinelStatusTool(se),
		sentinelAlertsTool(se),
		sentinelConfigTool(se),
		sentinelAcknowledgeTool(se),
	}
}

func sentinelStatusTool(se *sentinel.Engine) *agent.Tool {
	return &agent.Tool{
		Name:        "sentinel_status",
		Description: "Get the Security Sentinel engine status including running state and alert counts",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return se.Status(), nil
		},
	}
}

func sentinelAlertsTool(se *sentinel.Engine) *agent.Tool {
	return &agent.Tool{
		Name:        "sentinel_alerts",
		Description: "List security alerts from the Sentinel engine with optional severity filter",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"severity": map[string]interface{}{
					"type":        "string",
					"description": "Filter by severity level",
					"enum":        []string{"critical", "high", "medium", "low"},
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Maximum number of alerts to return (default: 20)",
				},
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			severity := toolparam.OptionalString(params, "severity", "")
			limit := toolparam.OptionalInt(params, "limit", 20)

			var alerts []sentinel.Alert
			if severity != "" {
				alerts = se.AlertsByLevel(sentinel.AlertSeverity(severity))
			} else {
				alerts = se.Alerts()
			}

			if len(alerts) > limit {
				alerts = alerts[len(alerts)-limit:]
			}

			items := make([]map[string]interface{}, len(alerts))
			for i, a := range alerts {
				items[i] = map[string]interface{}{
					"id":           a.ID,
					"severity":     string(a.Severity),
					"type":         a.Type,
					"message":      a.Message,
					"timestamp":    a.Timestamp.Format("2006-01-02T15:04:05Z"),
					"acknowledged": a.Acknowledged,
				}
				if a.DealID != "" {
					items[i]["dealId"] = a.DealID
				}
				if a.PeerDID != "" {
					items[i]["peerDid"] = a.PeerDID
				}
			}

			return map[string]interface{}{
				"count":  len(items),
				"alerts": items,
			}, nil
		},
	}
}

func sentinelConfigTool(se *sentinel.Engine) *agent.Tool {
	return &agent.Tool{
		Name:        "sentinel_config",
		Description: "Show current Security Sentinel detection thresholds and configuration",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			cfg := se.Config()
			return map[string]interface{}{
				"rapidCreationWindow":   cfg.RapidCreationWindow.String(),
				"rapidCreationMax":      cfg.RapidCreationMax,
				"largeWithdrawalAmount": cfg.LargeWithdrawalAmount,
				"disputeWindow":         cfg.DisputeWindow.String(),
				"disputeMax":            cfg.DisputeMax,
				"washTradeWindow":       cfg.WashTradeWindow.String(),
			}, nil
		},
	}
}

func sentinelAcknowledgeTool(se *sentinel.Engine) *agent.Tool {
	return &agent.Tool{
		Name:        "sentinel_acknowledge",
		Description: "Acknowledge and dismiss a security alert by ID",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"alertId": map[string]interface{}{"type": "string", "description": "Alert ID to acknowledge"},
			},
			"required": []string{"alertId"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			alertID, err := toolparam.RequireString(params, "alertId")
			if err != nil {
				return nil, err
			}

			if err := se.Acknowledge(alertID); err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"alertId":      alertID,
				"acknowledged": true,
			}, nil
		},
	}
}
