package toolchain

import (
	"context"
	"fmt"
	"time"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/wallet"
)

// WithApproval returns a middleware that gates tool execution behind an approval flow.
// Uses fail-closed: denies execution unless explicitly approved.
// The Provider routes requests to the appropriate channel (Gateway, Telegram, Discord, Slack, TTY).
// The GrantStore tracks "always allow" grants to auto-approve repeat invocations within a session.
// When limiter is non-nil, payment tools with an amount below the auto-approve threshold
// are executed without explicit user confirmation.
func WithApproval(ic config.InterceptorConfig, ap approval.Provider, gs *approval.GrantStore, limiter wallet.SpendingLimiter) Middleware {
	return func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
		if !NeedsApproval(tool, ic) {
			return next
		}

		return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			sessionKey := session.SessionKeyFromContext(ctx)
			if target := approval.ApprovalTargetFromContext(ctx); target != "" {
				sessionKey = target
			}

			// Check persistent grant — auto-approve if previously "always allowed".
			if gs != nil && gs.IsGranted(sessionKey, tool.Name) {
				return next(ctx, params)
			}

			// Auto-approve small payments via spending limiter threshold.
			if limiter != nil && (tool.Name == "p2p_pay" || tool.Name == "payment_send") {
				if amountStr, ok := params["amount"].(string); ok && amountStr != "" {
					amt, err := wallet.ParseUSDC(amountStr)
					if err == nil {
						if autoOK, checkErr := limiter.IsAutoApprovable(ctx, amt); checkErr == nil && autoOK {
							return next(ctx, params)
						}
					}
				}
			}

			req := approval.ApprovalRequest{
				ID:         fmt.Sprintf("req-%d", time.Now().UnixNano()),
				ToolName:   tool.Name,
				SessionKey: sessionKey,
				Params:     params,
				Summary:    BuildApprovalSummary(tool.Name, params),
				CreatedAt:  time.Now(),
			}
			resp, err := ap.RequestApproval(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("tool '%s' approval: %w", tool.Name, err)
			}
			if !resp.Approved {
				sk := session.SessionKeyFromContext(ctx)
				if sk == "" {
					return nil, fmt.Errorf("tool '%s' execution denied: no approval channel available (session key missing)", tool.Name)
				}
				return nil, fmt.Errorf("tool '%s' execution denied: user did not approve the action", tool.Name)
			}

			// Record persistent grant for this session+tool.
			if resp.AlwaysAllow && gs != nil {
				gs.Grant(sessionKey, tool.Name)
			}

			return next(ctx, params)
		}
	}
}

// NeedsApproval determines whether a tool requires approval based on the
// configured policy, explicit exemptions, and sensitive tool lists.
func NeedsApproval(t *agent.Tool, ic config.InterceptorConfig) bool {
	// ExemptTools always bypass approval.
	for _, name := range ic.ExemptTools {
		if name == t.Name {
			return false
		}
	}

	// SensitiveTools always require approval.
	for _, name := range ic.SensitiveTools {
		if name == t.Name {
			return true
		}
	}

	switch ic.ApprovalPolicy {
	case config.ApprovalPolicyAll:
		return true
	case config.ApprovalPolicyDangerous:
		return t.SafetyLevel.IsDangerous()
	case config.ApprovalPolicyConfigured:
		return false // only SensitiveTools (handled above)
	case config.ApprovalPolicyNone:
		return false
	default:
		return true // unknown policy → fail-safe
	}
}

// BuildApprovalSummary returns a human-readable description of what a tool
// invocation will do, suitable for display in approval messages.
func BuildApprovalSummary(toolName string, params map[string]interface{}) string {
	switch toolName {
	case "exec", "exec_bg":
		if cmd, ok := params["command"].(string); ok {
			return "Execute: " + Truncate(cmd, 200)
		}
	case "fs_write":
		path, _ := params["path"].(string)
		content, _ := params["content"].(string)
		return fmt.Sprintf("Write to %s (%d bytes)", path, len(content))
	case "fs_edit":
		path, _ := params["path"].(string)
		return "Edit file: " + path
	case "fs_delete":
		path, _ := params["path"].(string)
		return "Delete: " + path
	case "browser_navigate":
		url, _ := params["url"].(string)
		return "Navigate to: " + Truncate(url, 200)
	case "browser_action":
		action, _ := params["action"].(string)
		selector, _ := params["selector"].(string)
		if selector != "" {
			return fmt.Sprintf("Browser %s on: %s", action, Truncate(selector, 100))
		}
		return "Browser action: " + action
	case "secrets_store":
		name, _ := params["name"].(string)
		return "Store secret: " + name
	case "secrets_get":
		name, _ := params["name"].(string)
		return "Retrieve secret: " + name
	case "secrets_delete":
		name, _ := params["name"].(string)
		return "Delete secret: " + name
	case "crypto_encrypt":
		return "Encrypt data"
	case "crypto_decrypt":
		return "Decrypt ciphertext"
	case "crypto_sign":
		return "Generate digital signature"
	case "payment_send":
		amount, _ := params["amount"].(string)
		to, _ := params["to"].(string)
		purpose, _ := params["purpose"].(string)
		return fmt.Sprintf("Send %s USDC to %s (%s)", amount, Truncate(to, 12), Truncate(purpose, 50))
	case "payment_create_wallet":
		return "Create new blockchain wallet"
	case "payment_x402_fetch":
		url, _ := params["url"].(string)
		method, _ := params["method"].(string)
		if method == "" {
			method = "GET"
		}
		return fmt.Sprintf("X402 %s %s (auto-pay enabled)", method, Truncate(url, 150))
	case "cron_add":
		name, _ := params["name"].(string)
		scheduleType, _ := params["schedule_type"].(string)
		schedule, _ := params["schedule"].(string)
		return fmt.Sprintf("Create cron job: %s (%s=%s)", name, scheduleType, schedule)
	case "cron_remove":
		id, _ := params["id"].(string)
		return "Remove cron job: " + id
	case "bg_submit":
		prompt, _ := params["prompt"].(string)
		return "Submit background task: " + Truncate(prompt, 100)
	case "workflow_run":
		filePath, _ := params["file_path"].(string)
		if filePath != "" {
			return "Run workflow: " + filePath
		}
		return "Run inline workflow"
	case "workflow_cancel":
		runID, _ := params["run_id"].(string)
		return "Cancel workflow: " + runID
	case "p2p_pay":
		amount, _ := params["amount"].(string)
		peerDID, _ := params["peer_did"].(string)
		memo, _ := params["memo"].(string)
		if memo == "" {
			memo = "P2P payment"
		}
		return fmt.Sprintf("Pay %s USDC to peer %s (%s)", amount, Truncate(peerDID, 16), Truncate(memo, 50))
	}
	return "Tool: " + toolName
}

// Truncate shortens s to maxLen characters, appending "..." if truncated.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
