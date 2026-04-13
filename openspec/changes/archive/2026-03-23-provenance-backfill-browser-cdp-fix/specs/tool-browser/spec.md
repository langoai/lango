## MODIFIED Requirements

### Requirement: Browser recovery middleware handles CDP target errors
The `WithBrowserRecovery` middleware SHALL recover from CDP target errors for `browser_navigate` by resetting the browser session and retrying once with a fresh page.

#### Scenario: CDP target error on browser_navigate triggers retry
- **WHEN** `browser_navigate` returns an error containing "Inspected target navigated or closed"
- **THEN** the middleware SHALL close the browser session
- **AND** retry the navigation once with a fresh session
- **AND** if the retry also fails, return the error as-is

#### Scenario: CDP target error on browser_action does NOT trigger retry
- **WHEN** `browser_action` returns an error containing "Inspected target navigated or closed"
- **THEN** the middleware SHALL NOT retry the action
- **AND** the error SHALL be returned as-is
