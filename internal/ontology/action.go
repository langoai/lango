package ontology

import (
	"context"
	"fmt"
	"sync"

	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/graph"
)

// ActionType defines a reusable transactional operation on the ontology.
// ActionTypes are in-process Go closures registered at startup — NOT persisted DSL.
type ActionType struct {
	Name        string
	Description string
	// RequiredPerm is the permission the executor checks before running.
	// INVARIANT: must be >= max permission of any OntologyService method
	// called by Execute or Compensate. Violating this causes "executor
	// passed but service rejected" partial failures.
	RequiredPerm Permission
	// ParamSchema maps parameter names to descriptions (used for tool generation).
	ParamSchema map[string]string
	// Precondition validates whether the action can be executed.
	// Returns nil if preconditions are met.
	Precondition func(ctx context.Context, svc OntologyService, params map[string]string) error
	// Execute performs the action and returns its effects.
	Execute func(ctx context.Context, svc OntologyService, params map[string]string) (*ActionEffects, error)
	// Compensate reverses the effects on Execute failure. May be nil.
	Compensate func(ctx context.Context, svc OntologyService, effects *ActionEffects) error
}

// ActionRegistry stores registered ActionTypes by name.
type ActionRegistry struct {
	mu      sync.RWMutex
	actions map[string]*ActionType
}

// NewActionRegistry creates a new empty ActionRegistry.
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{actions: make(map[string]*ActionType)}
}

// Register adds an ActionType to the registry. Returns error if name is duplicate.
func (r *ActionRegistry) Register(a *ActionType) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.actions[a.Name]; exists {
		return fmt.Errorf("action %q already registered", a.Name)
	}
	r.actions[a.Name] = a
	return nil
}

// Get retrieves an ActionType by name.
func (r *ActionRegistry) Get(name string) (*ActionType, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.actions[name]
	return a, ok
}

// List returns all registered ActionTypes.
func (r *ActionRegistry) List() []*ActionType {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*ActionType, 0, len(r.actions))
	for _, a := range r.actions {
		result = append(result, a)
	}
	return result
}

// ActionExecutor orchestrates action execution with ACL, preconditions, logging, and compensation.
type ActionExecutor struct {
	registry *ActionRegistry
	svc      OntologyService
	acl      ACLPolicy
	logStore *ActionLogStore
}

// NewActionExecutor creates a new ActionExecutor.
func NewActionExecutor(svc OntologyService, reg *ActionRegistry, acl ACLPolicy, logStore *ActionLogStore) *ActionExecutor {
	return &ActionExecutor{
		registry: reg,
		svc:      svc,
		acl:      acl,
		logStore: logStore,
	}
}

// Execute runs the named action through the full lifecycle:
// ACL check → Precondition → Log(started) → Execute → Log(completed/failed/compensated).
func (e *ActionExecutor) Execute(ctx context.Context, actionName string, params map[string]string) (*ActionResult, error) {
	action, ok := e.registry.Get(actionName)
	if !ok {
		return nil, fmt.Errorf("action %q not found", actionName)
	}

	// 1. ACL check
	principal := ctxkeys.PrincipalFromContext(ctx)
	if principal == "" {
		principal = "system"
	}
	if e.acl != nil {
		if err := e.acl.Check(principal, action.RequiredPerm); err != nil {
			return nil, err
		}
	}

	// 2. Precondition
	if action.Precondition != nil {
		if err := action.Precondition(ctx, e.svc, params); err != nil {
			return nil, fmt.Errorf("precondition failed: %w", err)
		}
	}

	// 3. Create log entry
	logID, err := e.logStore.Create(ctx, actionName, principal, params)
	if err != nil {
		return nil, fmt.Errorf("create action log: %w", err)
	}

	// 4. Execute
	effects, execErr := action.Execute(ctx, e.svc, params)
	if execErr != nil {
		// 5. Compensate on failure
		status := ActionFailed
		if action.Compensate != nil && effects != nil {
			if compErr := action.Compensate(ctx, e.svc, effects); compErr == nil {
				status = ActionCompensated
				_ = e.logStore.Compensated(ctx, logID)
			} else {
				_ = e.logStore.Fail(ctx, logID, fmt.Sprintf("execute: %v; compensate: %v", execErr, compErr))
			}
		}
		if status == ActionFailed {
			_ = e.logStore.Fail(ctx, logID, execErr.Error())
		}
		return &ActionResult{
			LogID:  logID,
			Status: status,
			Error:  execErr.Error(),
		}, nil
	}

	// 6. Success
	_ = e.logStore.Complete(ctx, logID, effects)
	return &ActionResult{
		LogID:   logID,
		Status:  ActionCompleted,
		Effects: effects,
	}, nil
}

// --- Built-in Actions ---

// BuiltinLinkEntities returns an action that asserts a fact between two entities.
func BuiltinLinkEntities() *ActionType {
	return &ActionType{
		Name:         "link_entities",
		Description:  "Assert a fact linking two entities via a predicate",
		RequiredPerm: PermWrite, // calls AssertFact (PermWrite)
		ParamSchema: map[string]string{
			"subject":   "Subject entity ID",
			"predicate": "Predicate name",
			"object":    "Object entity ID",
			"source":    "Source identifier (e.g., 'manual', 'llm_extraction')",
		},
		Precondition: func(ctx context.Context, svc OntologyService, params map[string]string) error {
			for _, key := range []string{"subject", "predicate", "object", "source"} {
				if params[key] == "" {
					return fmt.Errorf("missing required parameter: %s", key)
				}
			}
			// Validate predicate exists
			pred, err := svc.GetPredicate(ctx, params["predicate"])
			if err != nil {
				return fmt.Errorf("invalid predicate %q: %w", params["predicate"], err)
			}
			if pred.Status != SchemaActive {
				return fmt.Errorf("predicate %q is not active", params["predicate"])
			}
			return nil
		},
		Execute: func(ctx context.Context, svc OntologyService, params map[string]string) (*ActionEffects, error) {
			_, err := svc.AssertFact(ctx, AssertionInput{
				Triple: graph.Triple{
					Subject:   params["subject"],
					Predicate: params["predicate"],
					Object:    params["object"],
				},
				Source: params["source"],
			})
			if err != nil {
				return nil, err
			}
			return &ActionEffects{
				FactsAsserted: []FactEffect{{
					Subject:   params["subject"],
					Predicate: params["predicate"],
					Object:    params["object"],
				}},
			}, nil
		},
		Compensate: func(ctx context.Context, svc OntologyService, effects *ActionEffects) error {
			for _, f := range effects.FactsAsserted {
				if err := svc.RetractFact(ctx, f.Subject, f.Predicate, f.Object, "action_compensation"); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

// BuiltinSetEntityStatus returns an action that sets an entity's status property.
func BuiltinSetEntityStatus() *ActionType {
	return &ActionType{
		Name:         "set_entity_status",
		Description:  "Set the status property of an entity",
		RequiredPerm: PermWrite, // calls SetEntityProperty (PermWrite)
		ParamSchema: map[string]string{
			"entity_id":   "Entity ID",
			"entity_type": "ObjectType name",
			"status":      "New status value",
		},
		Precondition: func(ctx context.Context, svc OntologyService, params map[string]string) error {
			for _, key := range []string{"entity_id", "entity_type", "status"} {
				if params[key] == "" {
					return fmt.Errorf("missing required parameter: %s", key)
				}
			}
			// Validate entity type exists
			_, err := svc.GetType(ctx, params["entity_type"])
			if err != nil {
				return fmt.Errorf("invalid entity type %q: %w", params["entity_type"], err)
			}
			return nil
		},
		Execute: func(ctx context.Context, svc OntologyService, params map[string]string) (*ActionEffects, error) {
			// Get current value for effects tracking
			props, _ := svc.GetEntityProperties(ctx, params["entity_id"])
			oldValue := ""
			if props != nil {
				oldValue = props["status"]
			}
			if err := svc.SetEntityProperty(ctx, params["entity_id"], params["entity_type"], "status", params["status"]); err != nil {
				return &ActionEffects{
					PropertiesSet: []PropertyEffect{{
						EntityID: params["entity_id"],
						Property: "status",
						OldValue: oldValue,
						NewValue: params["status"],
					}},
				}, err
			}
			return &ActionEffects{
				PropertiesSet: []PropertyEffect{{
					EntityID: params["entity_id"],
					Property: "status",
					OldValue: oldValue,
					NewValue: params["status"],
				}},
			}, nil
		},
		// No compensation — status change is not easily reversible without more context
	}
}
