package runledger

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/langoai/lango/internal/toolparam"
)

// callerRole identifies who is invoking a tool for access control.
type callerRole string

const (
	roleOrchestrator callerRole = "orchestrator"
	roleExecution    callerRole = "execution"
	roleAny          callerRole = "any"
)

// SystemCallerName is the explicit identity for trusted internal callers.
const SystemCallerName = "system"

// BuildTools creates all run_* tools with access control.
func BuildTools(store RunLedgerStore, pev *PEVEngine) []*agent.Tool {
	return []*agent.Tool{
		buildRunCreate(store),
		buildRunRead(store),
		buildRunActive(store),
		buildRunNote(store),
		buildRunProposeStepResult(store, pev),
		buildRunApplyPolicy(store),
		buildRunApproveStep(store, pev),
		buildRunResume(store),
	}
}

func buildRunCreate(store RunLedgerStore) *agent.Tool {
	return &agent.Tool{
		Name:        "run_create",
		Description: "Create a new Run from a planner's JSON plan. Only the orchestrator may call this.",
		SafetyLevel: agent.SafetyLevelModerate,
		Parameters: agent.Schema().
			Str("plan_json", "The planner's JSON output (goal, steps, acceptance_criteria)").
			Str("session_key", "Session key for this run").
			Str("original_request", "The user's original request text").
			Array("valid_agents", "string", "List of valid agent names for validation").
			Required("plan_json", "session_key", "original_request").
			Build(),
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			if err := checkRole(ctx, roleOrchestrator); err != nil {
				return nil, err
			}

			planJSON := toolparam.OptionalString(params, "plan_json", "")
			sessionKey := toolparam.OptionalString(params, "session_key", "")
			originalRequest := toolparam.OptionalString(params, "original_request", "")

			// Parse planner output.
			plan, err := ParsePlannerOutput(planJSON)
			if err != nil {
				return map[string]interface{}{
					"error":   "parse_failed",
					"message": err.Error(),
				}, nil
			}

			validAgents := toolparam.StringSlice(params, "valid_agents")

			// Validate plan schema.
			if err := ValidatePlanSchema(plan, validAgents); err != nil {
				return map[string]interface{}{
					"error":   "validation_failed",
					"message": err.Error(),
				}, nil
			}

			// Generate run ID.
			runID := uuid.New().String()

			// Append run_created event.
			if err := store.AppendJournalEvent(ctx, JournalEvent{
				RunID: runID,
				Type:  EventRunCreated,
				Payload: marshalPayload(RunCreatedPayload{
					SessionKey:      sessionKey,
					OriginalRequest: originalRequest,
					Goal:            plan.Goal,
				}),
			}); err != nil {
				return nil, fmt.Errorf("append run_created: %w", err)
			}

			// Convert plan to steps and criteria, then attach.
			steps, criteria := ConvertPlanToRunData(plan)
			if err := store.AppendJournalEvent(ctx, JournalEvent{
				RunID: runID,
				Type:  EventPlanAttached,
				Payload: marshalPayload(PlanAttachedPayload{
					Steps:              steps,
					AcceptanceCriteria: criteria,
				}),
			}); err != nil {
				return nil, fmt.Errorf("append plan_attached: %w", err)
			}

			snap, err := store.GetRunSnapshot(ctx, runID)
			if err != nil {
				return nil, fmt.Errorf("get snapshot: %w", err)
			}

			return map[string]interface{}{
				"status":     "created",
				"run_id":     runID,
				"goal":       plan.Goal,
				"step_count": len(steps),
				"summary":    snap.ToSummary(),
			}, nil
		},
	}
}

func buildRunRead(store RunLedgerStore) *agent.Tool {
	return &agent.Tool{
		Name:        "run_read",
		Description: "Read the current Run snapshot. Available to orchestrator and execution agents.",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: agent.Schema().
			Str("run_id", "The run ID to read").
			Required("run_id").
			Build(),
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			runID := toolparam.OptionalString(params, "run_id", "")
			snap, err := store.GetRunSnapshot(ctx, runID)
			if err != nil {
				return nil, fmt.Errorf("get run snapshot: %w", err)
			}
			return snap, nil
		},
	}
}

func buildRunActive(store RunLedgerStore) *agent.Tool {
	return &agent.Tool{
		Name:        "run_active",
		Description: "Get the currently active step for a run. Available to orchestrator and execution agents.",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: agent.Schema().
			Str("run_id", "The run ID to query").
			Required("run_id").
			Build(),
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			runID := toolparam.OptionalString(params, "run_id", "")
			snap, err := store.GetRunSnapshot(ctx, runID)
			if err != nil {
				return nil, fmt.Errorf("get run snapshot: %w", err)
			}

			if snap.CurrentStepID == "" {
				// Try to find next executable.
				next := snap.NextExecutableStep()
				if next == nil {
					return map[string]interface{}{
						"status":  "no_active_step",
						"run_id":  runID,
						"message": "No step is currently active or executable",
					}, nil
				}
				return map[string]interface{}{
					"status":     "next_available",
					"run_id":     runID,
					"next_step":  next,
					"run_status": snap.Status,
				}, nil
			}

			step := snap.FindStep(snap.CurrentStepID)
			return map[string]interface{}{
				"status":     "active",
				"run_id":     runID,
				"step":       step,
				"blocker":    snap.CurrentBlocker,
				"run_status": snap.Status,
			}, nil
		},
	}
}

func buildRunNote(store RunLedgerStore) *agent.Tool {
	return &agent.Tool{
		Name:        "run_note",
		Description: "Read or write a scratchpad note on a run. Available to orchestrator and execution agents.",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: agent.Schema().
			Str("run_id", "The run ID").
			Str("key", "Note key").
			Str("value", "Note value (omit to read)").
			Required("run_id", "key").
			Build(),
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			runID := toolparam.OptionalString(params, "run_id", "")
			key := toolparam.OptionalString(params, "key", "")
			value := toolparam.OptionalString(params, "value", "")
			hasValue := value != ""

			if hasValue && value != "" {
				// Write note.
				if err := store.AppendJournalEvent(ctx, JournalEvent{
					RunID: runID,
					Type:  EventNoteWritten,
					Payload: marshalPayload(NoteWrittenPayload{
						Key:   key,
						Value: value,
					}),
				}); err != nil {
					return nil, fmt.Errorf("write note: %w", err)
				}
				return map[string]interface{}{
					"status": "written",
					"key":    key,
				}, nil
			}

			// Read note.
			snap, err := store.GetRunSnapshot(ctx, runID)
			if err != nil {
				return nil, fmt.Errorf("get run snapshot: %w", err)
			}
			return map[string]interface{}{
				"key":   key,
				"value": snap.Notes[key],
			}, nil
		},
	}
}

func buildRunProposeStepResult(store RunLedgerStore, pev *PEVEngine) *agent.Tool {
	return &agent.Tool{
		Name:        "run_propose_step_result",
		Description: "Propose a step result with evidence. The execution agent cannot mark steps as complete — only propose results for PEV verification.",
		SafetyLevel: agent.SafetyLevelModerate,
		Parameters: agent.Schema().
			Str("run_id", "The run ID").
			Str("step_id", "The step ID").
			Str("result", "Summary of the work done").
			Str("evidence_json", "JSON array of evidence objects [{type, content}]").
			Required("run_id", "step_id", "result").
			Build(),
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			if err := checkRole(ctx, roleExecution); err != nil {
				return nil, err
			}

			runID := toolparam.OptionalString(params, "run_id", "")
			stepID := toolparam.OptionalString(params, "step_id", "")
			result := toolparam.OptionalString(params, "result", "")

			var evidence []Evidence
			if ejson := toolparam.OptionalString(params, "evidence_json", ""); ejson != "" {
				if err := json.Unmarshal([]byte(ejson), &evidence); err != nil {
					return nil, fmt.Errorf("parse evidence_json: %w", err)
				}
			}

			snap, err := store.GetRunSnapshot(ctx, runID)
			if err != nil {
				return nil, fmt.Errorf("get snapshot before propose: %w", err)
			}
			step := snap.FindStep(stepID)
			if step == nil {
				return nil, ErrStepNotFound
			}

			caller := toolchain.AgentNameFromContext(ctx)
			if step.OwnerAgent != "" && step.OwnerAgent != caller {
				return nil, fmt.Errorf("%w: step %q is owned by %q (caller: %q)",
					ErrAccessDenied, stepID, step.OwnerAgent, caller)
			}
			if step.Status != StepStatusInProgress {
				return nil, fmt.Errorf("step %q status is %q, expected in_progress",
					stepID, step.Status)
			}

			if err := store.AppendJournalEvent(ctx, JournalEvent{
				RunID: runID,
				Type:  EventStepResultProposed,
				Payload: marshalPayload(StepResultProposedPayload{
					StepID:   stepID,
					Result:   result,
					Evidence: evidence,
				}),
			}); err != nil {
				return nil, fmt.Errorf("append step_result_proposed: %w", err)
			}

			// Auto-trigger PEV verification.
			snap, err = store.GetRunSnapshot(ctx, runID)
			if err != nil {
				return nil, fmt.Errorf("get snapshot for PEV: %w", err)
			}
			step = snap.FindStep(stepID)

			policyReq, verifyErr := pev.Verify(ctx, runID, step)
			if verifyErr != nil {
				return nil, fmt.Errorf("PEV verify step %q: %w", stepID, verifyErr)
			}

			if policyReq != nil {
				return map[string]interface{}{
					"status":         "verification_failed",
					"run_id":         runID,
					"step_id":        stepID,
					"failure_reason": policyReq.Failure.Reason,
					"retry_count":    policyReq.RetryCount,
					"max_retries":    policyReq.MaxRetries,
					"message":        "Validation failed. Orchestrator must apply a policy decision.",
				}, nil
			}

			// Validation passed -> check run completion.
			response := map[string]interface{}{
				"status":  "verified",
				"run_id":  runID,
				"step_id": stepID,
			}

			runStatus, unmetDescs := checkRunCompletion(ctx, store, pev, runID)
			response["run_status"] = runStatus
			if len(unmetDescs) > 0 {
				response["unmet_criteria"] = unmetDescs
			}
			switch runStatus {
			case "completed":
				response["message"] = "Step verified. All steps done. Run completed."
			case "failed":
				response["message"] = "Step verified but run failed (step failures or unmet criteria)."
			default:
				response["message"] = "Step verified and completed. Run still in progress."
			}

			return response, nil
		},
	}
}

func buildRunApplyPolicy(store RunLedgerStore) *agent.Tool {
	return &agent.Tool{
		Name:        "run_apply_policy",
		Description: "Apply a policy decision to a failed step. Only the orchestrator may call this.",
		SafetyLevel: agent.SafetyLevelModerate,
		Parameters: agent.Schema().
			Str("run_id", "The run ID").
			Str("step_id", "The step ID").
			Enum("action", "Policy action", "retry", "decompose", "change_agent", "change_validator", "skip", "abort", "escalate").
			Str("reason", "Reason for this decision").
			Str("new_agent", "New agent name (for change_agent)").
			Str("new_steps_json", "JSON array of new steps (for decompose)").
			Str("new_validator_json", "JSON object for new validator (for change_validator)").
			Required("run_id", "step_id", "action", "reason").
			Build(),
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			if err := checkRole(ctx, roleOrchestrator); err != nil {
				return nil, err
			}

			runID := toolparam.OptionalString(params, "run_id", "")
			stepID := toolparam.OptionalString(params, "step_id", "")
			action := toolparam.OptionalString(params, "action", "")
			reason := toolparam.OptionalString(params, "reason", "")

			decision := PolicyDecision{
				Action: PolicyAction(action),
				Reason: reason,
			}

			if agent := toolparam.OptionalString(params, "new_agent", ""); agent != "" {
				decision.NewAgent = agent
			}
			if stepsJSON := toolparam.OptionalString(params, "new_steps_json", ""); stepsJSON != "" {
				if err := json.Unmarshal([]byte(stepsJSON), &decision.NewSteps); err != nil {
					return nil, fmt.Errorf("parse new_steps_json: %w", err)
				}
			}
			if validatorJSON := toolparam.OptionalString(params, "new_validator_json", ""); validatorJSON != "" {
				var vs ValidatorSpec
				if err := json.Unmarshal([]byte(validatorJSON), &vs); err != nil {
					return nil, fmt.Errorf("parse new_validator_json: %w", err)
				}
				decision.NewValidator = &vs
			}

			if err := store.AppendJournalEvent(ctx, JournalEvent{
				RunID: runID,
				Type:  EventPolicyDecisionApplied,
				Payload: marshalPayload(PolicyDecisionAppliedPayload{
					StepID:   stepID,
					Decision: decision,
				}),
			}); err != nil {
				return nil, fmt.Errorf("append policy_decision_applied: %w", err)
			}

			snap, err := store.GetRunSnapshot(ctx, runID)
			if err != nil {
				return nil, fmt.Errorf("get snapshot: %w", err)
			}

			return map[string]interface{}{
				"status":     "applied",
				"run_id":     runID,
				"step_id":    stepID,
				"action":     action,
				"run_status": snap.Status,
				"summary":    snap.ToSummary(),
			}, nil
		},
	}
}

func buildRunApproveStep(store RunLedgerStore, pev *PEVEngine) *agent.Tool {
	return &agent.Tool{
		Name:        "run_approve_step",
		Description: "Explicitly approve a step that requires orchestrator_approval. Only the orchestrator may call this.",
		SafetyLevel: agent.SafetyLevelModerate,
		Parameters: agent.Schema().
			Str("run_id", "The run ID").
			Str("step_id", "The step ID to approve").
			Str("reason", "Reason for approval").
			Required("run_id", "step_id").
			Build(),
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			if err := checkRole(ctx, roleOrchestrator); err != nil {
				return nil, err
			}

			runID := toolparam.OptionalString(params, "run_id", "")
			stepID := toolparam.OptionalString(params, "step_id", "")
			reason := toolparam.OptionalString(params, "reason", "")
			if reason == "" {
				reason = "orchestrator approved"
			}

			// Load snapshot and find step.
			snap, err := store.GetRunSnapshot(ctx, runID)
			if err != nil {
				return nil, fmt.Errorf("get snapshot: %w", err)
			}
			step := snap.FindStep(stepID)
			if step == nil {
				return nil, ErrStepNotFound
			}

			// Validator type must be orchestrator_approval.
			if step.Validator.Type != ValidatorOrchestratorApproval {
				return nil, fmt.Errorf("%w: run_approve_step only works for orchestrator_approval steps (this step uses %q)",
					ErrAccessDenied, step.Validator.Type)
			}

			// Step must be in verify_pending or failed status.
			// For orchestrator_approval steps, PEV auto-runs the validator which always
			// fails, transitioning the step to "failed". The orchestrator then approves it.
			if step.Status != StepStatusVerifyPending && step.Status != StepStatusFailed {
				return nil, fmt.Errorf("step %q status is %q, expected verify_pending or failed", stepID, step.Status)
			}

			// Record as validation passed.
			if err := store.RecordValidationResult(ctx, runID, stepID, ValidationResult{
				Passed: true,
				Reason: reason,
			}); err != nil {
				return nil, fmt.Errorf("record approval: %w", err)
			}

			// Check run completion.
			runStatus, unmetDescs := checkRunCompletion(ctx, store, pev, runID)

			response := map[string]interface{}{
				"status":     "approved",
				"run_id":     runID,
				"step_id":    stepID,
				"run_status": runStatus,
			}
			if len(unmetDescs) > 0 {
				response["unmet_criteria"] = unmetDescs
			}

			return response, nil
		},
	}
}

func buildRunResume(store RunLedgerStore) *agent.Tool {
	return &agent.Tool{
		Name:        "run_resume",
		Description: "Resume a paused run. Only the orchestrator may call this.",
		SafetyLevel: agent.SafetyLevelModerate,
		Parameters: agent.Schema().
			Str("run_id", "The run ID to resume").
			Required("run_id").
			Build(),
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			if err := checkRole(ctx, roleOrchestrator); err != nil {
				return nil, err
			}

			runID := toolparam.OptionalString(params, "run_id", "")
			agentName := toolchain.AgentNameFromContext(ctx)

			rm := NewResumeManager(store, time.Hour)
			snap, err := rm.Resume(ctx, runID, agentName)
			if err != nil {
				return nil, fmt.Errorf("resume run: %w", err)
			}

			return map[string]interface{}{
				"status":  "resumed",
				"run_id":  runID,
				"summary": snap.ToSummary(),
			}, nil
		},
	}
}

// checkRunCompletion checks if a run is complete and transitions its status.
// Returns the run status string and any unmet acceptance criteria descriptions.
func checkRunCompletion(ctx context.Context, store RunLedgerStore, pev *PEVEngine, runID string) (string, []string) {
	snap, err := store.GetRunSnapshot(ctx, runID)
	if err != nil {
		return "running", nil
	}

	if !snap.AllStepsSuccessful() {
		if snap.AllStepsTerminal() {
			_ = store.AppendJournalEvent(ctx, JournalEvent{
				RunID: runID,
				Type:  EventRunFailed,
				Payload: marshalPayload(RunFailedPayload{
					Reason: "one or more steps failed or interrupted",
				}),
			})
			if pev != nil {
				_ = pev.maybePruneRunHistory(ctx)
			}
			return "failed", nil
		}
		return "running", nil
	}

	// All steps successful -> verify acceptance criteria.
	beforeMet := make([]bool, len(snap.AcceptanceState))
	for i := range snap.AcceptanceState {
		beforeMet[i] = snap.AcceptanceState[i].Met
	}
	unmet, evaluated, _ := pev.VerifyAcceptanceCriteria(ctx, snap.AcceptanceState)

	// Journal newly met criteria.
	for i := range evaluated {
		if !beforeMet[i] && evaluated[i].Met {
			_ = store.AppendJournalEvent(ctx, JournalEvent{
				RunID: runID,
				Type:  EventCriterionMet,
				Payload: marshalPayload(CriterionMetPayload{
					Index:       i,
					Description: evaluated[i].Description,
				}),
			})
		}
	}

	if len(unmet) == 0 {
		_ = store.AppendJournalEvent(ctx, JournalEvent{
			RunID:   runID,
			Type:    EventRunCompleted,
			Payload: marshalPayload(RunCompletedPayload{Summary: "all steps and criteria satisfied"}),
		})
		if pev != nil {
			_ = pev.maybePruneRunHistory(ctx)
		}
		return "completed", nil
	}

	var descs []string
	for _, u := range unmet {
		descs = append(descs, u.Description)
	}
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID,
		Type:  EventRunFailed,
		Payload: marshalPayload(RunFailedPayload{
			Reason: "unmet acceptance criteria: " + strings.Join(descs, "; "),
		}),
	})
	if pev != nil {
		_ = pev.maybePruneRunHistory(ctx)
	}
	return "failed", descs
}

// checkRole verifies the caller has the required role.
// In the current implementation, orchestrator agents have "orchestrator" or
// "lango-orchestrator" as their agent name. Execution agents are everything else.
func checkRole(ctx context.Context, required callerRole) error {
	if required == roleAny {
		return nil
	}
	agentName := toolchain.AgentNameFromContext(ctx)
	if agentName == "" {
		return fmt.Errorf("%w: caller identity is required", ErrAccessDenied)
	}
	isOrchestrator := isOrchestratorAgentName(agentName)

	switch required {
	case roleOrchestrator:
		if !isOrchestrator {
			return fmt.Errorf("%w: only orchestrator can call this tool (caller: %q)", ErrAccessDenied, agentName)
		}
	case roleExecution:
		if isOrchestrator {
			return fmt.Errorf("%w: execution-only tool (caller: %q)", ErrAccessDenied, agentName)
		}
	}
	return nil
}

func isOrchestratorAgentName(agentName string) bool {
	return agentName == "orchestrator" ||
		agentName == "lango-orchestrator" ||
		agentName == SystemCallerName
}
