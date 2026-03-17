package protocol

import (
	"context"
	"encoding/json"
	"fmt"
)

// TeamInviteHandler handles team invitation requests.
type TeamInviteHandler func(ctx context.Context, peerDID string, payload TeamInvitePayload) (map[string]interface{}, error)

// TeamAcceptHandler handles team acceptance responses.
type TeamAcceptHandler func(ctx context.Context, peerDID string, payload TeamAcceptPayload) (map[string]interface{}, error)

// TeamTaskHandler handles team task delegation.
type TeamTaskHandler func(ctx context.Context, peerDID string, payload TeamTaskPayload) (map[string]interface{}, error)

// TeamResultHandler handles team task results.
type TeamResultHandler func(ctx context.Context, peerDID string, payload TeamResultPayload) (map[string]interface{}, error)

// TeamDisbandHandler handles team disband notifications.
type TeamDisbandHandler func(ctx context.Context, peerDID string, payload TeamDisbandPayload) (map[string]interface{}, error)

// TeamRouter dispatches team messages to type-specific handlers.
type TeamRouter struct {
	OnInvite  TeamInviteHandler
	OnAccept  TeamAcceptHandler
	OnTask    TeamTaskHandler
	OnResult  TeamResultHandler
	OnDisband TeamDisbandHandler
}

// Handle routes a team request to the appropriate handler based on request type.
func (r *TeamRouter) Handle(ctx context.Context, peerDID string, reqType RequestType, payload map[string]interface{}) (map[string]interface{}, error) {
	// Marshal/unmarshal for typed deserialization.
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal team payload: %w", err)
	}

	switch reqType {
	case RequestTeamInvite:
		if r.OnInvite == nil {
			return nil, fmt.Errorf("team invite handler not configured")
		}
		var p TeamInvitePayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("decode team invite: %w", err)
		}
		return r.OnInvite(ctx, peerDID, p)

	case RequestTeamAccept:
		if r.OnAccept == nil {
			return nil, fmt.Errorf("team accept handler not configured")
		}
		var p TeamAcceptPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("decode team accept: %w", err)
		}
		return r.OnAccept(ctx, peerDID, p)

	case RequestTeamTask:
		if r.OnTask == nil {
			return nil, fmt.Errorf("team task handler not configured")
		}
		var p TeamTaskPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("decode team task: %w", err)
		}
		return r.OnTask(ctx, peerDID, p)

	case RequestTeamResult:
		if r.OnResult == nil {
			return nil, fmt.Errorf("team result handler not configured")
		}
		var p TeamResultPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("decode team result: %w", err)
		}
		return r.OnResult(ctx, peerDID, p)

	case RequestTeamDisband:
		if r.OnDisband == nil {
			return nil, fmt.Errorf("team disband handler not configured")
		}
		var p TeamDisbandPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("decode team disband: %w", err)
		}
		return r.OnDisband(ctx, peerDID, p)

	default:
		return nil, fmt.Errorf("unknown team request type: %s", reqType)
	}
}
