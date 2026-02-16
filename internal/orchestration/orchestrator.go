package orchestration

import (
	"fmt"
	"strings"

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
	// delegation rounds per user turn. Zero means use default (3).
	MaxDelegationRounds int
}

// BuildAgentTree creates a hierarchical agent tree with an orchestrator root
// and specialized sub-agents. Sub-agents are only created when they have
// tools assigned (except Planner which is LLM-only and always included).
func BuildAgentTree(cfg Config) (adk_agent.Agent, error) {
	if cfg.AdaptTool == nil {
		return nil, fmt.Errorf("build agent tree: AdaptTool is required")
	}

	rs := PartitionTools(cfg.Tools)

	var subAgents []adk_agent.Agent
	var agentDescriptions []string

	// Executor: only if tools are assigned.
	if len(rs.Executor) > 0 {
		executorTools, err := adaptTools(cfg.AdaptTool, rs.Executor)
		if err != nil {
			return nil, fmt.Errorf("adapt executor tools: %w", err)
		}
		a, err := llmagent.New(llmagent.Config{
			Name:        "executor",
			Description: "Executes tools including shell commands, file operations, browser automation, and cryptographic operations. Delegate to this agent when the user needs to perform actions.",
			Model:       cfg.Model,
			Tools:       executorTools,
			Instruction: "You are the Executor agent. Execute tool calls precisely as requested. Report results accurately. If a tool fails, report the error without retrying unless explicitly asked.",
		})
		if err != nil {
			return nil, fmt.Errorf("create executor agent: %w", err)
		}
		subAgents = append(subAgents, a)
		agentDescriptions = append(agentDescriptions, "- executor: for running tools and performing actions")
	}

	// Researcher: only if tools are assigned.
	if len(rs.Researcher) > 0 {
		researcherTools, err := adaptTools(cfg.AdaptTool, rs.Researcher)
		if err != nil {
			return nil, fmt.Errorf("adapt researcher tools: %w", err)
		}
		a, err := llmagent.New(llmagent.Config{
			Name:        "researcher",
			Description: "Searches knowledge bases, performs RAG retrieval, and traverses the knowledge graph. Delegate to this agent for information lookup and research tasks.",
			Model:       cfg.Model,
			Tools:       researcherTools,
			Instruction: "You are the Researcher agent. Search and retrieve relevant information from knowledge bases, semantic search, and the knowledge graph. Provide comprehensive, well-organized results.",
		})
		if err != nil {
			return nil, fmt.Errorf("create researcher agent: %w", err)
		}
		subAgents = append(subAgents, a)
		agentDescriptions = append(agentDescriptions, "- researcher: for searching knowledge and information retrieval")
	}

	// Planner: always included (LLM-only reasoning agent).
	{
		a, err := llmagent.New(llmagent.Config{
			Name:        "planner",
			Description: "Decomposes complex tasks into steps and designs execution plans. Delegate to this agent when the user needs task planning or strategy.",
			Model:       cfg.Model,
			Instruction: "You are the Planner agent. Break complex tasks into clear, actionable steps. Consider dependencies between steps. Output structured plans.",
		})
		if err != nil {
			return nil, fmt.Errorf("create planner agent: %w", err)
		}
		subAgents = append(subAgents, a)
		agentDescriptions = append(agentDescriptions, "- planner: for task decomposition and planning")
	}

	// Memory Manager: only if memory tools are assigned.
	if len(rs.MemoryManager) > 0 {
		memoryTools, err := adaptTools(cfg.AdaptTool, rs.MemoryManager)
		if err != nil {
			return nil, fmt.Errorf("adapt memory tools: %w", err)
		}
		a, err := llmagent.New(llmagent.Config{
			Name:        "memory-manager",
			Description: "Manages conversational memory including observations, reflections, and the memory graph. Delegate to this agent for memory-related operations.",
			Model:       cfg.Model,
			Tools:       memoryTools,
			Instruction: "You are the Memory Manager agent. Manage observations, reflections, and the memory graph. Organize and retrieve relevant past interactions.",
		})
		if err != nil {
			return nil, fmt.Errorf("create memory agent: %w", err)
		}
		subAgents = append(subAgents, a)
		agentDescriptions = append(agentDescriptions, "- memory-manager: for memory operations")
	}

	// Append remote A2A agents if configured.
	subAgents = append(subAgents, cfg.RemoteAgents...)
	for _, ra := range cfg.RemoteAgents {
		agentDescriptions = append(agentDescriptions,
			fmt.Sprintf("- %s: %s (remote A2A agent)", ra.Name(), ra.Description()))
	}

	// Adapt all tools for the orchestrator so it can handle simple tasks
	// directly without delegating to sub-agents.
	allAdkTools, err := adaptTools(cfg.AdaptTool, cfg.Tools)
	if err != nil {
		return nil, fmt.Errorf("adapt orchestrator tools: %w", err)
	}

	// Build orchestrator instruction with direct tool usage guidance
	// and max delegation rounds guardrail.
	maxRounds := cfg.MaxDelegationRounds
	if maxRounds <= 0 {
		maxRounds = 3
	}

	orchestratorInstruction := cfg.SystemPrompt + fmt.Sprintf(`

You are the orchestrator. You have tools available for direct use AND specialized sub-agents for complex tasks.

## Direct Tool Usage
For simple, single-step tasks, call the appropriate tool directly. Do NOT delegate single-tool operations to a sub-agent.

## Sub-Agent Delegation
For complex, multi-step tasks requiring specialized reasoning, delegate to a sub-agent.
ONLY use these exact agent names:
%s

## Rules
- Simple tasks (single tool call, greetings, casual questions): handle directly.
- Complex tasks (multi-step reasoning, research synthesis): delegate to the appropriate sub-agent listed above.
- Conversational queries: respond directly without tools or delegation.
- NEVER invent agent names. Only use the exact names listed above.
- Maximum %d delegation rounds per user turn. After that, synthesize and respond.`, strings.Join(agentDescriptions, "\n"), maxRounds)

	orchestrator, err := llmagent.New(llmagent.Config{
		Name:        "lango-orchestrator",
		Description: "Lango Assistant Orchestrator",
		Model:       cfg.Model,
		Tools:       allAdkTools,
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
