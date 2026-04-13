## ADDED Requirements

### Requirement: Structured recovery avoids repeating a failed specialist
When structured orchestration handles a specialist failure, the recovery layer SHALL use the observed specialist identity to avoid immediately repeating the same failed specialist path.

#### Scenario: Specialist tool failure reroutes away from failed specialist
- **WHEN** the orchestrator delegates to `vault`, the specialist attempt fails with a tool error, and structured recovery is enabled
- **THEN** the recovery layer SHALL retry with a reroute hint instead of a blind same-input retry
- **AND** the hint SHALL identify `vault` as the failed specialist
- **AND** the hint SHALL instruct the orchestrator to choose a different specialist or answer directly

#### Scenario: Pre-specialist failure keeps generic retry behavior
- **WHEN** structured recovery handles a retryable failure before any specialist delegation is observed
- **THEN** the recovery layer MAY retry the same input
- **AND** it SHALL not fabricate a failed specialist identity
