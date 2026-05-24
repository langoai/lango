package app

import (
	"context"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/langoai/lango/internal/toolchain"
)

// buildModeAllowlistResolver returns a resolver that maps the active session
// mode (from context) to its expanded tool allowlist. When no mode is active,
// the resolver returns (nil, false) so the middleware passes through.
//
// The resolver expands @category references via the catalog at call time,
// so changes to category membership take effect without restart.
func buildModeAllowlistResolver(cfg *config.Config, catalog *toolcatalog.Catalog) toolchain.ModeAllowlistResolver {
	return func(ctx context.Context) (map[string]bool, bool) {
		modeName := session.ModeNameFromContext(ctx)
		if modeName == "" {
			return nil, false
		}
		if cfg == nil || catalog == nil {
			return nil, false
		}
		mode, ok := cfg.LookupMode(modeName)
		if !ok || len(mode.Tools) == 0 {
			return nil, false
		}
		return catalog.ResolveModeAllowlist(mode.Tools), true
	}
}
