package adk

import (
	"context"
	"fmt"
	"iter"
	"strings"

	adk_agent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"

	internal "github.com/langowarny/lango/internal/session"
)

// Agent wraps the ADK runner for integration with Lango.
type Agent struct {
	runner   *runner.Runner
	adkAgent adk_agent.Agent
}

// NewAgent creates a new Agent instance.
func NewAgent(ctx context.Context, tools []tool.Tool, mod model.LLM, systemPrompt string, store internal.Store) (*Agent, error) {
	// Create LLM Agent
	cfg := llmagent.Config{
		Name:        "lango-agent",
		Description: "Lango Assistant",
		Model:       mod,
		Tools:       tools,
		Instruction: systemPrompt,
	}

	adkAgent, err := llmagent.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create llm agent: %w", err)
	}

	// Create Session Service
	sessService := NewSessionServiceAdapter(store, "lango-agent")

	// Create Runner
	runnerCfg := runner.Config{
		AppName:        "lango",
		Agent:          adkAgent,
		SessionService: sessService,
	}

	r, err := runner.New(runnerCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %w", err)
	}

	return &Agent{
		runner:   r,
		adkAgent: adkAgent,
	}, nil
}

// NewAgentFromADK creates a Lango Agent wrapping a pre-built ADK agent.
// Used for multi-agent orchestration where the agent tree is built externally.
func NewAgentFromADK(adkAgent adk_agent.Agent, store internal.Store) (*Agent, error) {
	sessService := NewSessionServiceAdapter(store, adkAgent.Name())

	runnerCfg := runner.Config{
		AppName:        "lango",
		Agent:          adkAgent,
		SessionService: sessService,
	}

	r, err := runner.New(runnerCfg)
	if err != nil {
		return nil, fmt.Errorf("create runner: %w", err)
	}

	return &Agent{runner: r, adkAgent: adkAgent}, nil
}

// ADKAgent returns the underlying ADK agent, or nil if not available.
func (a *Agent) ADKAgent() adk_agent.Agent {
	return a.adkAgent
}

// Run executes the agent for a given session and returns an event iterator.
func (a *Agent) Run(ctx context.Context, sessionID string, input string) iter.Seq2[*session.Event, error] {
	// Create user content
	userMsg := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: input}},
	}

	// Config for run
	runCfg := adk_agent.RunConfig{
		// Defaults
	}

	// Execute via Runner
	return a.runner.Run(ctx, "user", sessionID, userMsg, runCfg)
}

// RunAndCollect executes the agent and returns the full text response.
func (a *Agent) RunAndCollect(ctx context.Context, sessionID, input string) (string, error) {
	var b strings.Builder

	for event, err := range a.Run(ctx, sessionID, input) {
		if err != nil {
			return "", fmt.Errorf("agent error: %w", err)
		}
		if event.Content != nil {
			for _, part := range event.Content.Parts {
				if part.Text != "" {
					b.WriteString(part.Text)
				}
			}
		}
	}

	return b.String(), nil
}
