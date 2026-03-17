package app

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/workspace"
)

// wireWorkspaceTeamBridge subscribes to team events and auto-manages workspace lifecycle.
func wireWorkspaceTeamBridge(
	bus *eventbus.Bus,
	wsMgr *workspace.Manager,
	tracker *workspace.ContributionTracker,
	gossip *workspace.WorkspaceGossip,
	log *zap.SugaredLogger,
) {
	// Map team IDs to workspace IDs for lifecycle tracking.
	var teamWorkspaces sync.Map // teamID → workspaceID

	// TeamFormed → auto-create workspace for the team.
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamFormedEvent) {
		wsName := fmt.Sprintf("team-%s", truncateID(ev.TeamID, 8))

		ws, err := wsMgr.Create(context.Background(), workspace.CreateRequest{
			Name: wsName,
			Goal: ev.Goal,
			Metadata: map[string]string{
				"teamID":    ev.TeamID,
				"teamName":  ev.Name,
				"leaderDID": ev.LeaderDID,
			},
		})
		if err != nil {
			log.Warnw("workspace-team bridge: create workspace", "teamID", ev.TeamID, "error", err)
			return
		}

		teamWorkspaces.Store(ev.TeamID, ws.ID)

		// Subscribe to workspace gossip so team messages propagate.
		if gossip != nil {
			if err := gossip.Subscribe(ws.ID); err != nil {
				log.Warnw("workspace-team bridge: gossip subscribe", "workspaceID", ws.ID, "error", err)
			}
		}

		log.Infow("workspace-team bridge: workspace created",
			"teamID", ev.TeamID,
			"workspaceID", ws.ID,
			"name", wsName,
		)
	})

	// TeamTaskCompleted → record contribution for the completed task.
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamTaskCompletedEvent) {
		if tracker == nil {
			return
		}

		wsIDVal, ok := teamWorkspaces.Load(ev.TeamID)
		if !ok {
			return
		}
		wsID := wsIDVal.(string)

		// Record a message-level contribution under the team ID.
		tracker.RecordMessage(wsID, ev.TeamID)
		log.Debugw("workspace-team bridge: contribution recorded",
			"teamID", ev.TeamID,
			"workspaceID", wsID,
			"tool", ev.ToolName,
			"successful", ev.Successful,
			"failed", ev.Failed,
		)
	})

	// TeamDisbanded → unsubscribe gossip for the workspace.
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamDisbandedEvent) {
		wsIDVal, ok := teamWorkspaces.Load(ev.TeamID)
		if !ok {
			return
		}
		wsID := wsIDVal.(string)
		teamWorkspaces.Delete(ev.TeamID)

		if gossip != nil {
			gossip.Unsubscribe(wsID)
		}

		log.Infow("workspace-team bridge: workspace unlinked on disband",
			"teamID", ev.TeamID,
			"workspaceID", wsID,
			"reason", ev.Reason,
		)
	})

	log.Info("workspace-team bridge wired")
}

// truncateID returns the first n characters of an ID string.
func truncateID(id string, n int) string {
	if len(id) <= n {
		return id
	}
	return id[:n]
}
