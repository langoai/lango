package app

import (
	"context"

	"github.com/langoai/lango/internal/appinit"
	bgpkg "github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/bootstrap"
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
	service  *mission.Service
	requests []string
}

func (h *missionApprovalHooks) OnApprovalRequested(_ context.Context, req toolchainpkg.ApprovalRequest) {
	if h == nil {
		return
	}
	h.requests = append(h.requests, req.Request.ToolName)
}

func (h *missionApprovalHooks) OnApprovalResolved(_ context.Context, _ toolchainpkg.ApprovalRequest, _ toolchainpkg.ApprovalResolution) {
}

func (h *missionApprovalHooks) MissionService() *mission.Service {
	if h == nil {
		return nil
	}
	return h.service
}

type missionBackgroundLinkHooks struct {
	service  *mission.Service
	taskIDs  []string
	prompts  []string
	sessions []string
}

func (h *missionBackgroundLinkHooks) LinkBackgroundTask(_ context.Context, taskID string, origin bgpkg.Origin, prompt string) error {
	if h == nil {
		return nil
	}
	h.taskIDs = append(h.taskIDs, taskID)
	h.prompts = append(h.prompts, prompt)
	h.sessions = append(h.sessions, origin.Session)
	return nil
}

func (h *missionBackgroundLinkHooks) MissionService() *mission.Service {
	if h == nil {
		return nil
	}
	return h.service
}

type missionRunLedgerLinkHooks struct {
	service          *mission.Service
	runIDs           []string
	sessionKeys      []string
	originalRequests []string
	goals            []string
}

func (h *missionRunLedgerLinkHooks) LinkRun(_ context.Context, runID string, sessionKey string, originalRequest string, goal string) error {
	if h == nil {
		return nil
	}
	h.runIDs = append(h.runIDs, runID)
	h.sessionKeys = append(h.sessionKeys, sessionKey)
	h.originalRequests = append(h.originalRequests, originalRequest)
	h.goals = append(h.goals, goal)
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
