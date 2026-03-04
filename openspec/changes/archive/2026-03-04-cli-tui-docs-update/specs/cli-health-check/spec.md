## MODIFIED Requirements

### Requirement: Advanced feature hints in onboard flow
The onboard flow SHALL display hints about advanced features after initial setup is complete. The hints SHALL inform users about agent memory, hooks, librarian, and learning system features that can be configured via settings or CLI.

#### Scenario: Onboard completion hints
- **WHEN** user completes the onboard wizard successfully
- **THEN** system displays hints mentioning:
  - Agent memory configuration via `lango memory agents` or TUI settings
  - Hook system configuration via `lango agent hooks` or TUI settings
  - Librarian configuration via `lango librarian status`

### Requirement: Feature discovery in doctor output
The doctor command output SHALL include brief hints about new CLI commands when relevant checks pass or are skipped, to aid feature discovery.

#### Scenario: Graph check with hint
- **WHEN** GraphStoreCheck returns StatusSkip because graph is disabled
- **THEN** the check message SHALL mention that graph can be enabled and managed via `lango graph` commands

#### Scenario: Multi-agent check with hint
- **WHEN** MultiAgentCheck returns StatusSkip because multi-agent is disabled
- **THEN** the check message SHALL mention that multi-agent can be configured via settings

### Requirement: Existing onboard flow unaffected
The addition of feature hints SHALL NOT change the core onboard flow steps or validation logic. Hints are displayed only after successful completion.

#### Scenario: Onboard steps unchanged
- **WHEN** user runs `lango onboard`
- **THEN** all existing onboard steps (provider selection, API key, channel setup) function identically to before the hint additions
