package collabview

import (
	"slices"
	"strings"
	"time"
)

type Projector struct{}

func NewProjector() *Projector {
	return &Projector{}
}

func (p *Projector) Project(in ProjectionInput) []CollaborationView {
	views := make([]CollaborationView, 0, len(in.Missions))
	for _, mission := range in.Missions {
		view := CollaborationView{
			MissionID:          mission.MissionID,
			CollaborationState: CollaborationStateSolo,
			UpdatedAt:          mission.UpdatedAt,
		}

		execRefs := make(map[string]struct{}, len(mission.ExecutionRefs))
		for _, ref := range mission.ExecutionRefs {
			ref = strings.TrimSpace(ref)
			if ref == "" {
				continue
			}
			execRefs[ref] = struct{}{}
		}

		participantSet := make(map[string]struct{})
		var latestAttributedOwner string
		var latestAttributedOwnerAt time.Time
		for _, run := range in.AgentRuns {
			if !attributed(execRefs, run.ExecutionRef) {
				continue
			}
			owner := strings.TrimSpace(run.RequestedAgent)
			addParticipant(participantSet, owner)
			if owner != "" && (latestAttributedOwner == "" || run.UpdatedAt.After(latestAttributedOwnerAt)) {
				latestAttributedOwner = owner
				latestAttributedOwnerAt = run.UpdatedAt
			}
			if run.UpdatedAt.After(view.UpdatedAt) {
				view.UpdatedAt = run.UpdatedAt
			}
			switch strings.TrimSpace(run.RuntimeCondition) {
			case "blocked_waiting_approval":
				view.CollaborationState = CollaborationStateBlockedOnApproval
				view.BlockedOn = strings.TrimSpace(run.BlockedReason)
			case "waiting_on_teammate":
				if view.CollaborationState != CollaborationStateBlockedOnApproval {
					view.CollaborationState = CollaborationStateWaitingOnTeammate
				}
			case "recovering", "resuming":
				if higherPriority(CollaborationStateRecovering, view.CollaborationState) {
					view.CollaborationState = CollaborationStateRecovering
				}
			}
		}

		for _, run := range in.RunExecutions {
			if !attributed(execRefs, run.ExecutionRef) {
				continue
			}
			if run.UpdatedAt.After(view.UpdatedAt) {
				view.UpdatedAt = run.UpdatedAt
			}
			if strings.TrimSpace(run.CurrentStepStatus) == "verify_pending" && higherPriority(CollaborationStateReviewing, view.CollaborationState) {
				view.CollaborationState = CollaborationStateReviewing
			}
		}

		for _, delegation := range in.Delegations {
			if !attributed(execRefs, delegation.ExecutionRef) {
				continue
			}
			view.HandoffEdges = append(view.HandoffEdges, HandoffEdge{
				From:      delegation.From,
				To:        delegation.To,
				Timestamp: delegation.Timestamp,
			})
			addParticipant(participantSet, strings.TrimSpace(delegation.From))
			addParticipant(participantSet, strings.TrimSpace(delegation.To))
			if delegation.Timestamp.After(view.UpdatedAt) {
				view.UpdatedAt = delegation.Timestamp
			}
			if higherPriority(CollaborationStateDelegating, view.CollaborationState) {
				view.CollaborationState = CollaborationStateDelegating
			}
		}

		for _, signal := range in.BudgetSignals {
			if !attributed(execRefs, signal.ExecutionRef) {
				continue
			}
			if view.BudgetSignal == nil || signal.Timestamp.After(view.BudgetSignal.Timestamp) {
				view.BudgetSignal = &BudgetSignal{
					Used:      signal.Used,
					Max:       signal.Max,
					Timestamp: signal.Timestamp,
				}
			}
			if signal.Timestamp.After(view.UpdatedAt) {
				view.UpdatedAt = signal.Timestamp
			}
		}

		for _, signal := range in.RecoverySignals {
			if !attributed(execRefs, signal.ExecutionRef) {
				continue
			}
			if view.LastRecovery == nil || signal.Timestamp.After(view.LastRecovery.Timestamp) {
				view.LastRecovery = &RecoverySignal{
					Action:     signal.Action,
					CauseClass: signal.CauseClass,
					Timestamp:  signal.Timestamp,
				}
			}
			if signal.Timestamp.After(view.UpdatedAt) {
				view.UpdatedAt = signal.Timestamp
			}
			if higherPriority(CollaborationStateRecovering, view.CollaborationState) {
				view.CollaborationState = CollaborationStateRecovering
			}
		}

		if len(view.HandoffEdges) > 0 {
			slices.SortFunc(view.HandoffEdges, func(a, b HandoffEdge) int {
				switch {
				case a.Timestamp.After(b.Timestamp):
					return -1
				case a.Timestamp.Before(b.Timestamp):
					return 1
				case a.From < b.From:
					return -1
				case a.From > b.From:
					return 1
				case a.To < b.To:
					return -1
				case a.To > b.To:
					return 1
				default:
					return 0
				}
			})
			view.ActiveOwner = view.HandoffEdges[0].To
		}

		if view.ActiveOwner == "" {
			view.ActiveOwner = latestAttributedOwner
		}

		if len(participantSet) > 0 {
			names := make([]string, 0, len(participantSet))
			for name := range participantSet {
				names = append(names, name)
			}
			slices.Sort(names)
			view.Participants = make([]ParticipantView, 0, len(names))
			for _, name := range names {
				view.Participants = append(view.Participants, ParticipantView{Name: name})
			}
		}

		views = append(views, view)
	}

	slices.SortFunc(views, func(a, b CollaborationView) int {
		switch {
		case a.UpdatedAt.After(b.UpdatedAt):
			return -1
		case a.UpdatedAt.Before(b.UpdatedAt):
			return 1
		case a.MissionID < b.MissionID:
			return -1
		case a.MissionID > b.MissionID:
			return 1
		default:
			return 0
		}
	})
	return views
}

func attributed(executionRefs map[string]struct{}, executionRef string) bool {
	executionRef = strings.TrimSpace(executionRef)
	if executionRef == "" {
		return false
	}
	_, ok := executionRefs[executionRef]
	return ok
}

func addParticipant(participants map[string]struct{}, name string) {
	if name == "" {
		return
	}
	participants[name] = struct{}{}
}

func higherPriority(candidate, current CollaborationState) bool {
	return collaborationPriority(candidate) < collaborationPriority(current)
}

func collaborationPriority(state CollaborationState) int {
	switch state {
	case CollaborationStateBlockedOnApproval:
		return 1
	case CollaborationStateWaitingOnTeammate:
		return 2
	case CollaborationStateRecovering:
		return 3
	case CollaborationStateReviewing:
		return 4
	case CollaborationStateDelegating:
		return 5
	case CollaborationStateSolo:
		return 6
	default:
		return 99
	}
}
