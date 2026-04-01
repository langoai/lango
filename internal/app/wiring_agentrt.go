package app

import (
	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/agentrt"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/turnrunner"
)

// initAgentRuntime wraps the executor with CoordinatingExecutor in structured mode.
// In classic mode (default), the executor is returned unchanged with a nil budget.
// errorFixProvider may be nil if learning system is not available.
func initAgentRuntime(
	cfg *config.Config,
	innerExecutor turnrunner.Executor,
	bus *eventbus.Bus,
	errorFixProvider adk.ErrorFixProvider,
) (turnrunner.Executor, *agentrt.BudgetPolicy) {
	if cfg.Agent.Orchestration.Mode != "structured" {
		return innerExecutor, nil
	}

	orchCfg := cfg.Agent.Orchestration

	guard := agentrt.NewDelegationGuard(orchCfg.CircuitBreaker, bus)
	budget := agentrt.NewBudgetPolicy(orchCfg.Budget)
	recovery := agentrt.NewRecoveryPolicy(orchCfg.Recovery, errorFixProvider)

	budget.SetAlertHandler(func(alert agentrt.BudgetAlert) {
		if bus != nil {
			bus.Publish(agentrt.BudgetAlertEvent{
				Resource:   alert.Resource,
				Used:       alert.Used,
				Limit:      alert.Limit,
				Percentage: alert.Percentage,
			})
		}
	})

	logger().Infow("structured orchestration mode enabled",
		"mode", orchCfg.Mode,
		"circuitBreaker.threshold", orchCfg.CircuitBreaker.FailureThreshold,
		"budget.toolCallLimit", orchCfg.Budget.ToolCallLimit,
		"budget.delegationLimit", orchCfg.Budget.DelegationLimit,
		"recovery.maxRetries", orchCfg.Recovery.MaxRetries,
	)

	return agentrt.NewCoordinatingExecutor(innerExecutor, guard, budget, recovery, bus), budget
}
