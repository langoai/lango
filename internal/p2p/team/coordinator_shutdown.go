package team

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/eventbus"
)

// GracefulShutdown performs an ordered team shutdown:
//  1. Set team status to StatusShuttingDown (blocks new task delegation)
//  2. Calculate proportional milestone settlement for completed work
//  3. Publish TeamGracefulShutdownEvent
//  4. Disband team with reason
func (c *Coordinator) GracefulShutdown(ctx context.Context, teamID string, reason string) error {
	t, err := c.GetTeam(teamID)
	if err != nil {
		return err
	}

	// 1. Block new tasks.
	t.mu.Lock()
	if t.Status == StatusShuttingDown || t.Status == StatusDisbanded {
		t.mu.Unlock()
		return fmt.Errorf("team %s already %s", teamID, t.Status)
	}
	t.Status = StatusShuttingDown
	t.mu.Unlock()

	c.logger.Infow("team graceful shutdown started", "teamID", teamID, "reason", reason)

	// 2. Count settled members: if the team has recorded spend, all active members contributed.
	members := t.ActiveMembers()
	settledCount := 0
	if t.Spent > 0 {
		settledCount = len(members)
	}

	// 3. Persist shutting-down state.
	if c.store != nil {
		if err := c.store.Save(t); err != nil {
			c.logger.Warnw("persist team during shutdown", "teamID", teamID, "error", err)
		}
	}

	// 4. Publish graceful shutdown event.
	if c.bus != nil {
		c.bus.Publish(eventbus.TeamGracefulShutdownEvent{
			TeamID:         teamID,
			Reason:         reason,
			BundlesCreated: 0, // workspace bundles handled by bridge layer
			MembersSettled: settledCount,
		})
	}

	// 5. Disband the team.
	if err := c.DisbandTeam(teamID); err != nil {
		return fmt.Errorf("disband team during shutdown: %w", err)
	}

	c.logger.Infow("team graceful shutdown completed",
		"teamID", teamID,
		"reason", reason,
		"membersSettled", settledCount,
	)
	return nil
}
