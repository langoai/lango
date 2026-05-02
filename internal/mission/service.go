package mission

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const defaultUserSourceKind = "user"

// Service coordinates durable mission lifecycle writes through the mission store.
// It intentionally keeps mission latest-state, history, and execution-link writes
// behind the store boundary.
type Service struct {
	store Store
}

// StartMissionInput creates the first durable mission row for direct user-started work.
type StartMissionInput struct {
	SessionKey  string
	Title       string
	Description string
	SourceKind  string
	SourceRef   string
	StartActive bool
}

// AcceptProposalInput creates the first durable mission row for an accepted proposal.
type AcceptProposalInput struct {
	SessionKey  string
	SourceKind  string
	SourceRef   string
	Title       string
	Description string
}

// WaitForDecisionInput stores coarse durable decision state on a mission row.
type WaitForDecisionInput struct {
	MissionID       string
	Reason          string
	ActorKind       string
	ActorRef        string
	DecisionKind    string
	DecisionSummary string
}

// BlockMissionInput stores a coarse non-decision blocker on a mission row.
type BlockMissionInput struct {
	MissionID     string
	Reason        string
	ActorKind     string
	ActorRef      string
	BlockedReason string
}

// ActivateMissionInput clears coarse blocking/decision state by returning the
// mission to active work.
type ActivateMissionInput struct {
	MissionID string
	Reason    string
	ActorKind string
	ActorRef  string
}

// AttachExecutionInput links one execution identity to an existing mission.
type AttachExecutionInput struct {
	MissionID     string
	ExecutionKind ExecutionKind
	ExecutionRef  string
	LinkRole      LinkRole
}

// RefreshMissionFromExecutionInput updates mission latest state after a linked
// execution changes status.
type RefreshMissionFromExecutionInput struct {
	ExecutionKind   ExecutionKind
	ExecutionRef    string
	ToStatus        Status
	Reason          string
	ActorKind       string
	ActorRef        string
	BlockedReason   string
	DecisionKind    string
	DecisionSummary string
	CompletedAt     *time.Time
}

// NewService creates a mission lifecycle service backed by the landed store contract.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// StartMission creates the first durable mission row for direct user-started work.
func (s *Service) StartMission(ctx context.Context, in StartMissionInput) (*Mission, error) {
	store, err := s.requireStore("start mission")
	if err != nil {
		return nil, err
	}

	sourceKind := strings.TrimSpace(in.SourceKind)
	if sourceKind == "" {
		sourceKind = defaultUserSourceKind
	}

	status := StatusPrepared
	if in.StartActive {
		status = StatusActive
	}

	return store.CreateMission(ctx, CreateMissionInput{
		SessionKey:       strings.TrimSpace(in.SessionKey),
		Title:            strings.TrimSpace(in.Title),
		Description:      strings.TrimSpace(in.Description),
		Status:           status,
		SourceKind:       sourceKind,
		SourceRef:        strings.TrimSpace(in.SourceRef),
		InitialReason:    "mission created",
		InitialActorKind: "user",
	})
}

// AcceptProposal creates the first durable mission row for an accepted transient proposal.
func (s *Service) AcceptProposal(ctx context.Context, in AcceptProposalInput) (*Mission, error) {
	store, err := s.requireStore("accept proposal")
	if err != nil {
		return nil, err
	}

	sourceKind := strings.TrimSpace(in.SourceKind)
	if sourceKind == "" {
		return nil, fmt.Errorf("accept proposal: source_kind is required")
	}

	return store.CreateMission(ctx, CreateMissionInput{
		SessionKey:       strings.TrimSpace(in.SessionKey),
		Title:            strings.TrimSpace(in.Title),
		Description:      strings.TrimSpace(in.Description),
		Status:           StatusPrepared,
		SourceKind:       sourceKind,
		SourceRef:        strings.TrimSpace(in.SourceRef),
		InitialReason:    "proposal accepted",
		InitialActorKind: "user",
	})
}

// MarkWaitingDecision stores coarse durable decision-needed state for a mission.
func (s *Service) MarkWaitingDecision(ctx context.Context, in WaitForDecisionInput) (*Mission, error) {
	return s.transition(ctx, TransitionMissionInput{
		MissionID:       strings.TrimSpace(in.MissionID),
		ToStatus:        StatusWaitingDecision,
		Reason:          strings.TrimSpace(in.Reason),
		ActorKind:       defaultActorKind(in.ActorKind),
		ActorRef:        strings.TrimSpace(in.ActorRef),
		DecisionKind:    strings.TrimSpace(in.DecisionKind),
		DecisionSummary: strings.TrimSpace(in.DecisionSummary),
	})
}

// MarkBlocked stores a coarse non-decision blocker for a mission.
func (s *Service) MarkBlocked(ctx context.Context, in BlockMissionInput) (*Mission, error) {
	return s.transition(ctx, TransitionMissionInput{
		MissionID:     strings.TrimSpace(in.MissionID),
		ToStatus:      StatusBlocked,
		Reason:        strings.TrimSpace(in.Reason),
		ActorKind:     defaultActorKind(in.ActorKind),
		ActorRef:      strings.TrimSpace(in.ActorRef),
		BlockedReason: strings.TrimSpace(in.BlockedReason),
	})
}

// MarkActive clears coarse decision/blocker state by moving the mission back to active.
func (s *Service) MarkActive(ctx context.Context, in ActivateMissionInput) (*Mission, error) {
	return s.transition(ctx, TransitionMissionInput{
		MissionID: strings.TrimSpace(in.MissionID),
		ToStatus:  StatusActive,
		Reason:    strings.TrimSpace(in.Reason),
		ActorKind: defaultActorKind(in.ActorKind),
		ActorRef:  strings.TrimSpace(in.ActorRef),
	})
}

// AttachExecution links a durable mission to one execution identity.
// The helper is idempotent for the same mission/execution pair.
func (s *Service) AttachExecution(ctx context.Context, in AttachExecutionInput) error {
	store, err := s.requireStore("attach execution")
	if err != nil {
		return err
	}

	executionRef := strings.TrimSpace(in.ExecutionRef)
	if executionRef == "" {
		return fmt.Errorf("attach execution: execution_ref is required")
	}

	existing, err := store.FindExecutionLinkByExecution(ctx, in.ExecutionKind, executionRef)
	if err != nil {
		return err
	}
	if existing != nil {
		if existing.MissionID.String() == strings.TrimSpace(in.MissionID) {
			return nil
		}
		return fmt.Errorf(
			"attach execution: execution %q/%q already linked to mission %q",
			in.ExecutionKind,
			executionRef,
			existing.MissionID.String(),
		)
	}

	return store.AppendExecutionLink(ctx, AppendExecutionLinkInput{
		MissionID:     strings.TrimSpace(in.MissionID),
		ExecutionKind: in.ExecutionKind,
		ExecutionRef:  executionRef,
		LinkRole:      in.LinkRole,
	})
}

// ResolveMissionByExecution returns the durable mission linked to one execution identity.
func (s *Service) ResolveMissionByExecution(ctx context.Context, executionKind ExecutionKind, executionRef string) (*Mission, error) {
	store, err := s.requireStore("resolve mission by execution")
	if err != nil {
		return nil, err
	}
	return store.FindMissionByExecution(ctx, executionKind, strings.TrimSpace(executionRef))
}

// RefreshMissionFromExecution updates mission latest state after a linked execution changes.
// When no mission is linked yet, it returns nil without error so execution truth can remain authoritative.
func (s *Service) RefreshMissionFromExecution(ctx context.Context, in RefreshMissionFromExecutionInput) (*Mission, error) {
	missionRow, err := s.ResolveMissionByExecution(ctx, in.ExecutionKind, in.ExecutionRef)
	if err != nil || missionRow == nil {
		return missionRow, err
	}
	return s.transition(ctx, TransitionMissionInput{
		MissionID:       missionRow.ID.String(),
		ToStatus:        in.ToStatus,
		Reason:          strings.TrimSpace(in.Reason),
		ActorKind:       defaultActorKind(in.ActorKind),
		ActorRef:        strings.TrimSpace(in.ActorRef),
		ExecutionKind:   string(in.ExecutionKind),
		ExecutionRef:    strings.TrimSpace(in.ExecutionRef),
		DecisionKind:    strings.TrimSpace(in.DecisionKind),
		DecisionSummary: strings.TrimSpace(in.DecisionSummary),
		BlockedReason:   strings.TrimSpace(in.BlockedReason),
		CompletedAt:     in.CompletedAt,
	})
}

func (s *Service) transition(ctx context.Context, in TransitionMissionInput) (*Mission, error) {
	store, err := s.requireStore("transition mission")
	if err != nil {
		return nil, err
	}
	return store.TransitionMission(ctx, in)
}

func (s *Service) requireStore(action string) (Store, error) {
	if s == nil || s.store == nil {
		return nil, fmt.Errorf("%s: mission store is required", action)
	}
	return s.store, nil
}

func defaultActorKind(actorKind string) string {
	if trimmed := strings.TrimSpace(actorKind); trimmed != "" {
		return trimmed
	}
	return "system"
}
