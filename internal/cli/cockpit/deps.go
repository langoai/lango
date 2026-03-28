package cockpit

import (
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/turnrunner"
)

// Deps holds the dependencies for the cockpit TUI.
// ApprovalProvider is NOT included — type assertion for SetTTYFallback
// is handled in cmd/lango/main.go's runCockpit().
type Deps struct {
	TurnRunner *turnrunner.Runner
	Config     *config.Config
	SessionKey string
}
