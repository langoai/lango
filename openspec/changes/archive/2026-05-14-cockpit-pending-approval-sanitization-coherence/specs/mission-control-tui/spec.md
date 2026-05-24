## MODIFIED Requirements

### Requirement: The latest pending approval renders as one live decision
Mission Control SHALL render the latest pending approval request as one live decision using the shared pending approval owner. Resolving the decision SHALL write to the original approval response channel and remove the pending item from other cockpit surfaces on the next render.

#### Scenario: Shared pending approval snapshot text is replay-safe
- **WHEN** tool names, request summaries, rule explanations, or risk labels contain ANSI/OSC escape sequences or embedded newlines before entering the shared pending approval owner
- **THEN** the cockpit pending approval registry SHALL strip those control sequences
- **AND** it SHALL normalize the stored pending approval text to a single line before replay
