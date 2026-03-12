package app

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/team"
)

// wireTeamEscrowBridge subscribes to team events and auto-manages escrow lifecycle.
func wireTeamEscrowBridge(bus *eventbus.Bus, escrowEngine *escrow.Engine, coord *team.Coordinator, log *zap.SugaredLogger) {
	// Map team IDs to escrow IDs for lifecycle tracking.
	var teamEscrows sync.Map // teamID -> escrowID

	// TeamFormed -> create escrow if team has budget > 0.
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamFormedEvent) {
		t, err := coord.GetTeam(ev.TeamID)
		if err != nil {
			log.Debugw("team-escrow bridge: team not found", "teamID", ev.TeamID, "error", err)
			return
		}

		if t.Budget <= 0 {
			return
		}

		// Find workers to create per-worker milestones.
		members := t.Members()
		var workers []*team.Member
		for _, m := range members {
			if m.Role == team.RoleWorker {
				workers = append(workers, m)
			}
		}
		if len(workers) == 0 {
			return
		}

		// Split budget equally among workers as milestones.
		totalAmount := floatToMicroUSDC(t.Budget)
		perWorker := new(big.Int).Div(totalAmount, big.NewInt(int64(len(workers))))
		// Adjust last worker to account for integer division rounding.
		remainder := new(big.Int).Sub(totalAmount, new(big.Int).Mul(perWorker, big.NewInt(int64(len(workers)))))

		milestones := make([]escrow.MilestoneRequest, len(workers))
		for i, w := range workers {
			amount := new(big.Int).Set(perWorker)
			if i == len(workers)-1 {
				amount.Add(amount, remainder)
			}
			milestones[i] = escrow.MilestoneRequest{
				Description: fmt.Sprintf("Task completion by %s", w.DID),
				Amount:      amount,
			}
		}

		// First worker is the "seller" for escrow purposes.
		sellerDID := workers[0].DID

		entry, err := escrowEngine.Create(context.Background(), escrow.CreateRequest{
			BuyerDID:   t.LeaderDID,
			SellerDID:  sellerDID,
			Amount:     totalAmount,
			Reason:     fmt.Sprintf("Team %s: %s", ev.Name, ev.Goal),
			TaskID:     ev.TeamID,
			Milestones: milestones,
		})
		if err != nil {
			log.Warnw("team-escrow bridge: create escrow", "teamID", ev.TeamID, "error", err)
			return
		}

		teamEscrows.Store(ev.TeamID, entry.ID)
		log.Infow("team-escrow bridge: escrow created",
			"teamID", ev.TeamID,
			"escrowID", entry.ID,
			"amount", totalAmount.String(),
			"milestones", len(milestones),
		)
	})

	// TeamTaskCompleted -> complete next pending milestone.
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamTaskCompletedEvent) {
		escrowIDVal, ok := teamEscrows.Load(ev.TeamID)
		if !ok {
			return
		}
		escrowID := escrowIDVal.(string)

		entry, err := escrowEngine.Get(escrowID)
		if err != nil {
			log.Debugw("team-escrow bridge: escrow not found", "escrowID", escrowID, "error", err)
			return
		}

		// CompleteMilestone requires StatusActive; skip if not yet active.
		if entry.Status != escrow.StatusActive {
			log.Debugw("team-escrow bridge: escrow not active, skipping milestone completion",
				"escrowID", escrowID, "status", entry.Status)
			return
		}

		// Find next pending milestone.
		for _, m := range entry.Milestones {
			if m.Status == escrow.MilestonePending {
				evidence := fmt.Sprintf("tool=%s successful=%d failed=%d duration=%s",
					ev.ToolName, ev.Successful, ev.Failed, ev.Duration)
				if _, err := escrowEngine.CompleteMilestone(context.Background(), escrowID, m.ID, evidence); err != nil {
					log.Warnw("team-escrow bridge: complete milestone",
						"escrowID", escrowID, "milestoneID", m.ID, "error", err)
				} else {
					log.Infow("team-escrow bridge: milestone completed",
						"escrowID", escrowID, "milestoneID", m.ID)
				}
				break
			}
		}
	})

	// TeamDisbanded -> release if all milestones done, otherwise dispute+refund.
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamDisbandedEvent) {
		escrowIDVal, ok := teamEscrows.Load(ev.TeamID)
		if !ok {
			return
		}
		escrowID := escrowIDVal.(string)
		teamEscrows.Delete(ev.TeamID)

		entry, err := escrowEngine.Get(escrowID)
		if err != nil {
			log.Debugw("team-escrow bridge: escrow not found on disband",
				"escrowID", escrowID, "error", err)
			return
		}

		if entry.AllMilestonesCompleted() {
			if _, err := escrowEngine.Release(context.Background(), escrowID); err != nil {
				log.Warnw("team-escrow bridge: release on disband",
					"escrowID", escrowID, "error", err)
			} else {
				log.Infow("team-escrow bridge: escrow released on disband",
					"escrowID", escrowID)
			}
		} else {
			// Dispute first (required transition before refund from active/completed).
			reason := ev.Reason
			if reason == "" {
				reason = "team disbanded with incomplete milestones"
			}
			if _, err := escrowEngine.Dispute(context.Background(), escrowID, reason); err != nil {
				log.Warnw("team-escrow bridge: dispute on disband",
					"escrowID", escrowID, "error", err)
				return
			}
			if _, err := escrowEngine.Refund(context.Background(), escrowID); err != nil {
				log.Warnw("team-escrow bridge: refund on disband",
					"escrowID", escrowID, "error", err)
			} else {
				log.Infow("team-escrow bridge: escrow refunded on disband",
					"escrowID", escrowID)
			}
		}
	})

	log.Info("team-escrow bridge wired")
}

