package app

import (
	"context"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/mission"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/langoai/lango/internal/wallet"
)

func init() {
	storage.RegisterEntMissionStoreFactory(func(client *ent.Client) any {
		return mission.NewEntStore(client)
	})
}

// missionValues holds the outputs of the mission persistence module.
type missionValues struct {
	store            mission.Store
	service          *mission.Service
	approvalObserver missionApprovalObserver
	executionLinker  missionExecutionLinkAdapter
}

// missionApprovalObserver is a placeholder hook for approval lifecycle wiring.
type missionApprovalObserver interface {
	MissionService() *mission.Service
}

// missionExecutionLinkAdapter is a placeholder hook for execution-link wiring.
type missionExecutionLinkAdapter interface {
	MissionService() *mission.Service
}

type missionApprovalHooks struct {
	service *mission.Service
}

func (h *missionApprovalHooks) MissionService() *mission.Service {
	if h == nil {
		return nil
	}
	return h.service
}

type missionExecutionLinkHooks struct {
	service *mission.Service
}

func (h *missionExecutionLinkHooks) MissionService() *mission.Service {
	if h == nil {
		return nil
	}
	return h.service
}

var buildApprovalMiddlewareWithMission = func(
	ic config.InterceptorConfig,
	ap approval.Provider,
	gs *approval.GrantStore,
	limiter wallet.SpendingLimiter,
	history *approval.HistoryStore,
	observer missionApprovalObserver,
) toolchain.Middleware {
	return toolchain.WithApproval(ic, ap, gs, limiter, history)
}

var buildBackgroundToolsWithMission = func(
	mgr *background.Manager,
	defaultDeliverTo []string,
	linker missionExecutionLinkAdapter,
) []*agent.Tool {
	return background.BuildTools(mgr, defaultDeliverTo)
}

var buildRunLedgerToolsWithMission = func(
	store runledger.RunLedgerStore,
	pev *runledger.PEVEngine,
	linker missionExecutionLinkAdapter,
) []*agent.Tool {
	return runledger.BuildTools(store, pev)
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
	_, ok := m.boot.Storage.Mission().(mission.Store)
	return ok
}

func (m *missionModule) Init(_ context.Context, _ appinit.Resolver) (*appinit.ModuleResult, error) {
	store, ok := m.boot.Storage.Mission().(mission.Store)
	if !ok || store == nil {
		return &appinit.ModuleResult{}, nil
	}

	service := mission.NewService(store)
	values := &missionValues{
		store:            store,
		service:          service,
		approvalObserver: &missionApprovalHooks{service: service},
		executionLinker:  &missionExecutionLinkHooks{service: service},
	}

	return &appinit.ModuleResult{
		Values: map[appinit.Provides]interface{}{
			appinit.ProvidesMission: values,
		},
	}, nil
}
