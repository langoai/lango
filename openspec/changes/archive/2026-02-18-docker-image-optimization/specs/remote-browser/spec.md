## ADDED Requirements

### Requirement: Remote browser WebSocket connection
The system SHALL support connecting to a remote browser instance via WebSocket URL instead of launching a local browser.

#### Scenario: Config-based remote browser
- **WHEN** `BrowserToolConfig.RemoteBrowserURL` is set (e.g., `ws://chrome:9222`)
- **THEN** the browser tool SHALL connect to the remote browser via `rod.New().ControlURL(url).Connect()`
- **AND** the system SHALL NOT attempt to launch a local browser

#### Scenario: Environment variable fallback
- **WHEN** `RemoteBrowserURL` is not set in config but `ROD_BROWSER_WS` environment variable is present
- **THEN** the system SHALL use the environment variable value as the WebSocket URL
- **AND** the system SHALL connect to the remote browser

#### Scenario: Local browser fallback
- **WHEN** neither `RemoteBrowserURL` config nor `ROD_BROWSER_WS` env var is set
- **THEN** the system SHALL fall back to the existing local browser launch behavior

#### Scenario: Remote connection failure
- **WHEN** the remote browser WebSocket URL is set but connection fails
- **THEN** the system SHALL return an error with message containing "connect remote browser"
