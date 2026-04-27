package app

import (
	"context"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/reputation"
	"github.com/langoai/lango/internal/p2p/team"
)

// initTeamReputationBridge wires events between the team coordinator and reputation system:
//  1. TeamMemberUnhealthyEvent -> RecordOperationalIncident -> consult trust entry -> kick if runtime entry is no longer allowed
//  2. TeamTaskCompletedEvent -> RecordSuccess for successful workers
//  3. ReputationChangedEvent -> consult trust entry -> TeamsForMember -> KickMember
func initTeamReputationBridge(
	bus *eventbus.Bus,
	coordinator *team.Coordinator,
	repStore *reputation.Store,
	minScore float64,
	log *zap.SugaredLogger,
) {
	ctx := context.Background()

	// 1. Unhealthy member -> record operational incident and possibly kick.
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamMemberUnhealthyEvent) {
		if repStore != nil {
			if err := repStore.RecordOperationalIncident(ctx, ev.MemberDID); err != nil {
				log.Errorw("team-reputation bridge: record operational incident", "did", ev.MemberDID, "error", err)
			}

			entry, err := runtimeTrustEntry(ctx, repStore, ev.MemberDID, minScore)
			if err != nil {
				log.Errorw("team-reputation bridge: get trust entry", "did", ev.MemberDID, "error", err)
				return
			}

			if reason, shouldKick := runtimeTrustKickReason(entry); shouldKick {
				if err := coordinator.KickMember(ctx, ev.TeamID, ev.MemberDID, reason); err != nil {
					log.Warnw("team-reputation bridge: kick unhealthy member",
						"teamID", ev.TeamID, "did", ev.MemberDID, "error", err)
				} else {
					log.Infow("team-reputation bridge: kicked unhealthy member",
						"teamID", ev.TeamID, "did", ev.MemberDID, "state", entry.State)
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
					log.Errorw("team-reputation bridge: record success",
						"did", m.DID, "error", err)
				}
			}
		}
	})

	// 3. Reputation changed -> consult trust entry -> kick from all teams.
	eventbus.SubscribeTyped(bus, func(ev eventbus.ReputationChangedEvent) {
		if repStore == nil {
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
			return
		}

		entry, err := runtimeTrustEntry(ctx, repStore, ev.PeerDID, minScore)
		if err != nil {
			log.Errorw("team-reputation bridge: get trust entry on change", "did", ev.PeerDID, "error", err)
			return
		}
		reason, shouldKick := runtimeTrustKickReason(entry)
		if !shouldKick {
			return
		}

		teamIDs := coordinator.TeamsForMember(ev.PeerDID)
		for _, teamID := range teamIDs {
			if err := coordinator.KickMember(ctx, teamID, ev.PeerDID, reason); err != nil {
				log.Warnw("team-reputation bridge: kick on reputation drop",
					"teamID", teamID, "did", ev.PeerDID, "error", err)
			} else {
				log.Infow("team-reputation bridge: kicked member on reputation drop",
					"teamID", teamID, "did", ev.PeerDID, "state", entry.State)
			}
		}
	})

	log.Infow("team-reputation bridge wired", "minReputationScore", minScore)
}
