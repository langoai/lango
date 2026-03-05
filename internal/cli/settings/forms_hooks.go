package settings

import (
	"strings"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

// NewHooksForm creates the Hooks configuration form.
func NewHooksForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Hooks Configuration")

	form.AddField(&tuicore.Field{
		Key: "hooks_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Hooks.Enabled,
		Description: "Enable the hook system for tool execution interception",
	})

	form.AddField(&tuicore.Field{
		Key: "hooks_security_filter", Label: "Security Filter", Type: tuicore.InputBool,
		Checked:     cfg.Hooks.SecurityFilter,
		Description: "Enable security filter hook to block dangerous commands",
	})

	form.AddField(&tuicore.Field{
		Key: "hooks_access_control", Label: "Access Control", Type: tuicore.InputBool,
		Checked:     cfg.Hooks.AccessControl,
		Description: "Enable per-agent tool access control hook",
	})

	form.AddField(&tuicore.Field{
		Key: "hooks_event_publishing", Label: "Event Publishing", Type: tuicore.InputBool,
		Checked:     cfg.Hooks.EventPublishing,
		Description: "Enable publishing tool execution events to the event bus",
	})

	form.AddField(&tuicore.Field{
		Key: "hooks_knowledge_save", Label: "Knowledge Save", Type: tuicore.InputBool,
		Checked:     cfg.Hooks.KnowledgeSave,
		Description: "Enable automatic knowledge saving from tool results",
	})

	form.AddField(&tuicore.Field{
		Key: "hooks_blocked_commands", Label: "Blocked Commands", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Hooks.BlockedCommands, ","),
		Placeholder: "rm -rf,shutdown (comma-separated)",
		Description: "Command patterns to block via the security filter hook",
	})

	return &form
}

// NewAgentMemoryForm creates the Agent Memory configuration form.
func NewAgentMemoryForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Agent Memory Configuration")

	form.AddField(&tuicore.Field{
		Key: "agent_memory_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.AgentMemory.Enabled,
		Description: "Enable per-agent persistent memory across sessions",
	})

	return &form
}
