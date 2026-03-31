## ADDED Requirements

### Requirement: ActionType definition
The system SHALL define an `ActionType` struct with `Name`, `Description`, `RequiredPerm` (Permission), `ParamSchema` (map of param name to description), and three closure fields: `Precondition`, `Execute`, `Compensate` (may be nil).

#### Scenario: ActionType with all fields
- **WHEN** an ActionType is created with Name, RequiredPerm, Precondition, Execute, and Compensate
- **THEN** all fields SHALL be accessible and the action SHALL be registerable

#### Scenario: ActionType with nil Compensate
- **WHEN** an ActionType is created with Compensate set to nil
- **THEN** the action SHALL be valid and executor SHALL skip compensation on failure

### Requirement: RequiredPerm invariant
`ActionType.RequiredPerm` MUST be >= the maximum permission required by any OntologyService method called within Execute or Compensate. This prevents "executor passed but service rejected" partial failures.

#### Scenario: RequiredPerm matches inner methods
- **WHEN** an action's Execute calls AssertFact (PermWrite) and SetEntityProperty (PermWrite)
- **THEN** RequiredPerm SHALL be at least PermWrite

#### Scenario: RequiredPerm too low causes inner rejection
- **WHEN** an action has RequiredPerm=PermRead but Execute calls AssertFact (PermWrite)
- **THEN** the executor ACL check passes but the inner AssertFact call SHALL return ErrPermissionDenied

### Requirement: ActionRegistry
The system SHALL provide an `ActionRegistry` with `Register(action)`, `Get(name)`, and `List()` methods. Duplicate registration (same name) SHALL return an error.

#### Scenario: Register and retrieve action
- **WHEN** an ActionType named "link_entities" is registered
- **THEN** `Get("link_entities")` SHALL return that action

#### Scenario: Duplicate registration
- **WHEN** an ActionType named "link_entities" is registered twice
- **THEN** the second registration SHALL return an error

#### Scenario: List all actions
- **WHEN** 2 actions are registered
- **THEN** `List()` SHALL return both actions

### Requirement: ActionExecutor lifecycle
The ActionExecutor SHALL orchestrate action execution in this order: (1) look up action from registry, (2) ACL check via directly-injected ACLPolicy, (3) run Precondition, (4) create ActionLog with status "started", (5) run Execute, (6) on success update ActionLog to "completed", (7) on Execute failure run Compensate if non-nil then update ActionLog to "compensated" or "failed".

#### Scenario: Successful execution
- **WHEN** Execute returns successfully with an ActionExecutor that has ACL allowing the principal
- **THEN** ActionLog status SHALL be "completed" and ActionResult.Effects SHALL contain the effects

#### Scenario: Precondition failure
- **WHEN** Precondition returns an error
- **THEN** Execute SHALL NOT run, no ActionLog SHALL be created, and error SHALL be returned

#### Scenario: Execute failure with compensation
- **WHEN** Execute fails and Compensate is non-nil
- **THEN** Compensate SHALL run and ActionLog status SHALL be "compensated"

#### Scenario: Execute failure without compensation
- **WHEN** Execute fails and Compensate is nil
- **THEN** ActionLog status SHALL be "failed"

#### Scenario: ACL denied
- **WHEN** the principal lacks RequiredPerm
- **THEN** the executor SHALL return ErrPermissionDenied without running Precondition or Execute

### Requirement: ActionLog persistence
The system SHALL persist action execution records in an Ent-backed `action_log` schema with fields: id (UUID), action_name, principal, params (JSON), status (enum: started/completed/failed/compensated), effects (JSON), error_message (optional), started_at, completed_at (optional).

#### Scenario: Log created on execution start
- **WHEN** an action passes ACL and precondition checks
- **THEN** an ActionLog record SHALL be created with status "started"

#### Scenario: Log updated on completion
- **WHEN** action execution completes successfully
- **THEN** the ActionLog record SHALL be updated to status "completed" with effects and completed_at

### Requirement: Built-in actions
The system SHALL provide 2 built-in actions registered at startup: `link_entities` (RequiredPerm=PermWrite, calls AssertFact) and `set_entity_status` (RequiredPerm=PermWrite, calls SetEntityProperty).

#### Scenario: link_entities success
- **WHEN** `link_entities` is executed with valid subject, predicate, object, source params
- **THEN** it SHALL assert a fact via OntologyService.AssertFact and return FactsAsserted in effects

#### Scenario: set_entity_status success
- **WHEN** `set_entity_status` is executed with valid entity_id, entity_type, status params
- **THEN** it SHALL set the property via OntologyService.SetEntityProperty and return PropertiesSet in effects

### Requirement: OntologyService action methods
OntologyService interface SHALL be extended with 4 methods: `ExecuteAction(ctx, actionName, params)`, `ListActions(ctx)`, `GetActionLog(ctx, logID)`, `ListActionLogs(ctx, actionName, limit)`.

#### Scenario: ExecuteAction delegates to executor
- **WHEN** `ExecuteAction` is called with a valid action name
- **THEN** it SHALL delegate to the ActionExecutor and return the ActionResult

#### Scenario: ListActions returns summaries
- **WHEN** `ListActions` is called
- **THEN** it SHALL return ActionSummary entries for all registered actions
