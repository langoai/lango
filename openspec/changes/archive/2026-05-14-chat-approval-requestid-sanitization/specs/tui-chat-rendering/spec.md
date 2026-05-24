## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Approval request-id annotations stay plain and single-line
- **WHEN** an approval transcript event renders a request ID containing ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the displayed request-id annotation to a single line before compacting it
