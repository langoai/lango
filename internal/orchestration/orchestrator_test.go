package orchestration

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adk_tool "google.golang.org/adk/tool"

	"github.com/langowarny/lango/internal/agent"
)

func newTestTool(name string) *agent.Tool {
	return &agent.Tool{
		Name:        name,
		Description: "test tool " + name,
		SafetyLevel: agent.SafetyLevelSafe,
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	}
}

// stubTool implements adk_tool.Tool for testing.
type stubTool struct {
	name string
}

func (s *stubTool) Name() string       { return s.name }
func (s *stubTool) Description() string { return "stub " + s.name }
func (s *stubTool) IsLongRunning() bool { return false }

// stubAdapter is a ToolAdapter that returns a stubTool without real ADK wiring.
func stubAdapter(t *agent.Tool) (adk_tool.Tool, error) {
	return &stubTool{name: t.Name}, nil
}

// failingAdapter always returns an error.
func failingAdapter(t *agent.Tool) (adk_tool.Tool, error) {
	return nil, fmt.Errorf("adapter error for %s", t.Name)
}

// --- PartitionTools tests ---

func TestPartitionTools(t *testing.T) {
	tests := []struct {
		name           string
		give           []*agent.Tool
		wantOperator   []string
		wantNavigator  []string
		wantVault      []string
		wantLibrarian  []string
		wantPlanner    []string
		wantChronicler []string
		wantUnmatched  []string
	}{
		{
			name: "operator prefixes",
			give: []*agent.Tool{
				newTestTool("exec_shell"),
				newTestTool("fs_read"),
				newTestTool("skill_deploy"),
			},
			wantOperator: []string{"exec_shell", "fs_read", "skill_deploy"},
		},
		{
			name: "navigator prefixes",
			give: []*agent.Tool{
				newTestTool("browser_navigate"),
				newTestTool("browser_screenshot"),
			},
			wantNavigator: []string{"browser_navigate", "browser_screenshot"},
		},
		{
			name: "vault prefixes",
			give: []*agent.Tool{
				newTestTool("crypto_sign"),
				newTestTool("secrets_get"),
				newTestTool("payment_send"),
			},
			wantVault: []string{"crypto_sign", "secrets_get", "payment_send"},
		},
		{
			name: "librarian prefixes including exact matches",
			give: []*agent.Tool{
				newTestTool("search_web"),
				newTestTool("rag_query"),
				newTestTool("graph_traverse"),
				newTestTool("save_knowledge_item"),
				newTestTool("save_learning_rule"),
				newTestTool("learning_stats"),
				newTestTool("learning_cleanup"),
				newTestTool("create_skill_x"),
				newTestTool("list_skills"),
				newTestTool("import_skill"),
			},
			wantLibrarian: []string{
				"search_web", "rag_query", "graph_traverse",
				"save_knowledge_item", "save_learning_rule",
				"learning_stats", "learning_cleanup",
				"create_skill_x", "list_skills", "import_skill",
			},
		},
		{
			name: "chronicler prefixes",
			give: []*agent.Tool{
				newTestTool("memory_store"),
				newTestTool("observe_event"),
				newTestTool("reflect_summary"),
			},
			wantChronicler: []string{"memory_store", "observe_event", "reflect_summary"},
		},
		{
			name: "unmatched tools tracked separately",
			give: []*agent.Tool{
				newTestTool("custom_action"),
				newTestTool("do_something"),
			},
			wantUnmatched: []string{"custom_action", "do_something"},
		},
		{
			name: "mixed tools partitioned correctly across 6 roles",
			give: []*agent.Tool{
				newTestTool("exec_run"),
				newTestTool("browser_click"),
				newTestTool("crypto_encrypt"),
				newTestTool("search_docs"),
				newTestTool("memory_save"),
				newTestTool("unknown_tool"),
			},
			wantOperator:   []string{"exec_run"},
			wantNavigator:  []string{"browser_click"},
			wantVault:      []string{"crypto_encrypt"},
			wantLibrarian:  []string{"search_docs"},
			wantChronicler: []string{"memory_save"},
			wantUnmatched:  []string{"unknown_tool"},
		},
		{
			name: "empty input",
			give: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PartitionTools(tt.give)

			assert.Equal(t, tt.wantOperator, toolNames(got.Operator), "operator tools")
			assert.Equal(t, tt.wantNavigator, toolNames(got.Navigator), "navigator tools")
			assert.Equal(t, tt.wantVault, toolNames(got.Vault), "vault tools")
			assert.Equal(t, tt.wantLibrarian, toolNames(got.Librarian), "librarian tools")
			assert.Equal(t, tt.wantPlanner, toolNames(got.Planner), "planner tools")
			assert.Equal(t, tt.wantChronicler, toolNames(got.Chronicler), "chronicler tools")
			assert.Equal(t, tt.wantUnmatched, toolNames(got.Unmatched), "unmatched tools")
		})
	}
}

func TestPartitionTools_PrefixPriority(t *testing.T) {
	// Verify librarian prefixes are checked before operator defaults.
	tools := []*agent.Tool{
		newTestTool("search_rag"),
		newTestTool("graph_node"),
		newTestTool("save_knowledge_data"),
		newTestTool("create_skill_new"),
	}

	got := PartitionTools(tools)

	assert.Empty(t, got.Operator, "no tools should go to operator")
	assert.Empty(t, got.Unmatched, "no tools should be unmatched")
	assert.Len(t, got.Librarian, 4, "all should be librarian")
}

// --- BuildAgentTree tests ---

func TestBuildAgentTree_NilAdaptTool(t *testing.T) {
	_, err := BuildAgentTree(Config{
		Tools:        []*agent.Tool{newTestTool("exec_shell")},
		Model:        nil,
		SystemPrompt: "test",
		AdaptTool:    nil,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AdaptTool is required")
}

func TestBuildAgentTree_Success(t *testing.T) {
	tools := []*agent.Tool{
		newTestTool("exec_shell"),     // operator
		newTestTool("browser_open"),   // navigator
		newTestTool("crypto_sign"),    // vault
		newTestTool("search_web"),     // librarian
		newTestTool("memory_store"),   // chronicler
		newTestTool("custom_unknown"), // unmatched
	}

	root, err := BuildAgentTree(Config{
		Tools:        tools,
		Model:        nil,
		SystemPrompt: "test prompt",
		AdaptTool:    stubAdapter,
	})
	require.NoError(t, err)
	require.NotNil(t, root)

	assert.Equal(t, "lango-orchestrator", root.Name())
	// operator, navigator, vault, librarian, planner (always), chronicler = 6
	assert.Len(t, root.SubAgents(), 6, "orchestrator should have 6 sub-agents")

	subNames := make([]string, len(root.SubAgents()))
	for i, sa := range root.SubAgents() {
		subNames[i] = sa.Name()
	}
	assert.Contains(t, subNames, "operator")
	assert.Contains(t, subNames, "navigator")
	assert.Contains(t, subNames, "vault")
	assert.Contains(t, subNames, "librarian")
	assert.Contains(t, subNames, "planner")
	assert.Contains(t, subNames, "chronicler")
}

func TestBuildAgentTree_NoTools(t *testing.T) {
	// No tools at all — only planner should be created (AlwaysInclude).
	root, err := BuildAgentTree(Config{
		Tools:        nil,
		Model:        nil,
		SystemPrompt: "test prompt",
		AdaptTool:    stubAdapter,
	})
	require.NoError(t, err)

	assert.Len(t, root.SubAgents(), 1)
	assert.Equal(t, "planner", root.SubAgents()[0].Name())
}

func TestBuildAgentTree_PartialAgents(t *testing.T) {
	// Only operator and librarian tools — other roles should be skipped.
	tools := []*agent.Tool{
		newTestTool("exec_shell"),
		newTestTool("search_web"),
	}

	root, err := BuildAgentTree(Config{
		Tools:        tools,
		Model:        nil,
		SystemPrompt: "test prompt",
		AdaptTool:    stubAdapter,
	})
	require.NoError(t, err)

	// operator + librarian + planner (always) = 3
	assert.Len(t, root.SubAgents(), 3)

	subNames := make([]string, len(root.SubAgents()))
	for i, sa := range root.SubAgents() {
		subNames[i] = sa.Name()
	}
	assert.Contains(t, subNames, "operator")
	assert.Contains(t, subNames, "librarian")
	assert.Contains(t, subNames, "planner")
	assert.NotContains(t, subNames, "navigator")
	assert.NotContains(t, subNames, "vault")
	assert.NotContains(t, subNames, "chronicler")
}

func TestBuildAgentTree_UnmatchedToolsNotAssigned(t *testing.T) {
	// Unmatched tools should NOT be assigned to any sub-agent.
	tools := []*agent.Tool{
		newTestTool("custom_action"),
		newTestTool("do_something"),
	}

	var adaptedTools []string
	trackingAdapter := func(tool *agent.Tool) (adk_tool.Tool, error) {
		adaptedTools = append(adaptedTools, tool.Name)
		return &stubTool{name: tool.Name}, nil
	}

	root, err := BuildAgentTree(Config{
		Tools:        tools,
		Model:        nil,
		SystemPrompt: "test prompt",
		AdaptTool:    trackingAdapter,
	})
	require.NoError(t, err)

	// Only planner (always included), no tools adapted since unmatched.
	assert.Len(t, root.SubAgents(), 1)
	assert.Equal(t, "planner", root.SubAgents()[0].Name())
	assert.Empty(t, adaptedTools, "unmatched tools should not be adapted")
}

func TestBuildAgentTree_RoutingTableInInstruction(t *testing.T) {
	// Build the same routing entries that BuildAgentTree would produce and
	// verify buildOrchestratorInstruction includes them. Agent.Instruction()
	// is not part of the public ADK interface, so we test the builder directly.
	tools := []*agent.Tool{
		newTestTool("exec_shell"),
		newTestTool("browser_open"),
		newTestTool("crypto_sign"),
		newTestTool("search_web"),
		newTestTool("memory_store"),
	}

	rs := PartitionTools(tools)
	var entries []routingEntry
	for _, spec := range agentSpecs {
		st := toolsForSpec(spec, rs)
		if len(st) == 0 && !spec.AlwaysInclude {
			continue
		}
		entries = append(entries, buildRoutingEntry(spec, capabilityDescription(st)))
	}

	inst := buildOrchestratorInstruction("test prompt", entries, 5, rs.Unmatched)

	assert.Contains(t, inst, "Routing Table")
	assert.Contains(t, inst, "### operator")
	assert.Contains(t, inst, "### navigator")
	assert.Contains(t, inst, "### vault")
	assert.Contains(t, inst, "### librarian")
	assert.Contains(t, inst, "### planner")
	assert.Contains(t, inst, "### chronicler")

	// Should contain decision protocol.
	assert.Contains(t, inst, "Decision Protocol")
	assert.Contains(t, inst, "CLASSIFY")
	assert.Contains(t, inst, "MATCH")
	assert.Contains(t, inst, "SELECT")
	assert.Contains(t, inst, "VERIFY")
	assert.Contains(t, inst, "DELEGATE")

	// Should contain rejection handling.
	assert.Contains(t, inst, "Rejection Handling")
	assert.Contains(t, inst, "[REJECT]")
}

func TestBuildAgentTree_RejectProtocolInInstructions(t *testing.T) {
	// Agent.Instruction() is not part of the public ADK interface, so we
	// verify reject protocol presence via the agentSpecs registry directly.
	// This is covered by TestAgentSpecs_AllHaveRejectProtocol as well,
	// but this test verifies that all specs used by BuildAgentTree include it.
	tools := []*agent.Tool{
		newTestTool("exec_shell"),
		newTestTool("browser_open"),
		newTestTool("crypto_sign"),
		newTestTool("search_web"),
		newTestTool("memory_store"),
	}

	rs := PartitionTools(tools)
	for _, spec := range agentSpecs {
		st := toolsForSpec(spec, rs)
		if len(st) == 0 && !spec.AlwaysInclude {
			continue
		}
		assert.Contains(t, spec.Instruction, "[REJECT]",
			"spec %q should have reject protocol in instruction", spec.Name)
	}
}

func TestBuildAgentTree_AdapterError(t *testing.T) {
	tools := []*agent.Tool{newTestTool("exec_shell")}

	_, err := BuildAgentTree(Config{
		Tools:        tools,
		Model:        nil,
		SystemPrompt: "test",
		AdaptTool:    failingAdapter,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "adapt operator tools")
}

func TestBuildAgentTree_OrchestratorHasNoDirectTools(t *testing.T) {
	var adaptedTools []string
	trackingAdapter := func(tool *agent.Tool) (adk_tool.Tool, error) {
		adaptedTools = append(adaptedTools, tool.Name)
		return &stubTool{name: tool.Name}, nil
	}

	tools := []*agent.Tool{
		newTestTool("exec_shell"),
		newTestTool("browser_open"),
		newTestTool("crypto_sign"),
		newTestTool("search_web"),
		newTestTool("memory_store"),
	}

	root, err := BuildAgentTree(Config{
		Tools:        tools,
		Model:        nil,
		SystemPrompt: "test prompt",
		AdaptTool:    trackingAdapter,
	})
	require.NoError(t, err)

	assert.Len(t, root.SubAgents(), 6,
		"orchestrator should have 6 sub-agents")

	// Each tool adapted exactly once (for its sub-agent, not the orchestrator).
	toolAdaptCounts := make(map[string]int, len(tools))
	for _, name := range adaptedTools {
		toolAdaptCounts[name]++
	}
	for _, tool := range tools {
		assert.Equal(t, 1, toolAdaptCounts[tool.Name],
			"tool %q should be adapted only once (for sub-agent)", tool.Name)
	}
}

func TestBuildAgentTree_DescriptionsUseCapabilities(t *testing.T) {
	tools := []*agent.Tool{
		newTestTool("exec_shell"),
		newTestTool("browser_navigate"),
		newTestTool("crypto_sign"),
		newTestTool("search_web"),
		newTestTool("memory_store"),
	}

	root, err := BuildAgentTree(Config{
		Tools:        tools,
		Model:        nil,
		SystemPrompt: "test prompt",
		AdaptTool:    stubAdapter,
	})
	require.NoError(t, err)

	// Verify sub-agent descriptions do NOT contain raw tool name prefixes.
	for _, sa := range root.SubAgents() {
		desc := sa.Description()
		assert.NotContains(t, desc, "exec_shell",
			"agent %q description should not contain raw tool names", sa.Name())
		assert.NotContains(t, desc, "browser_navigate",
			"agent %q description should not contain raw tool names", sa.Name())
		assert.NotContains(t, desc, "search_web",
			"agent %q description should not contain raw tool names", sa.Name())
		assert.NotContains(t, desc, "memory_store",
			"agent %q description should not contain raw tool names", sa.Name())
	}
}

// --- SubAgentPromptFunc tests ---

func TestBuildAgentTree_SubAgentPromptFunc(t *testing.T) {
	tools := []*agent.Tool{
		newTestTool("exec_shell"),
		newTestTool("search_web"),
	}

	// Track which agent names and default instructions the func receives.
	calls := make(map[string]string)
	promptFunc := func(agentName, defaultInstruction string) string {
		calls[agentName] = defaultInstruction
		return "## Custom Prompt for " + agentName + "\n" + defaultInstruction
	}

	root, err := BuildAgentTree(Config{
		Tools:          tools,
		Model:          nil,
		SystemPrompt:   "test prompt",
		AdaptTool:      stubAdapter,
		SubAgentPrompt: promptFunc,
	})
	require.NoError(t, err)
	require.NotNil(t, root)

	// operator, librarian, planner should all have been called.
	assert.Contains(t, calls, "operator")
	assert.Contains(t, calls, "librarian")
	assert.Contains(t, calls, "planner")

	// Default instructions should have been passed through.
	assert.Contains(t, calls["operator"], "## What You Do")
	assert.Contains(t, calls["planner"], "## What You Do")
}

func TestBuildAgentTree_NilSubAgentPromptFunc(t *testing.T) {
	// With nil SubAgentPrompt, agents should use spec.Instruction unchanged.
	tools := []*agent.Tool{newTestTool("exec_shell")}

	root, err := BuildAgentTree(Config{
		Tools:          tools,
		Model:          nil,
		SystemPrompt:   "test prompt",
		AdaptTool:      stubAdapter,
		SubAgentPrompt: nil,
	})
	require.NoError(t, err)
	require.NotNil(t, root)

	// operator + planner = 2 agents
	assert.Len(t, root.SubAgents(), 2)
}

func TestBuildAgentTree_SubAgentPromptFunc_AllAgents(t *testing.T) {
	tools := []*agent.Tool{
		newTestTool("exec_shell"),
		newTestTool("browser_open"),
		newTestTool("crypto_sign"),
		newTestTool("search_web"),
		newTestTool("memory_store"),
	}

	var calledAgents []string
	promptFunc := func(agentName, defaultInstruction string) string {
		calledAgents = append(calledAgents, agentName)
		return "SAFETY RULES\n\n" + defaultInstruction + "\n\nCONVERSATION RULES"
	}

	root, err := BuildAgentTree(Config{
		Tools:          tools,
		Model:          nil,
		SystemPrompt:   "test prompt",
		AdaptTool:      stubAdapter,
		SubAgentPrompt: promptFunc,
	})
	require.NoError(t, err)

	// All 6 agents should have been processed.
	assert.Len(t, calledAgents, 6)
	assert.Contains(t, calledAgents, "operator")
	assert.Contains(t, calledAgents, "navigator")
	assert.Contains(t, calledAgents, "vault")
	assert.Contains(t, calledAgents, "librarian")
	assert.Contains(t, calledAgents, "planner")
	assert.Contains(t, calledAgents, "chronicler")

	assert.Len(t, root.SubAgents(), 6)
}

// --- buildRoutingEntry tests ---

func TestBuildRoutingEntry(t *testing.T) {
	tests := []struct {
		name     string
		give     AgentSpec
		giveCaps string
		wantName string
		wantDesc string
	}{
		{
			name: "with capabilities",
			give: AgentSpec{
				Name:        "operator",
				Description: "System operations",
				Keywords:    []string{"run", "execute"},
				Accepts:     "A command",
				Returns:     "Command output",
				CannotDo:    []string{"web browsing"},
			},
			giveCaps: "command execution, file operations",
			wantName: "operator",
			wantDesc: "System operations. Capabilities: command execution, file operations",
		},
		{
			name: "without capabilities (planner)",
			give: AgentSpec{
				Name:        "planner",
				Description: "Task decomposition",
				Keywords:    []string{"plan"},
				Accepts:     "A task",
				Returns:     "A plan",
			},
			giveCaps: "",
			wantName: "planner",
			wantDesc: "Task decomposition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildRoutingEntry(tt.give, tt.giveCaps)

			assert.Equal(t, tt.wantName, got.Name)
			assert.Equal(t, tt.wantDesc, got.Description)
			assert.Equal(t, tt.give.Keywords, got.Keywords)
			assert.Equal(t, tt.give.Accepts, got.Accepts)
			assert.Equal(t, tt.give.Returns, got.Returns)
			assert.Equal(t, tt.give.CannotDo, got.CannotDo)
		})
	}
}

// --- toolCapability tests ---

func TestToolCapability(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{give: "exec_shell", want: "command execution"},
		{give: "fs_read", want: "file operations"},
		{give: "skill_deploy", want: "skill management"},
		{give: "browser_navigate", want: "web browsing"},
		{give: "crypto_sign", want: "cryptography"},
		{give: "secrets_get", want: "secret management"},
		{give: "payment_send", want: "blockchain payments (USDC on Base)"},
		{give: "search_web", want: "information search"},
		{give: "rag_query", want: "knowledge retrieval (RAG)"},
		{give: "graph_traverse", want: "knowledge graph traversal"},
		{give: "save_knowledge_item", want: "knowledge persistence"},
		{give: "save_learning_rule", want: "learning persistence"},
		{give: "learning_stats", want: "learning data management"},
		{give: "create_skill_x", want: "skill creation"},
		{give: "list_skills", want: "skill listing"},
		{give: "import_skill", want: "skill import from external sources"},
		{give: "memory_store", want: "memory storage and recall"},
		{give: "observe_event", want: "event observation"},
		{give: "reflect_summary", want: "reflection and summarization"},
		{give: "unknown_tool", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := toolCapability(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCapabilityDescription(t *testing.T) {
	tests := []struct {
		name string
		give []*agent.Tool
		want string
	}{
		{
			name: "operator tools described as capabilities",
			give: []*agent.Tool{
				newTestTool("exec_shell"),
				newTestTool("fs_read"),
				newTestTool("skill_run"),
			},
			want: "command execution, file operations, skill management",
		},
		{
			name: "duplicate capabilities are deduplicated",
			give: []*agent.Tool{
				newTestTool("exec_shell"),
				newTestTool("exec_run"),
			},
			want: "command execution",
		},
		{
			name: "unknown tools get general actions",
			give: []*agent.Tool{
				newTestTool("custom_action"),
			},
			want: "general actions",
		},
		{
			name: "empty tools",
			give: nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := capabilityDescription(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- buildOrchestratorInstruction tests ---

func TestBuildOrchestratorInstruction_ContainsRoutingTable(t *testing.T) {
	entries := []routingEntry{
		{
			Name:        "operator",
			Description: "System ops",
			Keywords:    []string{"run", "execute"},
			Accepts:     "A command",
			Returns:     "Output",
			CannotDo:    []string{"web browsing"},
		},
		{
			Name:        "planner",
			Description: "Planning",
			Keywords:    []string{"plan"},
			Accepts:     "A task",
			Returns:     "A plan",
		},
	}

	got := buildOrchestratorInstruction("base prompt", entries, 5, nil)

	assert.Contains(t, got, "base prompt")
	assert.Contains(t, got, "### operator")
	assert.Contains(t, got, "### planner")
	assert.Contains(t, got, "run, execute")
	assert.Contains(t, got, "web browsing")
	assert.Contains(t, got, "Decision Protocol")
	assert.Contains(t, got, "Maximum 5 delegation rounds")
}

func TestBuildOrchestratorInstruction_UnmatchedTools(t *testing.T) {
	unmatched := []*agent.Tool{
		newTestTool("custom_action"),
		newTestTool("special_op"),
	}

	got := buildOrchestratorInstruction("base", nil, 3, unmatched)

	assert.Contains(t, got, "Unmatched Tools")
	assert.Contains(t, got, "custom_action")
	assert.Contains(t, got, "special_op")
}

func TestBuildOrchestratorInstruction_NoUnmatchedTools(t *testing.T) {
	got := buildOrchestratorInstruction("base", nil, 5, nil)

	assert.NotContains(t, got, "Unmatched Tools")
}

// --- Agent spec consistency tests ---

func TestAgentSpecs_AllHaveRejectProtocol(t *testing.T) {
	for _, spec := range agentSpecs {
		assert.Contains(t, spec.Instruction, "[REJECT]",
			"spec %q must have reject protocol", spec.Name)
	}
}

func TestAgentSpecs_AllHaveKeywords(t *testing.T) {
	for _, spec := range agentSpecs {
		assert.NotEmpty(t, spec.Keywords,
			"spec %q must have keywords", spec.Name)
	}
}

func TestAgentSpecs_UniqueNames(t *testing.T) {
	seen := make(map[string]bool, len(agentSpecs))
	for _, spec := range agentSpecs {
		assert.False(t, seen[spec.Name],
			"duplicate agent spec name: %q", spec.Name)
		seen[spec.Name] = true
	}
}

func TestAgentSpecs_AllHaveAcceptsAndReturns(t *testing.T) {
	for _, spec := range agentSpecs {
		assert.NotEmpty(t, spec.Accepts,
			"spec %q must have Accepts", spec.Name)
		assert.NotEmpty(t, spec.Returns,
			"spec %q must have Returns", spec.Name)
	}
}

func TestAgentSpecs_PlannerHasNoPrefixes(t *testing.T) {
	for _, spec := range agentSpecs {
		if spec.Name == "planner" {
			assert.Empty(t, spec.Prefixes, "planner should have no prefixes")
			assert.True(t, spec.AlwaysInclude, "planner should be AlwaysInclude")
		}
	}
}

func TestAgentSpecs_InstructionStructure(t *testing.T) {
	requiredSections := []string{
		"## What You Do",
		"## Input Format",
		"## Output Format",
		"## Constraints",
	}

	for _, spec := range agentSpecs {
		for _, section := range requiredSections {
			assert.True(t, strings.Contains(spec.Instruction, section),
				"spec %q instruction must contain %q", spec.Name, section)
		}
	}
}

// --- helpers ---

// toolNames extracts names from a tool slice for assertions.
func toolNames(tools []*agent.Tool) []string {
	if len(tools) == 0 {
		return nil
	}
	names := make([]string, len(tools))
	for i, t := range tools {
		names[i] = t.Name
	}
	return names
}
