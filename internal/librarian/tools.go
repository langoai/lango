package librarian

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolparam"
)

// BuildTools creates proactive librarian agent tools.
func BuildTools(is *InquiryStore) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "librarian_pending_inquiries",
			Description: "List pending knowledge inquiries for the current session",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "librarian",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"session_key": map[string]interface{}{"type": "string", "description": "Session key (uses current session if empty)"},
					"limit":       map[string]interface{}{"type": "integer", "description": "Maximum results (default: 5)"},
				},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				sessionKey := toolparam.OptionalString(params, "session_key", session.SessionKeyFromContext(ctx))
				limit := toolparam.OptionalInt(params, "limit", 5)
				inquiries, err := is.ListPendingInquiries(ctx, sessionKey, limit)
				if err != nil {
					return nil, fmt.Errorf("list pending inquiries: %w", err)
				}
				return map[string]interface{}{"inquiries": inquiries, "count": len(inquiries)}, nil
			},
		},
		{
			Name:        "librarian_dismiss_inquiry",
			Description: "Dismiss a pending knowledge inquiry that the user does not want to answer",
			SafetyLevel: agent.SafetyLevelModerate,
			Capability: agent.ToolCapability{
				Category: "librarian",
				Activity: agent.ActivityManage,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"inquiry_id": map[string]interface{}{"type": "string", "description": "UUID of the inquiry to dismiss"},
				},
				"required": []string{"inquiry_id"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				idStr, err := toolparam.RequireString(params, "inquiry_id")
				if err != nil {
					return nil, err
				}
				id, err := uuid.Parse(idStr)
				if err != nil {
					return nil, fmt.Errorf("invalid inquiry_id: %w", err)
				}
				if err := is.DismissInquiry(ctx, id); err != nil {
					return nil, fmt.Errorf("dismiss inquiry: %w", err)
				}
				return map[string]interface{}{
					"status":  "dismissed",
					"message": fmt.Sprintf("Inquiry %s dismissed", idStr),
				}, nil
			},
		},
	}
}
