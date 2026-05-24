## ADDED Requirements

### Requirement: Bare root docs describe non-interactive fallback
Public CLI documentation that describes bare `lango` SHALL state that the workbench launch requires an interactive terminal and that non-interactive bare-root execution prints help instead of starting the TUI.

#### Scenario: Public bare root docs distinguish interactive and non-interactive behavior
- **WHEN** a user reads README, `docs/cli/index.md`, or `docs/cli/core.md`
- **THEN** the document SHALL state that bare `lango` launches the mission workbench in an interactive terminal
- **AND** the document SHALL state that non-interactive bare `lango` prints help and exits successfully without starting the TUI
- **AND** the document SHALL distinguish this fallback from `lango cockpit` and `lango chat` non-interactive errors
