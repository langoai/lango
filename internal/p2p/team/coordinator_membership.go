package team

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/eventbus"
)

// KickMember removes a member from a team with a reason and publishes a TeamMemberLeftEvent.
func (c *Coordinator) KickMember(_ context.Context, teamID string, memberDID string, reason string) error {
	t, err := c.GetTeam(teamID)
	if err != nil {
		return err
	}

	if err := t.RemoveMember(memberDID); err != nil {
		return fmt.Errorf("kick member %s from team %s: %w", memberDID, teamID, err)
	}

	// Persist updated team.
	if c.store != nil {
		if err := c.store.Save(t); err != nil {
			c.logger.Warnw("persist team after kick", "teamID", teamID, "error", err)
		}
	}

	// Publish leave event.
	if c.bus != nil {
		c.bus.Publish(eventbus.TeamMemberLeftEvent{
			TeamID:    teamID,
			MemberDID: memberDID,
			Reason:    reason,
		})
	}

	c.logger.Infow("member kicked from team",
		"teamID", teamID, "memberDID", memberDID, "reason", reason)
	return nil
}

// TeamsForMember returns all active team IDs that include the given DID.
func (c *Coordinator) TeamsForMember(did string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var teamIDs []string
	for _, t := range c.teams {
		if t.Status != StatusActive {
			continue
		}
		if m := t.GetMember(did); m != nil {
			teamIDs = append(teamIDs, t.ID)
		}
	}
	return teamIDs
}
