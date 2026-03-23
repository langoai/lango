## ADDED Requirements

### Requirement: Child summary preserves typed incomplete cause
When an isolated child session ends without a visible assistant summary, the parent-visible note SHALL preserve the classified incomplete cause instead of using a generic placeholder.

#### Scenario: Empty-after-tool-use note is explicit
- **WHEN** an isolated specialist ends without visible assistant completion after tool activity
- **THEN** the parent-visible note SHALL mention `empty_after_tool_use`
- **AND** it SHALL NOT promote raw tool output to a success summary
