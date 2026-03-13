package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/toolparam"
	"github.com/langoai/lango/internal/p2p/team"
)

// buildTeamTools creates team coordination tools.
func buildTeamTools(coord *team.Coordinator) []*agent.Tool {
	var tools []*agent.Tool

	// 1. team_form — creates a new team
	tools = append(tools, &agent.Tool{
		Name:        "team_form",
		Description: "Form a new P2P agent team by selecting agents with a specific capability",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name":        map[string]interface{}{"type": "string", "description": "Team name"},
				"goal":        map[string]interface{}{"type": "string", "description": "Team goal/mission"},
				"capability":  map[string]interface{}{"type": "string", "description": "Required capability for workers"},
				"memberCount": map[string]interface{}{"type": "integer", "description": "Number of workers to recruit"},
				"leaderDid":   map[string]interface{}{"type": "string", "description": "DID of the team leader"},
			},
			"required": []string{"name", "goal", "capability", "memberCount", "leaderDid"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			name, err := toolparam.RequireString(params, "name")
			if err != nil {
				return nil, err
			}
			goal := toolparam.OptionalString(params, "goal", "")
			capability, err := toolparam.RequireString(params, "capability")
			if err != nil {
				return nil, err
			}
			leaderDID, err := toolparam.RequireString(params, "leaderDid")
			if err != nil {
				return nil, err
			}
			memberCount := toolparam.OptionalInt(params, "memberCount", 1)

			t, err := coord.FormTeam(ctx, team.FormTeamRequest{
				TeamID:      uuid.New().String(),
				Name:        name,
				Goal:        goal,
				LeaderDID:   leaderDID,
				Capability:  capability,
				MemberCount: memberCount,
			})
			if err != nil {
				return nil, fmt.Errorf("form team: %w", err)
			}

			allMembers := t.Members()
			members := make([]map[string]interface{}, 0, len(allMembers))
			for _, m := range allMembers {
				members = append(members, map[string]interface{}{
					"did":    m.DID,
					"name":   m.Name,
					"role":   string(m.Role),
					"status": string(m.Status),
				})
			}

			return map[string]interface{}{
				"teamId":    t.ID,
				"name":      t.Name,
				"goal":      t.Goal,
				"status":    string(t.Status),
				"members":   members,
				"createdAt": t.CreatedAt.Format(time.RFC3339),
			}, nil
		},
	})

	// 2. team_delegate — delegates a task to team workers and collects results
	tools = append(tools, &agent.Tool{
		Name:        "team_delegate",
		Description: "Delegate a tool invocation to all workers in a team and resolve conflicts",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"teamId":   map[string]interface{}{"type": "string", "description": "Team ID"},
				"toolName": map[string]interface{}{"type": "string", "description": "Tool to invoke on workers"},
				"params":   map[string]interface{}{"type": "object", "description": "Parameters to pass to the tool"},
			},
			"required": []string{"teamId", "toolName"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			teamID, err := toolparam.RequireString(params, "teamId")
			if err != nil {
				return nil, err
			}
			toolName, err := toolparam.RequireString(params, "toolName")
			if err != nil {
				return nil, err
			}
			toolParams, _ := params["params"].(map[string]interface{})
			if toolParams == nil {
				toolParams = map[string]interface{}{}
			}

			results, err := coord.DelegateTask(ctx, teamID, toolName, toolParams)
			if err != nil {
				return nil, fmt.Errorf("delegate task: %w", err)
			}

			resolved, resolveErr := coord.CollectResults(teamID, toolName, results)

			// Build individual result summaries.
			resultSummaries := make([]map[string]interface{}, 0, len(results))
			for _, r := range results {
				entry := map[string]interface{}{
					"memberDid": r.MemberDID,
					"duration":  r.Duration.String(),
				}
				if r.Err != nil {
					entry["error"] = r.Err.Error()
				} else {
					entry["result"] = r.Result
				}
				resultSummaries = append(resultSummaries, entry)
			}

			response := map[string]interface{}{
				"teamId":            teamID,
				"toolName":          toolName,
				"individualResults": resultSummaries,
			}

			if resolveErr != nil {
				response["conflictError"] = resolveErr.Error()
			} else {
				response["resolvedResult"] = resolved
			}

			return response, nil
		},
	})

	// 3. team_status — returns detailed team information
	tools = append(tools, &agent.Tool{
		Name:        "team_status",
		Description: "Show detailed status of a team including members and budget",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"teamId": map[string]interface{}{"type": "string", "description": "Team ID"},
			},
			"required": []string{"teamId"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			teamID, err := toolparam.RequireString(params, "teamId")
			if err != nil {
				return nil, err
			}

			t, err := coord.GetTeam(teamID)
			if err != nil {
				return nil, err
			}

			allMembers := t.Members()
			members := make([]map[string]interface{}, 0, len(allMembers))
			for _, m := range allMembers {
				members = append(members, map[string]interface{}{
					"did":          m.DID,
					"name":         m.Name,
					"role":         string(m.Role),
					"status":       string(m.Status),
					"capabilities": m.Capabilities,
					"trustScore":   m.TrustScore,
					"joinedAt":     m.JoinedAt.Format(time.RFC3339),
				})
			}

			return map[string]interface{}{
				"teamId":    t.ID,
				"name":      t.Name,
				"goal":      t.Goal,
				"status":    string(t.Status),
				"leaderDid": t.LeaderDID,
				"budget":    t.Budget,
				"spent":     t.Spent,
				"members":   members,
				"createdAt": t.CreatedAt.Format(time.RFC3339),
			}, nil
		},
	})

	// 4. team_list — lists all active teams
	tools = append(tools, &agent.Tool{
		Name:        "team_list",
		Description: "List all active P2P agent teams",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			teams := coord.ListTeams()
			result := make([]map[string]interface{}, 0, len(teams))
			for _, t := range teams {
				result = append(result, map[string]interface{}{
					"teamId":  t.ID,
					"name":    t.Name,
					"goal":    t.Goal,
					"status":  string(t.Status),
					"members": t.MemberCount(),
				})
			}
			return map[string]interface{}{"teams": result, "count": len(result)}, nil
		},
	})

	// 5. team_disband — disbands a team
	tools = append(tools, &agent.Tool{
		Name:        "team_disband",
		Description: "Disband an existing P2P agent team",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"teamId": map[string]interface{}{"type": "string", "description": "Team ID to disband"},
			},
			"required": []string{"teamId"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			teamID, err := toolparam.RequireString(params, "teamId")
			if err != nil {
				return nil, err
			}

			if err := coord.DisbandTeam(teamID); err != nil {
				return nil, err
			}
			return map[string]interface{}{"disbanded": teamID}, nil
		},
	})

	return tools
}
