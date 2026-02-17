package orchestration

import (
	"fmt"

	adk_agent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	adk_tool "google.golang.org/adk/tool"

	"github.com/langowarny/lango/internal/agent"
)

// ToolAdapter converts an internal agent.Tool to an ADK tool.Tool.
// This is injected to avoid a direct dependency on the adk package,
// which carries transitive imports that may cause import cycles.
type ToolAdapter func(t *agent.Tool) (adk_tool.Tool, error)

// SubAgentPromptFunc builds the final instruction for a sub-agent.
// agentName is the spec name (e.g. "operator"), defaultInstruction is
// the hard-coded spec.Instruction. The function returns the assembled
// system prompt that should replace spec.Instruction.
// When nil, the original spec.Instruction is used (backward compatible).
type SubAgentPromptFunc func(agentName, defaultInstruction string) string

// Config holds orchestration configuration.
type Config struct {
	// Tools is the full set of available tools.
	Tools []*agent.Tool
	// Model is the primary LLM model adapter.
	Model model.LLM
	// SystemPrompt is the base system instruction.
	SystemPrompt string
	// AdaptTool converts an internal tool to an ADK tool.
	// Callers should pass adk.AdaptTool.
	AdaptTool ToolAdapter
	// RemoteAgents are external A2A agents to include as sub-agents.
	RemoteAgents []adk_agent.Agent
	// MaxDelegationRounds limits the number of orchestrator→sub-agent
	// delegation rounds per user turn. Zero means use default (5).
	MaxDelegationRounds int
	// SubAgentPrompt builds the final system prompt for each sub-agent.
	// When nil, the original spec.Instruction is used unchanged.
	SubAgentPrompt SubAgentPromptFunc
}

// BuildAgentTree creates a hierarchical agent tree with an orchestrator root
// and specialized sub-agents. Sub-agents are created data-driven from agentSpecs.
// Agents with no tools are skipped unless AlwaysInclude is set (e.g. Planner).
func BuildAgentTree(cfg Config) (adk_agent.Agent, error) {
	if cfg.AdaptTool == nil {
		return nil, fmt.Errorf("build agent tree: AdaptTool is required")
	}

	rs := PartitionTools(cfg.Tools)

	var subAgents []adk_agent.Agent
	var routingEntries []routingEntry

	for _, spec := range agentSpecs {
		tools := toolsForSpec(spec, rs)
		if len(tools) == 0 && !spec.AlwaysInclude {
			continue
		}

		var adkTools []adk_tool.Tool
		if len(tools) > 0 {
			var err error
			adkTools, err = adaptTools(cfg.AdaptTool, tools)
			if err != nil {
				return nil, fmt.Errorf("adapt %s tools: %w", spec.Name, err)
			}
		}

		caps := capabilityDescription(tools)
		desc := spec.Description
		if caps != "" {
			desc = fmt.Sprintf("%s. Capabilities: %s", spec.Description, caps)
		}

		instruction := spec.Instruction
		if cfg.SubAgentPrompt != nil {
			instruction = cfg.SubAgentPrompt(spec.Name, spec.Instruction)
		}

		a, err := llmagent.New(llmagent.Config{
			Name:        spec.Name,
			Description: desc,
			Model:       cfg.Model,
			Tools:       adkTools,
			Instruction: instruction,
		})
		if err != nil {
			return nil, fmt.Errorf("create %s agent: %w", spec.Name, err)
		}

		subAgents = append(subAgents, a)
		routingEntries = append(routingEntries, buildRoutingEntry(spec, caps))
	}

	// Append remote A2A agents if configured.
	for _, ra := range cfg.RemoteAgents {
		subAgents = append(subAgents, ra)
		routingEntries = append(routingEntries, routingEntry{
			Name:        ra.Name(),
			Description: fmt.Sprintf("%s (remote A2A agent)", ra.Description()),
			Keywords:    nil,
			Accepts:     "Varies by remote agent capability",
			Returns:     "Varies by remote agent capability",
		})
	}

	maxRounds := cfg.MaxDelegationRounds
	if maxRounds <= 0 {
		maxRounds = 5
	}

	orchestratorInstruction := buildOrchestratorInstruction(
		cfg.SystemPrompt, routingEntries, maxRounds, rs.Unmatched,
	)

	orchestrator, err := llmagent.New(llmagent.Config{
		Name:        "lango-orchestrator",
		Description: "Lango Assistant Orchestrator",
		Model:       cfg.Model,
		Tools:       nil,
		SubAgents:   subAgents,
		Instruction: orchestratorInstruction,
	})
	if err != nil {
		return nil, fmt.Errorf("create orchestrator agent: %w", err)
	}

	return orchestrator, nil
}

// adaptTools converts a slice of internal agent tools to ADK tools using the provided adapter.
func adaptTools(adapt ToolAdapter, tools []*agent.Tool) ([]adk_tool.Tool, error) {
	result := make([]adk_tool.Tool, 0, len(tools))
	for _, t := range tools {
		adapted, err := adapt(t)
		if err != nil {
			return nil, fmt.Errorf("adapt tool %q: %w", t.Name, err)
		}
		result = append(result, adapted)
	}
	return result, nil
}
