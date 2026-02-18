## ADDED Requirements

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

### Requirement: Panic recovery for rod/CDP calls
The browser tool SHALL recover from panics in go-rod/rod library calls and convert them into structured errors instead of crashing the process.

#### Scenario: Rod panic during navigation
- **WHEN** a rod API call panics during `Navigate`
- **THEN** the system SHALL recover the panic and return an error wrapping `ErrBrowserPanic`
- **AND** the process SHALL NOT crash

#### Scenario: Rod panic during screenshot
- **WHEN** a rod API call panics during `Screenshot`
- **THEN** the system SHALL recover the panic and return an error wrapping `ErrBrowserPanic`

#### Scenario: Rod panic during element interaction
- **WHEN** a rod API call panics during `Click`, `Type`, `GetText`, `GetElementInfo`, or `Eval`
- **THEN** the system SHALL recover the panic and return an error wrapping `ErrBrowserPanic`

#### Scenario: Rod panic during session creation
- **WHEN** a rod API call panics during `NewSession`
- **THEN** the system SHALL recover the panic and return an error wrapping `ErrBrowserPanic`

#### Scenario: Rod panic during close
- **WHEN** a rod API call panics during `Close`
- **THEN** the system SHALL recover the panic silently
- **AND** cleanup SHALL continue for remaining sessions

#### Scenario: Normal errors pass through unchanged
- **WHEN** a rod API call returns a normal error (no panic)
- **THEN** the error SHALL be returned as-is without `ErrBrowserPanic` wrapping

### Requirement: Auto-reconnect on browser panic
The SessionManager SHALL detect `ErrBrowserPanic` during session creation and attempt to reconnect by closing the browser and retrying once.

#### Scenario: Reconnect on EnsureSession panic
- **WHEN** `EnsureSession` receives `ErrBrowserPanic` from `NewSession`
- **THEN** the SessionManager SHALL close the browser tool
- **AND** the SessionManager SHALL retry `NewSession` exactly once
- **AND** if the retry succeeds, the new session ID SHALL be returned

#### Scenario: Reconnect retry fails
- **WHEN** `EnsureSession` receives `ErrBrowserPanic` and the retry also fails
- **THEN** the error SHALL be returned to the caller
- **AND** no further retries SHALL be attempted

### Requirement: Browser tool handler panic wrapper
The application layer SHALL wrap all browser tool handlers with panic recovery and retry logic.

#### Scenario: Handler-level panic recovery
- **WHEN** a browser tool handler panics during execution
- **THEN** the wrapper SHALL recover the panic and return an error wrapping `ErrBrowserPanic`

#### Scenario: Handler-level retry on ErrBrowserPanic
- **WHEN** a browser tool handler returns `ErrBrowserPanic`
- **THEN** the wrapper SHALL close the session manager and retry the handler once
- **AND** if the retry succeeds, the result SHALL be returned normally
