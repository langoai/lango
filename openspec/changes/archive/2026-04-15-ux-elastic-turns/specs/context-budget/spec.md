## MODIFIED Requirements

### Requirement: Graceful degradation on zero or negative available budget
When `available <= 0` (model window too small or base prompt too large), the `ContextBudgetManager` SHALL return unlimited (0) budgets for all sections, effectively disabling budget enforcement. The system SHALL log a warning. The `Degraded` flag SHALL NOT be used as an emergency compaction trigger — it indicates a configuration issue (base prompt too large for the model window) that session message compaction cannot resolve.

#### Scenario: Zero available budget triggers degraded mode
- **WHEN** the model window minus response reserve minus base prompt tokens is zero or negative
- **THEN** the system SHALL return unlimited (0) budgets for all sections
- **AND** SHALL set the `Degraded` flag to `true`
- **AND** SHALL log a warning

#### Scenario: Degraded flag does not trigger compaction
- **WHEN** `budgets.Degraded` is `true`
- **THEN** the `ContextAwareModelAdapter` SHALL NOT invoke session message compaction
- **AND** SHALL log a warning that the model window is too small for the current base prompt configuration
