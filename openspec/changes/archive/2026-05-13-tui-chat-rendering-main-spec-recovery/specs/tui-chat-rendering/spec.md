## MODIFIED Requirements
### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Streaming state visible
- **WHEN** the agent begins generating a response
- **THEN** the turn status strip SHALL show that generation is in progress and cancellation is available

#### Scenario: Approval state visible
- **WHEN** a tool approval request interrupts the current turn
- **THEN** the turn status strip SHALL show that approval is required

#### Scenario: Failed state visible
- **WHEN** a turn ends in failure without producing a successful completion
- **THEN** the turn status strip SHALL show a failed state until the next user interaction resets it

#### Scenario: Idle and failed help describe double-press quit
- **WHEN** the chat help bar is rendered in the `idle` or `failed` states
- **THEN** the `Ctrl+C` binding SHALL describe quitting via the double-press path rather than immediate single-press exit

#### Scenario: Idle and failed help advertise immediate Ctrl+D quit
- **WHEN** the chat help bar is rendered in the `idle` or `failed` states
- **THEN** it SHALL advertise `Ctrl+D` as the immediate quit path

#### Scenario: Chat help uses unified transcript-scroll wording
- **WHEN** a user reads chat help for `PgUp/PgDn`
- **THEN** the scrolling action SHALL be described with one consistent transcript-scroll phrase across in-product and public docs

#### Scenario: Approval surfaces use unified session-allow wording
- **WHEN** a chat approval action bar is rendered for either inline or fullscreen approval UI
- **THEN** the `s` action SHALL be labeled `allow session`

#### Scenario: Approval surfaces use consistent deny-key wording
- **WHEN** a chat approval surface renders a deny affordance
- **THEN** it SHALL label the deny keys consistently as `d/Esc`

#### Scenario: Approval confirm prompt reflects the pending action key
- **WHEN** a critical-risk approval surface is in confirm-pending state
- **THEN** the visible confirm prompt SHALL name the actual pending action key (`a` or `s`) rather than a hard-coded default

#### Scenario: Approval confirm prompt keeps deny path visible
- **WHEN** a critical-risk approval surface is in confirm-pending state
- **THEN** the visible confirm prompt SHALL mention that `d` or `Esc` still denies the request

#### Scenario: Fullscreen approval dialog surfaces split-toggle help when diff exists
- **WHEN** the fullscreen approval dialog renders diff content
- **THEN** its action bar SHALL advertise the `t` key for toggling split mode

#### Scenario: Fullscreen approval dialog shows scroll help only when diff overflow exists
- **WHEN** the fullscreen approval dialog renders diff content that exceeds the visible diff area
- **THEN** its action bar SHALL advertise `↑/k` and `↓/j` scrolling

#### Scenario: Fullscreen approval dialog hides inert scroll help when diff fits
- **WHEN** the fullscreen approval dialog renders diff content that fits within the visible diff area
- **THEN** its action bar SHALL omit `↑/k` and `↓/j` scrolling

#### Scenario: Fullscreen approval diff clamps scroll offset when the diff fits
- **WHEN** the fullscreen approval dialog renders a diff that fits within the visible diff area
- **THEN** its scroll offset SHALL clamp to zero
- **AND** the full short diff SHALL remain visible
