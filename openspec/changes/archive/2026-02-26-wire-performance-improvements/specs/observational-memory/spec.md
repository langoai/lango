## MODIFIED Requirements

### Requirement: Configurable memory token budget
The wiring layer SHALL pass `observationalMemory.memoryTokenBudget` to `ContextAwareModelAdapter.WithMemoryTokenBudget()` when the value is greater than 0.

#### Scenario: Custom memory token budget
- **WHEN** config sets `observationalMemory.memoryTokenBudget: 6000`
- **THEN** the memory section in the system prompt SHALL be capped at 6000 tokens

#### Scenario: Default memory token budget
- **WHEN** config omits `observationalMemory.memoryTokenBudget`
- **THEN** the default (4000 tokens) SHALL be used

### Requirement: Configurable reflection consolidation threshold
The wiring layer SHALL call `Buffer.SetReflectionConsolidationThreshold()` when `observationalMemory.reflectionConsolidationThreshold` is greater than 0.

#### Scenario: Custom consolidation threshold
- **WHEN** config sets `observationalMemory.reflectionConsolidationThreshold: 3`
- **THEN** meta-reflection SHALL trigger after 3 reflections accumulate

#### Scenario: Default consolidation threshold
- **WHEN** config omits `observationalMemory.reflectionConsolidationThreshold`
- **THEN** the default threshold (5) SHALL be used
