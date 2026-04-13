## ADDED Requirements

### Requirement: Exec tool middleware execution order includes WithPolicy
The exec tool middleware chain SHALL include `WithPolicy` as the outermost middleware, applied after `WithApproval` in `app.go`. The execution order SHALL be: `WithPolicy → WithApproval → WithPrincipal → WithHooks → WithOutputManager → WithLearning → Handler`.

#### Scenario: Policy middleware applied as outermost
- **WHEN** the tool middleware chain is built in `app.go`
- **THEN** `WithPolicy` SHALL be applied after `WithApproval` (making it outermost)
- **AND** for `exec`/`exec_bg` tools, the policy evaluation SHALL run before approval

#### Scenario: Blocked command does not reach approval
- **WHEN** `WithPolicy` blocks a command
- **THEN** the approval middleware SHALL NOT be invoked
- **AND** `BlockedResult{Blocked: true}` SHALL be returned directly

### Requirement: Handler guards serve as defense-in-depth fallback
The existing handler guards (`langoGuard`, `pathGuard`) in `BuildTools` SHALL be preserved unchanged. They serve as a deterministic-only fallback that checks a strict subset of what the middleware checks.

#### Scenario: Handler guards remain operational
- **WHEN** `exec` handler receives a command that passes `WithPolicy` middleware
- **THEN** the existing `checkGuards` closure SHALL still run `langoGuard` then `pathGuard`
- **AND** a command blocked by handler guards is also blocked by the middleware (invariant)

### Requirement: PolicyDecisionEvent added to event bus
The `internal/eventbus/events.go` SHALL define `EventPolicyDecision` constant and `PolicyDecisionEvent` struct. The event SHALL carry command, unwrapped command, verdict string, reason code, message, session key, and agent name.

#### Scenario: PolicyDecisionEvent follows existing event pattern
- **WHEN** `PolicyDecisionEvent` is defined
- **THEN** it SHALL implement the `Event` interface with `EventName()` returning `"policy.decision"`

### Requirement: PolicyEvaluator created in app.go Phase B
The `PolicyEvaluator` SHALL be created in `app.go` Phase B using `fv.CmdGuard` and `fv.AutoAvail` from the resolved `foundationValues`. The event bus SHALL be passed as nil when `cfg.Hooks.EventPublishing` is disabled.

#### Scenario: Evaluator receives nil bus when publishing disabled
- **WHEN** `cfg.Hooks.EventPublishing` is false
- **THEN** the PolicyEvaluator SHALL receive a nil bus
- **AND** no `PolicyDecisionEvent` SHALL be published for any verdict
