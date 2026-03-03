package orchestration

import (
	"fmt"

	adk_agent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	adk_tool "google.golang.org/adk/tool"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/p2p/agentpool"
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
	// delegation rounds per user turn. Zero means use default (10).
	MaxDelegationRounds int
	// SubAgentPrompt builds the final system prompt for each sub-agent.
	// When nil, the original spec.Instruction is used unchanged.
	SubAgentPrompt SubAgentPromptFunc
	// UniversalTools are tools given directly to the orchestrator
	// (e.g. builtin_list/builtin_invoke dispatchers).
	UniversalTools []*agent.Tool
	// Specs overrides the default built-in agent specifications.
	// When nil, the built-in agentSpecs are used (backward compatible).
	Specs []AgentSpec
	// DynamicAgents provides P2P agents discovered at runtime.
	// When set, discovered agents are added to the routing table.
	DynamicAgents agentpool.DynamicAgentProvider
}

// BuildAgentTree creates a hierarchical agent tree with an orchestrator root
// and specialized sub-agents. Sub-agents are created data-driven from specs.
// When cfg.Specs is nil, the built-in agentSpecs are used (backward compatible).
// Agents with no tools are skipped unless AlwaysInclude is set (e.g. Planner).
func BuildAgentTree(cfg Config) (adk_agent.Agent, error) {
	if cfg.AdaptTool == nil {
		return nil, fmt.Errorf("build agent tree: AdaptTool is required")
	}

	// Determine which specs to use: explicit or built-in defaults.
	specs := cfg.Specs
	if specs == nil {
		specs = agentSpecs
	}

	// Use dynamic partitioning when explicit specs are provided,
	// otherwise fall back to the legacy RoleToolSet path for backward compatibility.
	var subAgents []adk_agent.Agent
	var routingEntries []routingEntry
	var unmatchedTools []*agent.Tool

	if cfg.Specs != nil {
		ds := PartitionToolsDynamic(cfg.Tools, specs)
		unmatchedTools = ds.Unmatched()

		for _, spec := range specs {
			tools := ds[spec.Name]
			if len(tools) == 0 && !spec.AlwaysInclude {
				continue
			}

			sa, entry, err := buildSubAgent(cfg, spec, tools)
			if err != nil {
				return nil, err
			}
			subAgents = append(subAgents, sa)
			routingEntries = append(routingEntries, entry)
		}
	} else {
		rs := PartitionTools(cfg.Tools)
		unmatchedTools = rs.Unmatched

		for _, spec := range specs {
			tools := toolsForSpec(spec, rs)
			if len(tools) == 0 && !spec.AlwaysInclude {
				continue
			}

			sa, entry, err := buildSubAgent(cfg, spec, tools)
			if err != nil {
				return nil, err
			}
			subAgents = append(subAgents, sa)
			routingEntries = append(routingEntries, entry)
		}
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

	// Append P2P dynamic agents to routing table.
	// These agents are invoked through p2p_invoke tool rather than direct delegation,
	// but they appear in the routing table so the orchestrator can decide when to use them.
	if cfg.DynamicAgents != nil {
		for _, da := range cfg.DynamicAgents.AvailableAgents() {
			routingEntries = append(routingEntries, routingEntry{
				Name:         fmt.Sprintf("p2p:%s", da.Name),
				Description:  fmt.Sprintf("%s (P2P remote agent, trust=%.2f)", da.Description, da.TrustScore),
				Keywords:     nil,
				Capabilities: da.Capabilities,
				Accepts:      "Use p2p_invoke tool with peer DID: " + da.DID,
				Returns:      "Remote tool execution results via P2P protocol",
			})
		}
	}

	maxRounds := cfg.MaxDelegationRounds
	if maxRounds <= 0 {
		maxRounds = 10
	}

	orchestratorInstruction := buildOrchestratorInstruction(
		cfg.SystemPrompt, routingEntries, maxRounds, unmatchedTools,
	)

	orchestrator, err := llmagent.New(llmagent.Config{
		Name:        "lango-orchestrator",
		Description: "Lango Assistant Orchestrator",
		Model:       cfg.Model,
		SubAgents:   subAgents,
		Instruction: orchestratorInstruction,
	})
	if err != nil {
		return nil, fmt.Errorf("create orchestrator agent: %w", err)
	}

	return orchestrator, nil
}

// buildSubAgent creates a single sub-agent from a spec and its assigned tools.
func buildSubAgent(cfg Config, spec AgentSpec, tools []*agent.Tool) (adk_agent.Agent, routingEntry, error) {
	var adkTools []adk_tool.Tool
	if len(tools) > 0 {
		var err error
		adkTools, err = adaptTools(cfg.AdaptTool, tools)
		if err != nil {
			return nil, routingEntry{}, fmt.Errorf("adapt %s tools: %w", spec.Name, err)
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
		return nil, routingEntry{}, fmt.Errorf("create %s agent: %w", spec.Name, err)
	}

	return a, buildRoutingEntry(spec, caps), nil
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
