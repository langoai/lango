## ADDED Requirements

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

#### Scenario: Thread-safe browser initialization
- **WHEN** multiple browser tool calls are made concurrently
- **THEN** only one initialization attempt SHALL execute at a time
- **AND** subsequent concurrent calls SHALL wait for and share the result

#### Scenario: Retry on connection failure
- **WHEN** browser initialization fails (e.g., remote Chrome not ready)
- **THEN** the failed state SHALL NOT be cached permanently
- **AND** the next browser tool call SHALL retry initialization

#### Scenario: No partial initialization
- **WHEN** `Connect()` fails during browser initialization
- **THEN** the browser field SHALL remain nil
- **AND** subsequent calls SHALL NOT observe a non-nil but disconnected browser

#### Scenario: Re-initialization after close
- **WHEN** `Close()` is called and browser resources are cleaned up
- **THEN** the initialization guard SHALL be reset
- **AND** the next browser tool call SHALL re-initialize from scratch

### Requirement: Page interaction via browser_action
The system SHALL multiplex page interactions through a single `browser_action` tool.

#### Scenario: Click action
- **WHEN** `browser_action` is called with `action: "click"` and a CSS `selector`
- **THEN** the system SHALL click the matching element

#### Scenario: Type action
- **WHEN** `browser_action` is called with `action: "type"`, a CSS `selector`, and `text`
- **THEN** the system SHALL input the text into the matching element

#### Scenario: Eval action
- **WHEN** `browser_action` is called with `action: "eval"` and JavaScript in `text`
- **THEN** the system SHALL evaluate the script and return the result

#### Scenario: Get text action
- **WHEN** `browser_action` is called with `action: "get_text"` and a CSS `selector`
- **THEN** the system SHALL return the text content of the matching element

#### Scenario: Get element info action
- **WHEN** `browser_action` is called with `action: "get_element_info"` and a CSS `selector`
- **THEN** the system SHALL return tag name, id, className, innerText, href, and value

#### Scenario: Wait action
- **WHEN** `browser_action` is called with `action: "wait"`, a CSS `selector`, and optional `timeout`
- **THEN** the system SHALL wait for the element to appear (default: 10s)

### Requirement: Screenshot capture
The system SHALL capture screenshots of the current browser page.

#### Scenario: Viewport screenshot
- **WHEN** `browser_screenshot` is called with `fullPage: false` (default)
- **THEN** the system SHALL return a base64-encoded PNG of the visible viewport

#### Scenario: Full page screenshot
- **WHEN** `browser_screenshot` is called with `fullPage: true`
- **THEN** the system SHALL return a base64-encoded PNG of the full scrollable page

### Requirement: Opt-in configuration
Browser tools SHALL be disabled by default and require explicit opt-in.

#### Scenario: Default disabled
- **GIVEN** no `tools.browser.enabled` config is set
- **THEN** browser tools SHALL NOT be registered and no Chromium process SHALL be started

#### Scenario: Enabled
- **GIVEN** `tools.browser.enabled: true` in configuration
- **THEN** browser tools SHALL be registered and available to the agent

### Requirement: Browser config fields exposed in TUI
The Onboard TUI Tools form SHALL expose the `enabled` and `sessionTimeout` fields for browser tool configuration.

#### Scenario: Browser enabled toggle in TUI
- **WHEN** user navigates to Tools configuration in the onboard wizard
- **THEN** a "Browser Enabled" boolean toggle SHALL be displayed before the "Browser Headless" toggle

#### Scenario: Browser session timeout in TUI
- **WHEN** user navigates to Tools configuration in the onboard wizard
- **THEN** a "Browser Session Timeout" duration text field SHALL be displayed after the "Browser Headless" toggle
- **AND** the field SHALL accept Go duration strings (e.g., "5m", "10m")

### Requirement: Browser binary auto-detection
The system SHALL auto-detect system-installed browser binaries using `launcher.LookPath()` when no explicit `BrowserBin` config is set.

#### Scenario: System browser found
- **WHEN** `BrowserBin` config is empty
- **AND** `launcher.LookPath()` finds a system browser
- **THEN** the system SHALL use the detected binary path

#### Scenario: Explicit browser path
- **WHEN** `BrowserBin` config is set to a non-empty path
- **THEN** the system SHALL use the configured path regardless of LookPath result

#### Scenario: No browser found
- **WHEN** `BrowserBin` config is empty
- **AND** `launcher.LookPath()` does not find a system browser
- **THEN** the system SHALL fall back to go-rod's default browser download behavior

### Requirement: BrowserBin config field
The `BrowserToolConfig` SHALL include a `BrowserBin` string field for specifying an explicit browser binary path.

#### Scenario: Config field recognition
- **WHEN** `tools.browser.browserBin: "/usr/bin/chromium"` is set in configuration
- **THEN** the system SHALL pass the path to `launcher.Bin()`

### Requirement: Lifecycle cleanup
The system SHALL clean up browser resources on shutdown.

#### Scenario: Graceful shutdown
- **WHEN** the application stops
- **THEN** all browser sessions SHALL be closed and the Chromium process terminated
