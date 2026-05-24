package toolchain

import (
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/toolcatalog"
)

// BuildHookRegistry constructs the tool execution hook registry from config.
// Pass nil for bus when running outside a full app (e.g. CLI snapshot mode);
// EventBus hooks will be omitted. Pass nil for knowledgeSaver when the hook
// only needs to be inspected (snapshot), not executed. When catalog is non-nil,
// SaveableTools is derived from tool metadata; otherwise falls back to the
// hard-coded DefaultSaveableTools constant.
func BuildHookRegistry(cfg *config.Config, bus *eventbus.Bus, knowledgeSaver KnowledgeSaver, catalog *toolcatalog.Catalog) *HookRegistry {
	hookRegistry := NewHookRegistry()
	hookRegistry.RegisterPre(NewSecurityFilterHook(cfg.Hooks.BlockedCommands))
	if cfg.Hooks.AccessControl {
		hookRegistry.RegisterPre(NewAgentAccessControlHook(nil))
	}
	if (cfg.Hooks.Enabled || cfg.Agent.MultiAgent) && cfg.Hooks.EventPublishing && bus != nil {
		ebHook := NewEventBusHook(bus)
		hookRegistry.RegisterPre(ebHook)
		hookRegistry.RegisterPost(ebHook)
	}
	if cfg.Hooks.KnowledgeSave {
		saveableTools := DefaultSaveableTools
		if catalog != nil {
			if derived := catalog.SaveableToolNames(); len(derived) > 0 {
				saveableTools = derived
			}
		}
		hookRegistry.RegisterPost(NewKnowledgeSaveHook(knowledgeSaver, saveableTools))
	}
	return hookRegistry
}
