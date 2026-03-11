package app

import (
	"context"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/reputation"
	"github.com/langoai/lango/internal/p2p/team"
)

// initTeamReputationBridge wires events between the team coordinator and reputation system:
//  1. TeamMemberUnhealthyEvent -> RecordTimeout -> check score -> KickMember if below threshold
//  2. TeamTaskCompletedEvent -> RecordSuccess for successful workers
//  3. ReputationChangedEvent -> check if dropped below threshold -> TeamsForMember -> KickMember
func initTeamReputationBridge(
	bus *eventbus.Bus,
	coordinator *team.Coordinator,
	repStore *reputation.Store,
	minScore float64,
	log *zap.SugaredLogger,
) {
	ctx := context.Background()

	// 1. Unhealthy member -> record timeout and possibly kick.
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamMemberUnhealthyEvent) {
		if repStore != nil {
			if err := repStore.RecordTimeout(ctx, ev.MemberDID); err != nil {
				log.Debugw("team-reputation bridge: record timeout", "did", ev.MemberDID, "error", err)
			}

			score, err := repStore.GetScore(ctx, ev.MemberDID)
			if err != nil {
				log.Debugw("team-reputation bridge: get score", "did", ev.MemberDID, "error", err)
				return
			}

			if score < minScore {
				if err := coordinator.KickMember(ctx, ev.TeamID, ev.MemberDID, "reputation below threshold after unhealthy"); err != nil {
					log.Debugw("team-reputation bridge: kick unhealthy member",
						"teamID", ev.TeamID, "did", ev.MemberDID, "error", err)
				} else {
					log.Infow("team-reputation bridge: kicked unhealthy member",
						"teamID", ev.TeamID, "did", ev.MemberDID, "score", score)
				}
			}
		}
	})

	// 2. Task completed -> record success for successful workers.
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamTaskCompletedEvent) {
		if repStore == nil || ev.Successful <= 0 {
			return
		}

		t, err := coordinator.GetTeam(ev.TeamID)
		if err != nil {
			log.Debugw("team-reputation bridge: team not found for task completion",
				"teamID", ev.TeamID, "error", err)
			return
		}

		for _, m := range t.Members() {
			if m.Role == team.RoleWorker && m.Status != team.MemberFailed {
				if err := repStore.RecordSuccess(ctx, m.DID); err != nil {
					log.Debugw("team-reputation bridge: record success",
						"did", m.DID, "error", err)
				}
			}
		}
	})

	// 3. Reputation changed -> check if dropped below threshold -> kick from all teams.
	eventbus.SubscribeTyped(bus, func(ev eventbus.ReputationChangedEvent) {
		if ev.NewScore >= minScore {
			return
		}

		teamIDs := coordinator.TeamsForMember(ev.PeerDID)
		for _, teamID := range teamIDs {
			if err := coordinator.KickMember(ctx, teamID, ev.PeerDID, "reputation dropped below threshold"); err != nil {
				log.Debugw("team-reputation bridge: kick on reputation drop",
					"teamID", teamID, "did", ev.PeerDID, "error", err)
			} else {
				log.Infow("team-reputation bridge: kicked member on reputation drop",
					"teamID", teamID, "did", ev.PeerDID, "newScore", ev.NewScore)
			}
		}
	})

	log.Infow("team-reputation bridge wired", "minReputationScore", minScore)
}
