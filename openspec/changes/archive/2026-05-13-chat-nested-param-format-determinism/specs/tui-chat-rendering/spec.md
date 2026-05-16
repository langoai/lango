## MODIFIED Requirements

### Requirement: Approval surfaces keep chat rendering stable
The chat TUI SHALL keep approval rendering stable across inline and fullscreen surfaces, including defensive fallback paths for internal renderer state.

#### Scenario: Structured param values render deterministically
- **WHEN** an approval surface or tool lifecycle preview renders nested structured param values
- **THEN** it SHALL format those values deterministically instead of relying on raw map iteration output
