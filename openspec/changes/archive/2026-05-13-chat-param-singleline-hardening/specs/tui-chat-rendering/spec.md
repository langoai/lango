## MODIFIED Requirements

### Requirement: Approval surfaces keep chat rendering stable
The chat TUI SHALL keep approval rendering stable across inline and fullscreen surfaces, including defensive fallback paths for internal renderer state.

#### Scenario: Param display text stays single-line safe
- **WHEN** an approval surface or tool lifecycle preview renders a string param containing line breaks
- **THEN** it SHALL collapse that value into single-line-safe display text instead of introducing layout-breaking new lines
