## MODIFIED Requirements

### Requirement: Approval surfaces keep chat rendering stable
The chat TUI SHALL keep approval rendering stable across inline and fullscreen surfaces, including defensive fallback paths for internal renderer state.

#### Scenario: Approval params render in stable key order
- **WHEN** an approval surface renders request params from a map-backed payload
- **THEN** it SHALL render those params in deterministic key order instead of raw map iteration order
