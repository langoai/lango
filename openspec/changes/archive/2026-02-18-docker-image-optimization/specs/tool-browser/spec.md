## MODIFIED Requirements

### Requirement: Browser automation via go-rod
The system SHALL provide browser automation tools powered by go-rod for web page interaction, with support for both local and remote browser instances.

#### Scenario: Browser navigation
- **WHEN** `browser_navigate` is called with a URL
- **THEN** the system SHALL navigate to the URL, wait for page load, and return title, URL, and text snippet

#### Scenario: Implicit session management
- **WHEN** any browser tool is called without a prior session
- **THEN** the system SHALL auto-create a browser session and reuse it for subsequent calls
- **AND** the LLM SHALL NOT need to manage session IDs

#### Scenario: Remote browser configuration
- **WHEN** `RemoteBrowserURL` is set in browser tool config
- **THEN** the browser tool SHALL store the URL in its `Config.RemoteBrowserURL` field
- **AND** the value SHALL be wired from `BrowserToolConfig.RemoteBrowserURL` in app initialization
