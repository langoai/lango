## MODIFIED Requirements

### Requirement: Context Engineering doctor check
`AllChecks()` SHALL include a `ContextHealthCheck` that reports aggregated context subsystem status. The check SHALL be registered before individual subsystem checks (Embedding, Graph, Memory, Librarian).

#### Scenario: Balanced profile with silent disabled embedding
- **WHEN** `contextProfile: balanced` is set AND embedding provider is not configured
- **THEN** doctor check reports StatusWarn with message including "1 silently disabled" and suggestion to configure embedding provider

#### Scenario: All context subsystems healthy
- **WHEN** `contextProfile: full` is set AND all subsystems initialize successfully
- **THEN** doctor check reports StatusPass with message summarizing active profile and enabled subsystem count

#### Scenario: No profile set and nothing enabled
- **WHEN** `contextProfile` is not set AND no context subsystems are enabled
- **THEN** doctor check reports StatusSkip or StatusWarn suggesting user set a context profile
