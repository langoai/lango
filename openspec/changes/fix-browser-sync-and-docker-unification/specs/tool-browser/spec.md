## MODIFIED Requirements

### Requirement: Browser automation via go-rod
The system SHALL provide browser automation tools powered by go-rod for web page interaction, with local browser launch support.

#### Scenario: Browser navigation
- **WHEN** `browser_navigate` is called with a URL
- **THEN** the system SHALL navigate to the URL, wait for page load, and return title, URL, and text snippet

#### Scenario: Implicit session management
- **WHEN** any browser tool is called without a prior session
- **THEN** the system SHALL auto-create a browser session and reuse it for subsequent calls
- **AND** the LLM SHALL NOT need to manage session IDs

#### Scenario: Thread-safe browser initialization
- **WHEN** multiple browser tool calls are made concurrently
- **THEN** the system SHALL use a `sync.Mutex` + `bool` guard pattern for initialization
- **AND** only one initialization attempt SHALL execute at a time
- **AND** subsequent concurrent calls SHALL wait for and share the result

#### Scenario: Retry on initialization failure
- **WHEN** browser initialization fails (e.g., Chromium not found)
- **THEN** the `initDone` flag SHALL remain false
- **AND** the next browser tool call SHALL retry initialization

#### Scenario: No partial initialization
- **WHEN** `Connect()` fails during browser initialization
- **THEN** the browser field SHALL remain nil
- **AND** subsequent calls SHALL NOT observe a non-nil but disconnected browser

#### Scenario: Re-initialization after close
- **WHEN** `Close()` is called and browser resources are cleaned up
- **THEN** the `initDone` flag SHALL be reset to false under `initMu`
- **AND** the next browser tool call SHALL re-initialize from scratch

## REMOVED Requirements

### Requirement: Remote browser configuration
**Reason**: Remote browser support removed. Single Docker image always includes Chromium, eliminating the need for remote WebSocket connections.
**Migration**: Remove `tools.browser.remoteBrowserUrl` from config. Use the unified Docker image with Chromium included.
