package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/toolchain"
)

func newHooksCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "hooks",
		Short: "Show active hook configuration and registry snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			h := cfg.Hooks

			registry := app.BuildHookRegistry(cfg, nil, nil)

			if jsonOutput {
				return printJSON(h, registry)
			}

			return printText(h, registry)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

type hooksConfigOutput struct {
	Enabled         bool     `json:"enabled"`
	SecurityFilter  bool     `json:"security_filter"`
	AccessControl   bool     `json:"access_control"`
	EventPublishing bool     `json:"event_publishing"`
	KnowledgeSave   bool     `json:"knowledge_save"`
	BlockedCommands []string `json:"blocked_commands,omitempty"`
}

type hookInfo struct {
	Name     string         `json:"name"`
	Priority int            `json:"priority"`
	Phase    string         `json:"phase"`
	Wirable  bool           `json:"wirable"`
	Reason   string         `json:"reason,omitempty"`
	Details  map[string]any `json:"details,omitempty"`
}

type registryOutput struct {
	PreHooks  []hookInfo `json:"preHooks"`
	PostHooks []hookInfo `json:"postHooks"`
}

type fullOutput struct {
	hooksConfigOutput
	Registry registryOutput `json:"registry"`
}

func buildRegistryOutput(registry *toolchain.HookRegistry, cfg config.HooksConfig) registryOutput {
	var out registryOutput

	for _, h := range registry.PreHooks() {
		out.PreHooks = append(out.PreHooks, hookInfo{
			Name:     h.Name(),
			Priority: h.Priority(),
			Phase:    "pre",
			Wirable:  true,
		})
	}

	for _, h := range registry.PostHooks() {
		info := hookInfo{
			Name:     h.Name(),
			Priority: h.Priority(),
			Phase:    "post",
			Wirable:  true,
		}
		if kh, ok := h.(*toolchain.KnowledgeSaveHook); ok {
			tools := make([]string, 0, len(kh.SaveableTools))
			for t := range kh.SaveableTools {
				tools = append(tools, t)
			}
			info.Details = map[string]any{"saveableTools": tools}
		}
		out.PostHooks = append(out.PostHooks, info)
	}

	if cfg.EventPublishing {
		hasEB := false
		for _, h := range registry.PreHooks() {
			if h.Name() == "eventbus" {
				hasEB = true
				break
			}
		}
		if !hasEB {
			placeholder := hookInfo{
				Name:    "eventbus",
				Phase:   "pre+post",
				Wirable: false,
				Reason:  "requires a running event bus (unavailable in CLI mode)",
			}
			out.PreHooks = append(out.PreHooks, placeholder)
		}
	}

	return out
}

func printJSON(h config.HooksConfig, registry *toolchain.HookRegistry) error {
	out := fullOutput{
		hooksConfigOutput: hooksConfigOutput{
			Enabled:         h.Enabled,
			SecurityFilter:  h.SecurityFilter,
			AccessControl:   h.AccessControl,
			EventPublishing: h.EventPublishing,
			KnowledgeSave:   h.KnowledgeSave,
			BlockedCommands: h.BlockedCommands,
		},
		Registry: buildRegistryOutput(registry, h),
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func printText(h config.HooksConfig, registry *toolchain.HookRegistry) error {
	fmt.Println("Hook Configuration")

	if !h.Enabled {
		fmt.Println("  Hooks are disabled")
	}

	fmt.Printf("  Enabled:          %v\n", h.Enabled)
	fmt.Printf("  Security Filter:  %v\n", h.SecurityFilter)
	fmt.Printf("  Access Control:   %v\n", h.AccessControl)
	fmt.Printf("  Event Publishing: %v\n", h.EventPublishing)
	fmt.Printf("  Knowledge Save:   %v\n", h.KnowledgeSave)
	if len(h.BlockedCommands) > 0 {
		fmt.Printf("  Blocked Commands: %s\n", strings.Join(h.BlockedCommands, ", "))
	} else {
		fmt.Printf("  Blocked Commands: (none)\n")
	}

	fmt.Println()
	fmt.Println("Registered Hooks")

	regOut := buildRegistryOutput(registry, h)

	if len(regOut.PreHooks) > 0 {
		fmt.Println("  Pre-hooks:")
		for _, hi := range regOut.PreHooks {
			if hi.Wirable {
				fmt.Printf("    %-25s priority=%d  wirable=%v\n", hi.Name, hi.Priority, hi.Wirable)
			} else {
				fmt.Printf("    %-25s wirable=false  (%s)\n", hi.Name, hi.Reason)
			}
		}
	}

	if len(regOut.PostHooks) > 0 {
		fmt.Println("  Post-hooks:")
		for _, hi := range regOut.PostHooks {
			if hi.Wirable {
				fmt.Printf("    %-25s priority=%d  wirable=%v\n", hi.Name, hi.Priority, hi.Wirable)
			} else {
				fmt.Printf("    %-25s wirable=false  (%s)\n", hi.Name, hi.Reason)
			}
			if tools, ok := hi.Details["saveableTools"]; ok {
				if toolList, ok := tools.([]string); ok && len(toolList) > 0 {
					fmt.Printf("      saveableTools: %s\n", strings.Join(toolList, ", "))
				}
			}
		}
	}

	if len(regOut.PreHooks) == 0 && len(regOut.PostHooks) == 0 {
		fmt.Println("  (none)")
	}

	return nil
}
