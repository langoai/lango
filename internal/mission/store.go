package mission

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/langoai/lango/internal/ent"
	entmission "github.com/langoai/lango/internal/ent/mission"
	entmissionexecutionlink "github.com/langoai/lango/internal/ent/missionexecutionlink"
	entmissionstatehistory "github.com/langoai/lango/internal/ent/missionstatehistory"
)

// Mission is the durable mission latest-state row.
type Mission = ent.Mission

// ExecutionLink is a durable mission-to-execution link row.
type ExecutionLink = ent.MissionExecutionLink

// Status is the durable mission status enum.
type Status = entmission.Status

// ExecutionKind is the execution-link identity enum.
type ExecutionKind = entmissionexecutionlink.ExecutionKind

// LinkRole is the execution-link role enum.
type LinkRole = entmissionexecutionlink.LinkRole

const (
	StatusPrepared        = entmission.StatusPrepared
	StatusActive          = entmission.StatusActive
	StatusWaitingDecision = entmission.StatusWaitingDecision
	StatusBlocked         = entmission.StatusBlocked
	StatusDone            = entmission.StatusDone
	StatusCancelled       = entmission.StatusCancelled

	ExecutionKindRunLedgerRun    = entmissionexecutionlink.ExecutionKindRunledgerRun
	ExecutionKindTaskOSExecution = entmissionexecutionlink.ExecutionKindTaskOsExecution

	LinkRolePrimary  = entmissionexecutionlink.LinkRolePrimary
	LinkRoleFollowup = entmissionexecutionlink.LinkRoleFollowup
	LinkRoleRetry    = entmissionexecutionlink.LinkRoleRetry
	LinkRoleResearch = entmissionexecutionlink.LinkRoleResearch
	LinkRoleDraft    = entmissionexecutionlink.LinkRoleDraft
	LinkRoleHandoff  = entmissionexecutionlink.LinkRoleHandoff
)

// Store persists durable mission state and execution links.
type Store interface {
	CreateMission(ctx context.Context, in CreateMissionInput) (*Mission, error)
	GetMission(ctx context.Context, missionID string) (*Mission, error)
	ListMissionsBySession(ctx context.Context, sessionKey string, limit int) ([]*Mission, error)
	TransitionMission(ctx context.Context, in TransitionMissionInput) (*Mission, error)
	AppendExecutionLink(ctx context.Context, in AppendExecutionLinkInput) error
	ListExecutionLinks(ctx context.Context, missionID string) ([]*ExecutionLink, error)
	FindExecutionLinkByExecution(ctx context.Context, executionKind ExecutionKind, executionRef string) (*ExecutionLink, error)
	FindMissionByExecution(ctx context.Context, executionKind ExecutionKind, executionRef string) (*Mission, error)
}

// CreateMissionInput describes one durable mission latest-state row.
type CreateMissionInput struct {
	SessionKey             string
	Title                  string
	Description            string
	Status                 Status
	SourceKind             string
	SourceRef              string
	CurrentBlockedReason   string
	CurrentDecisionKind    string
	CurrentDecisionSummary string
	CompletedAt            *time.Time
	InitialReason          string
	InitialActorKind       string
	InitialActorRef        string
	InitialExecutionKind   string
	InitialExecutionRef    string
	InitialDecisionKind    string
	InitialDecisionSummary string
	InitialPayload         map[string]any
}

// TransitionMissionInput describes a single durable mission state transition.
type TransitionMissionInput struct {
	MissionID       string
	ToStatus        Status
	Reason          string
	ActorKind       string
	ActorRef        string
	ExecutionKind   string
	ExecutionRef    string
	DecisionKind    string
	DecisionSummary string
	BlockedReason   string
	Payload         map[string]any
	CompletedAt     *time.Time
}

// AppendExecutionLinkInput appends one durable mission-to-execution link.
type AppendExecutionLinkInput struct {
	MissionID     string
	ExecutionKind ExecutionKind
	ExecutionRef  string
	LinkRole      LinkRole
}

// EntStore implements Store using the shared ent client.
type EntStore struct {
	client *ent.Client
}

// NewEntStore creates a new mission store backed by ent.
func NewEntStore(client *ent.Client) *EntStore {
	return &EntStore{client: client}
}

func (s *EntStore) CreateMission(ctx context.Context, in CreateMissionInput) (*Mission, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("create mission: store unavailable")
	}
	sessionKey := strings.TrimSpace(in.SessionKey)
	title := strings.TrimSpace(in.Title)
	sourceKind := strings.TrimSpace(in.SourceKind)
	if sessionKey == "" {
		return nil, fmt.Errorf("create mission: session_key is required")
	}
	if title == "" {
		return nil, fmt.Errorf("create mission: title is required")
	}
	if sourceKind == "" {
		return nil, fmt.Errorf("create mission: source_kind is required")
	}
	if in.Status != "" {
		if err := entmission.StatusValidator(in.Status); err != nil {
			return nil, fmt.Errorf("create mission: %w", err)
		}
	}
	latest := normalizeLatestState(
		in.Status,
		in.CurrentBlockedReason,
		in.CurrentDecisionKind,
		in.CurrentDecisionSummary,
		in.CompletedAt,
		nil,
		time.Now(),
	)
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("create mission: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	builder := tx.Mission.Create().
		SetSessionKey(sessionKey).
		SetTitle(title).
		SetSourceKind(sourceKind).
		SetStatus(latest.status)

	if description := strings.TrimSpace(in.Description); description != "" {
		builder.SetDescription(description)
	}
	if sourceRef := strings.TrimSpace(in.SourceRef); sourceRef != "" {
		builder.SetSourceRef(sourceRef)
	}
	if latest.blockedReason != nil {
		builder.SetCurrentBlockedReason(*latest.blockedReason)
	}
	if latest.decisionKind != nil {
		builder.SetCurrentDecisionKind(*latest.decisionKind)
	}
	if latest.decisionSummary != nil {
		builder.SetCurrentDecisionSummary(*latest.decisionSummary)
	}
	if latest.completedAt != nil {
		builder.SetCompletedAt(*latest.completedAt)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create mission: %w", err)
	}

	if actorKind := strings.TrimSpace(in.InitialActorKind); actorKind != "" {
		historyBuilder := tx.MissionStateHistory.Create().
			SetMissionID(row.ID).
			SetSeq(1).
			SetToStatus(toHistoryStatus(latest.status)).
			SetActorKind(actorKind)
		if reason := strings.TrimSpace(in.InitialReason); reason != "" {
			historyBuilder.SetReason(reason)
		}
		if actorRef := strings.TrimSpace(in.InitialActorRef); actorRef != "" {
			historyBuilder.SetActorRef(actorRef)
		}
		if executionKind := strings.TrimSpace(in.InitialExecutionKind); executionKind != "" {
			historyBuilder.SetExecutionKind(executionKind)
		}
		if executionRef := strings.TrimSpace(in.InitialExecutionRef); executionRef != "" {
			historyBuilder.SetExecutionRef(executionRef)
		}
		if decisionKind := strings.TrimSpace(in.InitialDecisionKind); decisionKind != "" {
			historyBuilder.SetDecisionKind(decisionKind)
		}
		if decisionSummary := strings.TrimSpace(in.InitialDecisionSummary); decisionSummary != "" {
			historyBuilder.SetDecisionSummary(decisionSummary)
		}
		if len(in.InitialPayload) > 0 {
			historyBuilder.SetPayload(in.InitialPayload)
		}
		if _, err := historyBuilder.Save(ctx); err != nil {
			return nil, fmt.Errorf("create mission: append initial history: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("create mission: commit: %w", err)
	}
	return row.Unwrap(), nil
}

func (s *EntStore) GetMission(ctx context.Context, missionID string) (*Mission, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("get mission: store unavailable")
	}
	id, err := parseMissionID(missionID)
	if err != nil {
		return nil, err
	}
	row, err := s.client.Mission.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get mission %q: %w", missionID, err)
	}
	return row, nil
}

func (s *EntStore) ListMissionsBySession(ctx context.Context, sessionKey string, limit int) ([]*Mission, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("list missions: store unavailable")
	}
	sessionKey = strings.TrimSpace(sessionKey)
	if sessionKey == "" {
		return nil, fmt.Errorf("list missions: session_key is required")
	}

	query := s.client.Mission.Query().
		Where(entmission.SessionKey(sessionKey)).
		Order(
			entmission.ByUpdatedAt(sql.OrderDesc()),
			entmission.ByCreatedAt(sql.OrderDesc()),
		)
	if limit > 0 {
		query = query.Limit(limit)
	}

	rows, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list missions for session %q: %w", sessionKey, err)
	}
	return rows, nil
}

func (s *EntStore) TransitionMission(ctx context.Context, in TransitionMissionInput) (_ *Mission, err error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("transition mission: store unavailable")
	}
	missionID, err := parseMissionID(in.MissionID)
	if err != nil {
		return nil, err
	}
	if in.ToStatus == "" {
		return nil, fmt.Errorf("transition mission: to_status is required")
	}
	if err := entmission.StatusValidator(in.ToStatus); err != nil {
		return nil, fmt.Errorf("transition mission: %w", err)
	}
	actorKind := strings.TrimSpace(in.ActorKind)
	if actorKind == "" {
		return nil, fmt.Errorf("transition mission: actor_kind is required")
	}

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("transition mission: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	row, err := tx.Mission.Get(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("transition mission %q: %w", in.MissionID, err)
	}
	if !isAllowedMissionTransition(row.Status, in.ToStatus) {
		return nil, fmt.Errorf("transition mission %q: invalid transition %q -> %q", in.MissionID, row.Status, in.ToStatus)
	}

	nextSeq, err := nextMissionHistorySeq(ctx, tx, missionID)
	if err != nil {
		return nil, fmt.Errorf("transition mission %q: %w", in.MissionID, err)
	}

	historyBuilder := tx.MissionStateHistory.Create().
		SetMissionID(missionID).
		SetSeq(nextSeq).
		SetToStatus(toHistoryStatus(in.ToStatus)).
		SetActorKind(actorKind)
	if fromStatus, ok := toHistoryFromStatus(row.Status); ok {
		historyBuilder.SetFromStatus(fromStatus)
	}
	if reason := strings.TrimSpace(in.Reason); reason != "" {
		historyBuilder.SetReason(reason)
	}
	if actorRef := strings.TrimSpace(in.ActorRef); actorRef != "" {
		historyBuilder.SetActorRef(actorRef)
	}
	if executionKind := strings.TrimSpace(in.ExecutionKind); executionKind != "" {
		historyBuilder.SetExecutionKind(executionKind)
	}
	if executionRef := strings.TrimSpace(in.ExecutionRef); executionRef != "" {
		historyBuilder.SetExecutionRef(executionRef)
	}
	if decisionKind := strings.TrimSpace(in.DecisionKind); decisionKind != "" {
		historyBuilder.SetDecisionKind(decisionKind)
	}
	if decisionSummary := strings.TrimSpace(in.DecisionSummary); decisionSummary != "" {
		historyBuilder.SetDecisionSummary(decisionSummary)
	}
	if len(in.Payload) > 0 {
		historyBuilder.SetPayload(in.Payload)
	}
	if _, err := historyBuilder.Save(ctx); err != nil {
		return nil, fmt.Errorf("transition mission %q: append history: %w", in.MissionID, err)
	}

	latest := normalizeLatestState(
		in.ToStatus,
		in.BlockedReason,
		in.DecisionKind,
		in.DecisionSummary,
		in.CompletedAt,
		row.CompletedAt,
		time.Now(),
	)
	update := tx.Mission.UpdateOneID(missionID).SetStatus(latest.status)
	if latest.blockedReason != nil {
		update.SetCurrentBlockedReason(*latest.blockedReason)
	} else {
		update.ClearCurrentBlockedReason()
	}
	if latest.decisionKind != nil {
		update.SetCurrentDecisionKind(*latest.decisionKind)
	} else {
		update.ClearCurrentDecisionKind()
	}
	if latest.decisionSummary != nil {
		update.SetCurrentDecisionSummary(*latest.decisionSummary)
	} else {
		update.ClearCurrentDecisionSummary()
	}
	if latest.completedAt != nil {
		update.SetCompletedAt(*latest.completedAt)
	} else {
		update.ClearCompletedAt()
	}

	updatedRow, err := update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("transition mission %q: update latest state: %w", in.MissionID, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("transition mission %q: commit: %w", in.MissionID, err)
	}
	return updatedRow.Unwrap(), nil
}

func (s *EntStore) AppendExecutionLink(ctx context.Context, in AppendExecutionLinkInput) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("append execution link: store unavailable")
	}
	missionID, err := parseMissionID(in.MissionID)
	if err != nil {
		return err
	}
	if err := entmissionexecutionlink.ExecutionKindValidator(in.ExecutionKind); err != nil {
		return fmt.Errorf("append execution link: %w", err)
	}
	executionRef := strings.TrimSpace(in.ExecutionRef)
	if executionRef == "" {
		return fmt.Errorf("append execution link: execution_ref is required")
	}
	linkRole := in.LinkRole
	if linkRole == "" {
		linkRole = entmissionexecutionlink.LinkRolePrimary
	}
	if err := entmissionexecutionlink.LinkRoleValidator(linkRole); err != nil {
		return fmt.Errorf("append execution link: %w", err)
	}
	if _, err := s.client.Mission.Get(ctx, missionID); err != nil {
		return fmt.Errorf("append execution link: mission %q: %w", in.MissionID, err)
	}
	_, err = s.client.MissionExecutionLink.Create().
		SetMissionID(missionID).
		SetExecutionKind(in.ExecutionKind).
		SetExecutionRef(executionRef).
		SetLinkRole(linkRole).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("append execution link: %w", err)
	}
	return nil
}

func (s *EntStore) ListExecutionLinks(ctx context.Context, missionID string) ([]*ExecutionLink, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("list execution links: store unavailable")
	}
	id, err := parseMissionID(missionID)
	if err != nil {
		return nil, err
	}
	rows, err := s.client.MissionExecutionLink.Query().
		Where(entmissionexecutionlink.MissionID(id)).
		Order(
			entmissionexecutionlink.ByCreatedAt(),
			entmissionexecutionlink.ByExecutionRef(),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list execution links for mission %q: %w", missionID, err)
	}
	return rows, nil
}

func (s *EntStore) FindExecutionLinkByExecution(ctx context.Context, executionKind ExecutionKind, executionRef string) (*ExecutionLink, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("find execution link: store unavailable")
	}
	if err := entmissionexecutionlink.ExecutionKindValidator(executionKind); err != nil {
		return nil, fmt.Errorf("find execution link: %w", err)
	}
	executionRef = strings.TrimSpace(executionRef)
	if executionRef == "" {
		return nil, fmt.Errorf("find execution link: execution_ref is required")
	}
	row, err := s.client.MissionExecutionLink.Query().
		Where(
			entmissionexecutionlink.ExecutionKindEQ(executionKind),
			entmissionexecutionlink.ExecutionRefEQ(executionRef),
		).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find execution link %q/%q: %w", executionKind, executionRef, err)
	}
	return row, nil
}

func (s *EntStore) FindMissionByExecution(ctx context.Context, executionKind ExecutionKind, executionRef string) (*Mission, error) {
	link, err := s.FindExecutionLinkByExecution(ctx, executionKind, executionRef)
	if err != nil || link == nil {
		return nil, err
	}
	row, err := s.client.Mission.Get(ctx, link.MissionID)
	if err != nil {
		return nil, fmt.Errorf("find mission by execution %q/%q: %w", executionKind, executionRef, err)
	}
	return row, nil
}

func nextMissionHistorySeq(ctx context.Context, tx *ent.Tx, missionID uuid.UUID) (int64, error) {
	row, err := tx.MissionStateHistory.Query().
		Where(entmissionstatehistory.MissionID(missionID)).
		Order(entmissionstatehistory.BySeq(sql.OrderDesc())).
		First(ctx)
	if ent.IsNotFound(err) {
		return 1, nil
	}
	if err != nil {
		return 0, fmt.Errorf("query mission history seq: %w", err)
	}
	return row.Seq + 1, nil
}

func parseMissionID(missionID string) (uuid.UUID, error) {
	trimmed := strings.TrimSpace(missionID)
	if trimmed == "" {
		return uuid.UUID{}, fmt.Errorf("mission id is required")
	}
	id, err := uuid.Parse(trimmed)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse mission id %q: %w", missionID, err)
	}
	return id, nil
}

func toHistoryFromStatus(status entmission.Status) (entmissionstatehistory.FromStatus, bool) {
	switch status {
	case entmission.StatusPrepared:
		return entmissionstatehistory.FromStatusPrepared, true
	case entmission.StatusActive:
		return entmissionstatehistory.FromStatusActive, true
	case entmission.StatusWaitingDecision:
		return entmissionstatehistory.FromStatusWaitingDecision, true
	case entmission.StatusBlocked:
		return entmissionstatehistory.FromStatusBlocked, true
	case entmission.StatusDone:
		return entmissionstatehistory.FromStatusDone, true
	case entmission.StatusCancelled:
		return entmissionstatehistory.FromStatusCancelled, true
	default:
		return "", false
	}
}

func toHistoryStatus(status entmission.Status) entmissionstatehistory.ToStatus {
	switch status {
	case entmission.StatusPrepared:
		return entmissionstatehistory.ToStatusPrepared
	case entmission.StatusActive:
		return entmissionstatehistory.ToStatusActive
	case entmission.StatusWaitingDecision:
		return entmissionstatehistory.ToStatusWaitingDecision
	case entmission.StatusBlocked:
		return entmissionstatehistory.ToStatusBlocked
	case entmission.StatusDone:
		return entmissionstatehistory.ToStatusDone
	case entmission.StatusCancelled:
		return entmissionstatehistory.ToStatusCancelled
	default:
		return entmissionstatehistory.ToStatusPrepared
	}
}

func isAllowedMissionTransition(from, to entmission.Status) bool {
	switch from {
	case entmission.StatusPrepared:
		return to == entmission.StatusActive || to == entmission.StatusWaitingDecision || to == entmission.StatusBlocked
	case entmission.StatusActive:
		return to == entmission.StatusWaitingDecision || to == entmission.StatusBlocked || to == entmission.StatusDone || to == entmission.StatusCancelled
	case entmission.StatusWaitingDecision:
		return to == entmission.StatusWaitingDecision || to == entmission.StatusActive || to == entmission.StatusBlocked || to == entmission.StatusCancelled
	case entmission.StatusBlocked:
		return to == entmission.StatusActive || to == entmission.StatusCancelled
	default:
		return false
	}
}

type latestState struct {
	status          entmission.Status
	blockedReason   *string
	decisionKind    *string
	decisionSummary *string
	completedAt     *time.Time
}

func normalizeLatestState(
	status entmission.Status,
	blockedReason string,
	decisionKind string,
	decisionSummary string,
	completedAt *time.Time,
	existingCompletedAt *time.Time,
	now time.Time,
) latestState {
	if status == "" {
		status = entmission.StatusPrepared
	}

	normalized := latestState{status: status}
	trimmedBlockedReason := trimStringPtr(blockedReason)
	trimmedDecisionKind := trimStringPtr(decisionKind)
	trimmedDecisionSummary := trimStringPtr(decisionSummary)

	switch status {
	case entmission.StatusBlocked:
		normalized.blockedReason = trimmedBlockedReason
	case entmission.StatusWaitingDecision:
		normalized.decisionKind = trimmedDecisionKind
		normalized.decisionSummary = trimmedDecisionSummary
	case entmission.StatusDone, entmission.StatusCancelled:
		switch {
		case completedAt != nil:
			normalized.completedAt = completedAt
		case existingCompletedAt != nil:
			normalized.completedAt = existingCompletedAt
		default:
			terminalAt := now
			normalized.completedAt = &terminalAt
		}
	}

	return normalized
}

func trimStringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
