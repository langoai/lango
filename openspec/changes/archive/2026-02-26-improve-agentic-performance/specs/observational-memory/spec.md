## ADDED Requirements

### Requirement: Memory token budgeting in context assembly
The `ContextAwareModelAdapter` SHALL enforce a token budget when assembling the memory section into the system prompt. Reflections SHALL be included first (higher information density), then observations fill the remaining budget.

#### Scenario: Default memory token budget
- **WHEN** no explicit budget is configured via `WithMemoryTokenBudget`
- **THEN** the default budget SHALL be 4000 tokens

#### Scenario: Reflections exceed budget
- **WHEN** reflections alone exceed the token budget
- **THEN** the system SHALL include reflections up to the budget limit and skip all observations

#### Scenario: Budget shared between reflections and observations
- **WHEN** reflections use part of the budget
- **THEN** observations SHALL fill the remaining budget, stopping when the next observation would exceed it

#### Scenario: Custom budget via WithMemoryTokenBudget
- **WHEN** `WithMemoryTokenBudget(budget)` is called with a positive value
- **THEN** the adapter SHALL use that budget instead of the default 4000

### Requirement: Auto meta-reflection on accumulation
The `memory.Buffer` SHALL automatically trigger meta-reflection when the number of reflections in a session exceeds a configurable consolidation threshold.

#### Scenario: Default consolidation threshold
- **WHEN** no explicit threshold is configured
- **THEN** the default threshold SHALL be 5 reflections

#### Scenario: Meta-reflection triggered
- **WHEN** `process()` completes and the session has >= threshold reflections
- **THEN** `ReflectOnReflections` SHALL be called to consolidate them

#### Scenario: Meta-reflection failure is non-fatal
- **WHEN** `ReflectOnReflections` returns an error
- **THEN** the system SHALL log the error and continue normal operation
