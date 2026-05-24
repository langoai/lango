## MODIFIED Requirements
### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Approval surfaces use consistent deny-key wording
- **WHEN** a chat approval surface renders a deny affordance
- **THEN** it SHALL label the deny keys consistently as `d/Esc`
