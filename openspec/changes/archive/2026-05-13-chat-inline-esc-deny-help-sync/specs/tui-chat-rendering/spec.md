## MODIFIED Requirements
### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Inline approval strip advertises both deny keys
- **WHEN** the inline approval strip is rendered
- **THEN** its deny affordance SHALL mention both `d` and `Esc`
