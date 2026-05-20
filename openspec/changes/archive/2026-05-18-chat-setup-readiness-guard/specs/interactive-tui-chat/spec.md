## MODIFIED Requirements

### Requirement: Interactive TUI chat on bare invocation
Running `lango chat` SHALL start the interactive terminal chat session using Bubble Tea. `lango serve` SHALL continue to work as the full gateway plus channels mode. Slice 6 SHALL remove the older bare-`lango` chat interpretation from this surface contract.

#### Scenario: Focused chat shows setup-required state for incomplete profiles
- **WHEN** the active config does not satisfy `config.EvaluateAgentSetup`
- **THEN** focused chat SHALL render setup-required state instead of a ready/send state
- **AND** the setup guidance SHALL mention `lango onboard`, `lango settings`, and `lango doctor`

#### Scenario: Focused chat blocks normal turns until setup is ready
- **WHEN** the active config does not satisfy `config.EvaluateAgentSetup`
- **AND** the user submits non-slash input in focused chat
- **THEN** focused chat SHALL NOT call the turn runner
- **AND** focused chat SHALL keep the draft input available
- **AND** focused chat SHALL append actionable setup guidance

#### Scenario: Focused chat keeps slash commands available before setup
- **WHEN** the active config does not satisfy `config.EvaluateAgentSetup`
- **AND** the user submits a slash command in focused chat
- **THEN** focused chat SHALL dispatch the slash command instead of blocking it as a normal turn

#### Scenario: Focused chat submits normally after setup is ready
- **WHEN** the active config satisfies `config.EvaluateAgentSetup`
- **AND** the user submits non-slash input in focused chat
- **THEN** focused chat SHALL call the turn runner as before
