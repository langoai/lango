package app

import (
	"context"
	"fmt"
	"time"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/p2p/workspace"
)

// buildWorkspaceTools creates workspace management tools.
func buildWorkspaceTools(wc *wsComponents) []*agent.Tool {
	var tools []*agent.Tool

	// Workspace CRUD tools
	tools = append(tools, &agent.Tool{
		Name:        "p2p_workspace_create",
		Description: "Create a new P2P collaborative workspace for agents to share code and messages",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{"type": "string", "description": "Workspace name"},
				"goal": map[string]interface{}{"type": "string", "description": "Workspace goal/description"},
			},
			"required": []string{"name", "goal"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			name, _ := params["name"].(string)
			goal, _ := params["goal"].(string)
			if name == "" {
				return nil, fmt.Errorf("missing name parameter")
			}

			ws, err := wc.manager.Create(ctx, workspace.CreateRequest{
				Name: name,
				Goal: goal,
			})
			if err != nil {
				return nil, fmt.Errorf("create workspace: %w", err)
			}

			// Subscribe to gossip topic.
			if wc.gossip != nil {
				_ = wc.gossip.Subscribe(ws.ID)
			}

			return map[string]interface{}{
				"id":        ws.ID,
				"name":      ws.Name,
				"goal":      ws.Goal,
				"status":    string(ws.Status),
				"createdAt": ws.CreatedAt.Format(time.RFC3339),
			}, nil
		},
	})

	tools = append(tools, &agent.Tool{
		Name:        "p2p_workspace_join",
		Description: "Join an existing P2P workspace",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspaceId": map[string]interface{}{"type": "string", "description": "Workspace ID to join"},
			},
			"required": []string{"workspaceId"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			wsID, _ := params["workspaceId"].(string)
			if wsID == "" {
				return nil, fmt.Errorf("missing workspaceId parameter")
			}

			if err := wc.manager.Join(ctx, wsID); err != nil {
				return nil, err
			}

			if wc.gossip != nil {
				_ = wc.gossip.Subscribe(wsID)
			}

			return map[string]interface{}{"joined": wsID}, nil
		},
	})

	tools = append(tools, &agent.Tool{
		Name:        "p2p_workspace_leave",
		Description: "Leave a P2P workspace",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspaceId": map[string]interface{}{"type": "string", "description": "Workspace ID to leave"},
			},
			"required": []string{"workspaceId"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			wsID, _ := params["workspaceId"].(string)
			if wsID == "" {
				return nil, fmt.Errorf("missing workspaceId parameter")
			}

			if err := wc.manager.Leave(ctx, wsID); err != nil {
				return nil, err
			}

			if wc.gossip != nil {
				wc.gossip.Unsubscribe(wsID)
			}

			return map[string]interface{}{"left": wsID}, nil
		},
	})

	tools = append(tools, &agent.Tool{
		Name:        "p2p_workspace_list",
		Description: "List all P2P workspaces",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			list, err := wc.manager.List(ctx)
			if err != nil {
				return nil, err
			}

			result := make([]map[string]interface{}, 0, len(list))
			for _, ws := range list {
				result = append(result, map[string]interface{}{
					"id":      ws.ID,
					"name":    ws.Name,
					"goal":    ws.Goal,
					"status":  string(ws.Status),
					"members": len(ws.Members),
				})
			}
			return map[string]interface{}{"workspaces": result, "count": len(result)}, nil
		},
	})

	tools = append(tools, &agent.Tool{
		Name:        "p2p_workspace_status",
		Description: "Show detailed status of a P2P workspace",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspaceId": map[string]interface{}{"type": "string", "description": "Workspace ID"},
			},
			"required": []string{"workspaceId"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			wsID, _ := params["workspaceId"].(string)
			if wsID == "" {
				return nil, fmt.Errorf("missing workspaceId parameter")
			}

			ws, err := wc.manager.Get(ctx, wsID)
			if err != nil {
				return nil, err
			}

			members := make([]map[string]interface{}, 0, len(ws.Members))
			for _, m := range ws.Members {
				members = append(members, map[string]interface{}{
					"did":      m.DID,
					"name":     m.Name,
					"role":     m.Role,
					"joinedAt": m.JoinedAt.Format(time.RFC3339),
				})
			}

			result := map[string]interface{}{
				"id":        ws.ID,
				"name":      ws.Name,
				"goal":      ws.Goal,
				"status":    string(ws.Status),
				"members":   members,
				"createdAt": ws.CreatedAt.Format(time.RFC3339),
				"updatedAt": ws.UpdatedAt.Format(time.RFC3339),
			}

			// Add contribution data if tracking is enabled.
			if wc.tracker != nil {
				contribs := wc.tracker.List(wsID)
				contribData := make([]map[string]interface{}, 0, len(contribs))
				for _, c := range contribs {
					contribData = append(contribData, map[string]interface{}{
						"did":        c.DID,
						"commits":    c.Commits,
						"codeBytes":  c.CodeBytes,
						"messages":   c.Messages,
						"lastActive": c.LastActive.Format(time.RFC3339),
					})
				}
				result["contributions"] = contribData
			}

			return result, nil
		},
	})

	// Workspace messaging tools
	tools = append(tools, &agent.Tool{
		Name:        "p2p_workspace_post",
		Description: "Post a message to a P2P workspace (broadcast to all members via GossipSub)",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspaceId": map[string]interface{}{"type": "string", "description": "Workspace ID"},
				"type":        map[string]interface{}{"type": "string", "description": "Message type: TASK_PROPOSAL, LOG_STREAM, COMMIT_SIGNAL, KNOWLEDGE_SHARE"},
				"content":     map[string]interface{}{"type": "string", "description": "Message content"},
				"parentId":    map[string]interface{}{"type": "string", "description": "Parent message ID (for replies)"},
			},
			"required": []string{"workspaceId", "content"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			wsID, _ := params["workspaceId"].(string)
			content, _ := params["content"].(string)
			msgType, _ := params["type"].(string)
			parentID, _ := params["parentId"].(string)

			if wsID == "" || content == "" {
				return nil, fmt.Errorf("missing workspaceId or content parameter")
			}

			if msgType == "" {
				msgType = string(workspace.MessageTypeKnowledgeShare)
			}

			msg := workspace.Message{
				Type:      workspace.MessageType(msgType),
				Content:   content,
				ParentID:  parentID,
				Timestamp: time.Now(),
			}

			// Persist locally.
			if err := wc.manager.Post(ctx, wsID, msg); err != nil {
				return nil, err
			}

			// Broadcast via GossipSub.
			if wc.gossip != nil {
				msg.WorkspaceID = wsID
				_ = wc.gossip.Publish(ctx, wsID, msg)
			}

			return map[string]interface{}{
				"posted":      true,
				"messageId":   msg.ID,
				"workspaceId": wsID,
			}, nil
		},
	})

	tools = append(tools, &agent.Tool{
		Name:        "p2p_workspace_read",
		Description: "Read messages from a P2P workspace",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspaceId": map[string]interface{}{"type": "string", "description": "Workspace ID"},
				"limit":       map[string]interface{}{"type": "integer", "description": "Max messages to return (default 20)"},
				"type":        map[string]interface{}{"type": "string", "description": "Filter by message type"},
			},
			"required": []string{"workspaceId"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			wsID, _ := params["workspaceId"].(string)
			if wsID == "" {
				return nil, fmt.Errorf("missing workspaceId parameter")
			}

			limit := 20
			if l, ok := params["limit"].(float64); ok {
				limit = int(l)
			}

			opts := workspace.ReadOptions{Limit: limit}
			if t, ok := params["type"].(string); ok && t != "" {
				opts.Types = []string{t}
			}

			messages, err := wc.manager.Read(ctx, wsID, opts)
			if err != nil {
				return nil, err
			}

			result := make([]map[string]interface{}, 0, len(messages))
			for _, m := range messages {
				result = append(result, map[string]interface{}{
					"id":        m.ID,
					"type":      string(m.Type),
					"sender":    m.SenderDID,
					"content":   m.Content,
					"parentId":  m.ParentID,
					"timestamp": m.Timestamp.Format(time.RFC3339),
				})
			}
			return map[string]interface{}{"messages": result, "count": len(result)}, nil
		},
	})

	// Git tools
	if wc.gitService != nil {
		tools = append(tools, buildGitTools(wc)...)
	}

	return tools
}

// buildGitTools creates git bundle tools.
func buildGitTools(wc *wsComponents) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "p2p_git_init",
			Description: "Initialize a git repository for a P2P workspace",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"workspaceId": map[string]interface{}{"type": "string", "description": "Workspace ID"},
				},
				"required": []string{"workspaceId"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				wsID, _ := params["workspaceId"].(string)
				if wsID == "" {
					return nil, fmt.Errorf("missing workspaceId parameter")
				}
				if err := wc.gitService.Init(ctx, wsID); err != nil {
					return nil, err
				}
				return map[string]interface{}{"initialized": true, "workspaceId": wsID}, nil
			},
		},
		{
			Name:        "p2p_git_push",
			Description: "Create a git bundle from workspace repo and broadcast to peers",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"workspaceId": map[string]interface{}{"type": "string", "description": "Workspace ID"},
					"message":     map[string]interface{}{"type": "string", "description": "Push description"},
				},
				"required": []string{"workspaceId"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				wsID, _ := params["workspaceId"].(string)
				message, _ := params["message"].(string)
				if wsID == "" {
					return nil, fmt.Errorf("missing workspaceId parameter")
				}

				bundle, hash, err := wc.gitService.CreateBundle(ctx, wsID)
				if err != nil {
					return nil, err
				}
				if bundle == nil {
					return map[string]interface{}{"pushed": false, "reason": "empty repository"}, nil
				}

				result := map[string]interface{}{
					"pushed":     true,
					"headCommit": hash,
					"bundleSize": len(bundle),
					"message":    message,
				}

				// Broadcast commit signal via workspace gossip.
				if wc.gossip != nil {
					msg := workspace.Message{
						Type:    workspace.MessageTypeCommitSignal,
						Content: fmt.Sprintf("pushed bundle (head: %s): %s", hash, message),
						Metadata: map[string]string{
							"headCommit": hash,
							"bundleSize": fmt.Sprintf("%d", len(bundle)),
						},
						Timestamp: time.Now(),
					}
					_ = wc.gossip.Publish(ctx, wsID, msg)
				}

				return result, nil
			},
		},
		{
			Name:        "p2p_git_log",
			Description: "Show commit log for a workspace's git repository",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"workspaceId": map[string]interface{}{"type": "string", "description": "Workspace ID"},
					"limit":       map[string]interface{}{"type": "integer", "description": "Max commits (default 20)"},
				},
				"required": []string{"workspaceId"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				wsID, _ := params["workspaceId"].(string)
				if wsID == "" {
					return nil, fmt.Errorf("missing workspaceId parameter")
				}
				limit := 20
				if l, ok := params["limit"].(float64); ok {
					limit = int(l)
				}
				commits, err := wc.gitService.Log(ctx, wsID, limit)
				if err != nil {
					return nil, err
				}
				result := make([]map[string]interface{}, 0, len(commits))
				for _, c := range commits {
					result = append(result, map[string]interface{}{
						"hash":      c.Hash,
						"message":   c.Message,
						"author":    c.Author,
						"timestamp": c.Timestamp.Format(time.RFC3339),
					})
				}
				return map[string]interface{}{"commits": result, "count": len(result)}, nil
			},
		},
		{
			Name:        "p2p_git_diff",
			Description: "Show diff between two commits in a workspace repository",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"workspaceId": map[string]interface{}{"type": "string", "description": "Workspace ID"},
					"from":        map[string]interface{}{"type": "string", "description": "Source commit hash"},
					"to":          map[string]interface{}{"type": "string", "description": "Target commit hash"},
				},
				"required": []string{"workspaceId", "from", "to"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				wsID, _ := params["workspaceId"].(string)
				from, _ := params["from"].(string)
				to, _ := params["to"].(string)
				if wsID == "" || from == "" || to == "" {
					return nil, fmt.Errorf("missing required parameters")
				}
				diff, err := wc.gitService.Diff(ctx, wsID, from, to)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{"diff": diff}, nil
			},
		},
		{
			Name:        "p2p_git_leaves",
			Description: "Find DAG leaf commits (commits with no children) in a workspace repository",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"workspaceId": map[string]interface{}{"type": "string", "description": "Workspace ID"},
				},
				"required": []string{"workspaceId"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				wsID, _ := params["workspaceId"].(string)
				if wsID == "" {
					return nil, fmt.Errorf("missing workspaceId parameter")
				}
				leaves, err := wc.gitService.Leaves(ctx, wsID)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{"leaves": leaves, "count": len(leaves)}, nil
			},
		},
	}
}
