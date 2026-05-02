package app

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/approval"
	bgpkg "github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/mission"
	runledgerpkg "github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/storage"
	toolchainpkg "github.com/langoai/lango/internal/toolchain"
)

// missionValues holds the outputs of the mission persistence module.
type missionValues struct {
	store            mission.Store
	service          *mission.Service
	approvalObserver toolchainpkg.ApprovalObserver
	backgroundLinker bgpkg.MissionExecutionLinker
	runLedgerLinker  runledgerpkg.MissionExecutionLinker
}

type missionApprovalHooks struct {
	mu       sync.Mutex
	service  *mission.Service
	requests []string
}

func (h *missionApprovalHooks) OnApprovalRequested(ctx context.Context, req toolchainpkg.ApprovalRequest) {
	if h == nil {
		return
	}
	h.mu.Lock()
	h.requests = append(h.requests, req.Request.ToolName)
	h.mu.Unlock()
	missionID := h.resolveMissionID(ctx, req.Request)
	if missionID == "" || h.service == nil {
		return
	}
	if _, err := h.service.MarkWaitingDecision(ctx, mission.WaitForDecisionInput{
		MissionID:       missionID,
		Reason:          "approval requested",
		ActorKind:       "system",
		ActorRef:        "approval-middleware",
		DecisionKind:    approvalDecisionKind(req.Request),
		DecisionSummary: waitingDecisionSummary(req.Request, ""),
	}); err != nil {
		logger().Warnw("mark mission waiting_decision", "mission_id", missionID, "error", err)
	}
}

func (h *missionApprovalHooks) OnApprovalResolved(ctx context.Context, req toolchainpkg.ApprovalRequest, resolution toolchainpkg.ApprovalResolution) {
	if h == nil || h.service == nil {
		return
	}
	missionID := h.resolveMissionID(ctx, req.Request)
	if missionID == "" {
		return
	}
	switch {
	case resolution.Err == nil && resolution.Response.Approved:
		if _, err := h.service.MarkActive(ctx, mission.ActivateMissionInput{
			MissionID: missionID,
			Reason:    "approval granted",
			ActorKind: "system",
			ActorRef:  approvalResolutionActorRef(resolution),
		}); err != nil && !isInvalidMissionTransition(err) {
			logger().Warnw("mark mission active after approval grant", "mission_id", missionID, "error", err)
		}
	case resolution.Err != nil:
		h.rewriteWaitingDecision(ctx, missionID, req.Request, resolution, waitingDecisionSummary(req.Request, approvalResolutionLabel(resolution)))
	case !resolution.Response.Approved:
		h.rewriteWaitingDecision(ctx, missionID, req.Request, resolution, waitingDecisionSummary(req.Request, approvalResolutionLabel(resolution)))
	}
}

func (h *missionApprovalHooks) MissionService() *mission.Service {
	if h == nil {
		return nil
	}
	return h.service
}

type missionBackgroundLinkHooks struct {
	mu       sync.Mutex
	service  *mission.Service
	taskIDs  []string
	prompts  []string
	sessions []string
}

func (h *missionBackgroundLinkHooks) LinkBackgroundTask(ctx context.Context, taskID string, origin bgpkg.Origin, prompt string) error {
	if h == nil {
		return nil
	}
	h.mu.Lock()
	h.taskIDs = append(h.taskIDs, taskID)
	h.prompts = append(h.prompts, prompt)
	h.sessions = append(h.sessions, origin.Session)
	h.mu.Unlock()
	if h.service == nil {
		return nil
	}
	missionID := strings.TrimSpace(ctxkeys.MissionIDFromContext(ctx))
	if missionID == "" {
		return nil
	}
	if err := h.service.AttachExecution(ctx, mission.AttachExecutionInput{
		MissionID:     missionID,
		ExecutionKind: mission.ExecutionKindTaskOSExecution,
		ExecutionRef:  taskID,
		LinkRole:      mission.LinkRolePrimary,
	}); err != nil {
		logger().Warnw("attach background task to mission", "mission_id", missionID, "task_id", taskID, "error", err)
		return err
	}
	return nil
}

func (h *missionBackgroundLinkHooks) MissionService() *mission.Service {
	if h == nil {
		return nil
	}
	return h.service
}

type missionRunLedgerLinkHooks struct {
	mu               sync.Mutex
	service          *mission.Service
	runIDs           []string
	sessionKeys      []string
	originalRequests []string
	goals            []string
}

func (h *missionRunLedgerLinkHooks) LinkRun(ctx context.Context, runID string, sessionKey string, originalRequest string, goal string) error {
	if h == nil {
		return nil
	}
	h.mu.Lock()
	h.runIDs = append(h.runIDs, runID)
	h.sessionKeys = append(h.sessionKeys, sessionKey)
	h.originalRequests = append(h.originalRequests, originalRequest)
	h.goals = append(h.goals, goal)
	h.mu.Unlock()
	if h.service == nil {
		return nil
	}
	missionID := strings.TrimSpace(ctxkeys.MissionIDFromContext(ctx))
	if missionID == "" {
		return nil
	}
	if err := h.service.AttachExecution(ctx, mission.AttachExecutionInput{
		MissionID:     missionID,
		ExecutionKind: mission.ExecutionKindRunLedgerRun,
		ExecutionRef:  runID,
		LinkRole:      mission.LinkRolePrimary,
	}); err != nil {
		logger().Warnw("attach runledger run to mission", "mission_id", missionID, "run_id", runID, "error", err)
		return err
	}
	return nil
}

func (h *missionRunLedgerLinkHooks) MissionService() *mission.Service {
	if h == nil {
		return nil
	}
	return h.service
}

// missionModule wires durable mission storage and service through the app boundary.
type missionModule struct {
	boot *bootstrap.Result
}

func (m *missionModule) Name() string { return "mission" }

func (m *missionModule) Provides() []appinit.Provides {
	return []appinit.Provides{appinit.ProvidesMission}
}

func (m *missionModule) DependsOn() []appinit.Provides { return nil }

func (m *missionModule) Enabled() bool {
	if m == nil || m.boot == nil || m.boot.Storage == nil {
		return false
	}
	_, ok := storage.ResolveEntBacked(m.boot.Storage, mission.NewEntStore)
	return ok
}

func (m *missionModule) Init(_ context.Context, _ appinit.Resolver) (*appinit.ModuleResult, error) {
	store, ok := storage.ResolveEntBacked(m.boot.Storage, mission.NewEntStore)
	if !ok || store == nil {
		return &appinit.ModuleResult{}, nil
	}

	service := mission.NewService(store)
	values := &missionValues{
		store:            store,
		service:          service,
		approvalObserver: &missionApprovalHooks{service: service},
		backgroundLinker: &missionBackgroundLinkHooks{service: service},
		runLedgerLinker:  &missionRunLedgerLinkHooks{service: service},
	}

	return &appinit.ModuleResult{
		Values: map[appinit.Provides]interface{}{
			appinit.ProvidesMission: values,
		},
	}, nil
}

func (h *missionApprovalHooks) resolveMissionID(ctx context.Context, req approval.ApprovalRequest) string {
	if h == nil || h.service == nil {
		return ""
	}
	if missionID := strings.TrimSpace(req.MissionID); missionID != "" {
		return missionID
	}
	if missionID := strings.TrimSpace(ctxkeys.MissionIDFromContext(ctx)); missionID != "" {
		return missionID
	}
	executionKind := strings.TrimSpace(req.ExecutionKind)
	executionRef := strings.TrimSpace(req.ExecutionRef)
	if executionKind == "" || executionRef == "" {
		return ""
	}
	row, err := h.service.ResolveMissionByExecution(ctx, mission.ExecutionKind(executionKind), executionRef)
	if err != nil {
		logger().Warnw("resolve mission by execution for approval", "execution_kind", executionKind, "execution_ref", executionRef, "error", err)
		return ""
	}
	if row == nil {
		return ""
	}
	return row.ID.String()
}

func (h *missionApprovalHooks) rewriteWaitingDecision(
	ctx context.Context,
	missionID string,
	req approval.ApprovalRequest,
	resolution toolchainpkg.ApprovalResolution,
	summary string,
) {
	if _, err := h.service.MarkWaitingDecision(ctx, mission.WaitForDecisionInput{
		MissionID:       missionID,
		Reason:          "approval remains unresolved",
		ActorKind:       "system",
		ActorRef:        approvalResolutionActorRef(resolution),
		DecisionKind:    approvalDecisionKind(req),
		DecisionSummary: summary,
	}); err != nil {
		logger().Warnw("rewrite mission waiting_decision", "mission_id", missionID, "error", err)
	}
}

func approvalDecisionKind(req approval.ApprovalRequest) string {
	return "tool_approval"
}

func waitingDecisionSummary(req approval.ApprovalRequest, resolutionLabel string) string {
	base := strings.TrimSpace(req.Summary)
	if base == "" {
		base = strings.TrimSpace(req.ToolName)
	}
	if resolutionLabel == "" {
		return base
	}
	return base + " (" + resolutionLabel + ")"
}

func approvalResolutionLabel(resolution toolchainpkg.ApprovalResolution) string {
	switch {
	case resolution.Err == nil && resolution.Response.Approved:
		return "approved"
	case resolution.Err == nil:
		return "denied"
	case errors.Is(resolution.Err, approval.ErrTimeout):
		return "timed out"
	case errors.Is(resolution.Err, approval.ErrUnavailable):
		return "unavailable"
	default:
		return "denied"
	}
}

func approvalResolutionActorRef(resolution toolchainpkg.ApprovalResolution) string {
	if provider := strings.TrimSpace(resolution.Response.Provider); provider != "" {
		return provider
	}
	if provider := strings.TrimSpace(approval.ProviderFromError(resolution.Err)); provider != "" {
		return provider
	}
	return "approval-middleware"
}

func isInvalidMissionTransition(err error) bool {
	return err != nil && strings.Contains(err.Error(), "invalid transition")
}
