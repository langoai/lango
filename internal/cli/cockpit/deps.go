package cockpit

import (
	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/langoai/lango/internal/turnrunner"
)

// Deps holds the dependencies for the cockpit TUI.
// ApprovalProvider is NOT included — type assertion for SetTTYFallback
// is handled in cmd/lango/main.go's runCockpit().
type Deps struct {
	TurnRunner        *turnrunner.Runner
	Config            *config.Config
	SessionKey        string
	SessionStore      session.Store // optional; enables /mode to persist session mode
	ToolCatalog       *toolcatalog.Catalog
	MetricsCollector  *observability.MetricsCollector
	FeatureStatuses   *app.StatusCollector
	ConfigStore       storage.ConfigProfileStore
	ProfileName       string
	BackgroundManager *background.Manager    // optional, nil when unavailable
	EventBus          *eventbus.Bus          // optional, enables channel event subscription
	ApprovalHistory   *approval.HistoryStore // optional, approval decision history
	GrantStore        *approval.GrantStore   // optional, persistent session grants
}
