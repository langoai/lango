package cockpit

import (
	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/langoai/lango/internal/turnrunner"
)

// Deps holds the dependencies for the cockpit TUI.
// ApprovalProvider is NOT included — type assertion for SetTTYFallback
// is handled in cmd/lango/main.go's runCockpit().
type Deps struct {
	TurnRunner       *turnrunner.Runner
	Config           *config.Config
	SessionKey       string
	ToolCatalog      *toolcatalog.Catalog
	MetricsCollector *observability.MetricsCollector
	FeatureStatuses  *app.StatusCollector
	ConfigStore      *configstore.Store
	ProfileName      string
}
