## MODIFIED Requirements
### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Approval surfaces use unified session-allow wording
- **WHEN** a chat approval action bar is rendered for either inline or fullscreen approval UI
- **THEN** the `s` action SHALL be labeled `allow session`
