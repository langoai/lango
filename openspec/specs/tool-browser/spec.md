## Purpose

Capability spec for tool-browser. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Browser automation via go-rod
The system SHALL provide browser automation tools powered by go-rod for web page interaction, with local browser launch support.

#### Scenario: Browser navigation
- **WHEN** `browser_navigate` is called with a URL
- **THEN** the system SHALL navigate to the URL, wait for page load, and return a structured page snapshot including title, URL, snippet, headings, links, action candidates, page type, result count, and empty-state signal

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

### Requirement: Browser-native web search
The browser toolset SHALL provide a `browser_search` tool that accepts a search query and returns structured search results without requiring the agent to manually drive a search engine page step by step.

#### Scenario: Search query returns structured results
- **WHEN** `browser_search` is called with a query and optional result limit
- **THEN** the system SHALL navigate to a browser-accessible search results page
- **AND** it SHALL return a structured list of results containing title, URL, and snippet
- **AND** it SHALL include `resultCount`, `empty`, and page type fields

#### Scenario: Search results fallback to visible links
- **WHEN** the page-specific search result selectors do not match any result cards
- **THEN** the system SHALL fall back to extracting visible links with text and URLs

#### Scenario: Search churn diagnostics
- **WHEN** a single request performs 3 or more `browser_search` calls
- **THEN** the system SHALL emit a warning log including session, request ID, agent, search count, queries, and current URL
- **AND** `RecordSearch()` SHALL return `(count int, queries []string, shouldWarn bool, limitReached bool)`
- **AND** when `count > MaxSearchesPerRequest`, `limitReached` SHALL be true

#### Scenario: RecordSearch preserves currentURL on empty input
- **WHEN** `RecordSearch` is called with an empty `currentURL`
- **THEN** the previously stored `currentURL` SHALL be preserved

### Requirement: Browser search hard limit per request
The browser `Search()` function SHALL enforce a maximum of `MaxSearchesPerRequest` (2) calls per agent request. When the limit is exceeded, the function SHALL return `ErrSearchLimitReached` (a sentinel error) instead of executing the search. Returning an error instead of a structured response ensures the model interprets the limit as a tool failure, providing a stronger convergence signal.

#### Scenario: Third search attempt is blocked
- **WHEN** the agent calls `browser_search` for the 3rd time in the same request
- **THEN** `Search()` SHALL return `(nil, ErrSearchLimitReached)` — an error, not a SearchResponse

#### Scenario: First two searches execute normally
- **WHEN** the agent calls `browser_search` for the 1st or 2nd time in the same request
- **THEN** `Search()` SHALL execute the search normally and return results

### Requirement: SearchResponse convergence fields
`SearchResponse` SHALL include `LimitReached bool`, `NextStep string`, and `Warning string` fields. Normal search results SHALL populate `NextStep` with guidance based on result state.

#### Scenario: Search returns results
- **WHEN** `browser_search` returns results with `resultCount > 0`
- **THEN** `NextStep` SHALL contain guidance to present results or navigate to a result URL, explicitly stating not to search again

#### Scenario: Search returns no results
- **WHEN** `browser_search` returns results with `resultCount == 0`
- **THEN** `NextStep` SHALL contain guidance to reformulate the query once or inform the user

### Requirement: Page observation
The browser toolset SHALL provide a `browser_observe` tool that returns structured actionable elements from the current page.

#### Scenario: Observe actionable elements
- **WHEN** `browser_observe` is called on the current page
- **THEN** the system SHALL return clickable and input-capable elements with stable selectors and descriptive metadata

### Requirement: Structured page extraction
The browser toolset SHALL provide a `browser_extract` tool for structured extraction from the current page.

#### Scenario: Summary extraction
- **WHEN** `browser_extract` is called with mode `summary`
- **THEN** the system SHALL return page title, URL, snippet, headings, links, action candidates, page type, result count, and empty-state signal

#### Scenario: Search result extraction
- **WHEN** `browser_extract` is called with mode `search_results`
- **THEN** the system SHALL return structured search result items from the current page
- **AND** it SHALL include `resultCount`, `empty`, and page type fields

#### Scenario: Article extraction
- **WHEN** `browser_extract` is called with mode `article`
- **THEN** the system SHALL return the main textual content and headings from the current page
- **AND** it SHALL include current URL, page type, and empty-state signal

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

#### Scenario: CDP target error on browser_navigate triggers retry
- **WHEN** `browser_navigate` returns an error containing "Inspected target navigated or closed"
- **THEN** the middleware SHALL close the browser session
- **AND** retry the navigation once with a fresh session
- **AND** if the retry also fails, return the error as-is

#### Scenario: CDP target error on browser_action does NOT trigger retry
- **WHEN** `browser_action` returns an error containing "Inspected target navigated or closed"
- **THEN** the middleware SHALL NOT retry the action
- **AND** the error SHALL be returned as-is



### Requirement: Private network URL blocking for P2P
The `browser_navigate` handler MUST validate URLs against a private network blocklist when the context carries a P2P origin marker. Blocked addresses: `localhost`, `127.0.0.0/8`, `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`, `169.254.0.0/16`, `[::1]`, and `file://` scheme. `ValidateURLForP2P` MUST resolve non-IP hostnames via `net.LookupIP` and check all resolved IPs against private ranges. After navigation completes, the handler MUST always retrieve the final page URL and re-validate it via `ValidateURLForP2P`, regardless of whether the final URL string matches the original request URL. This prevents both redirect-based SSRF and DNS rebinding attacks where the same hostname resolves to a different IP at navigation time.

#### Scenario: Internal URL blocked in P2P context
- **WHEN** a P2P peer navigates to `http://127.0.0.1:8080/admin`
- **THEN** the handler returns `ErrBlockedURL` without creating a browser session

#### Scenario: Private network IP blocked
- **WHEN** a P2P peer navigates to `http://10.0.0.1/internal`
- **THEN** the handler returns `ErrBlockedURL`

#### Scenario: File scheme blocked
- **WHEN** a P2P peer navigates to `file:///etc/passwd`
- **THEN** the handler returns `ErrBlockedURL`

#### Scenario: External URL allowed in P2P context
- **WHEN** a P2P peer navigates to `https://example.com`
- **THEN** navigation proceeds normally

#### Scenario: URL validation skipped for local context
- **WHEN** a local (non-P2P) user navigates to `http://localhost:3000`
- **THEN** navigation proceeds normally (no restriction)

#### Scenario: Hostname resolving to private IP blocked
- **WHEN** a P2P peer navigates to `http://metadata.internal` which resolves to `169.254.169.254`
- **THEN** `ValidateURLForP2P` returns `ErrBlockedURL`

#### Scenario: Redirect to internal address blocked
- **WHEN** a P2P peer navigates to `https://external.com` which redirects to `http://127.0.0.1:8080`
- **THEN** the handler navigates to `about:blank`
- **AND** returns an error wrapping `ErrBlockedURL`

### Requirement: Eval action blocking for P2P
The `browser_action` handler MUST reject `eval` actions when the context carries a P2P origin marker, returning `ErrEvalBlockedP2P` before creating a browser session.

#### Scenario: Eval blocked for P2P peer
- **WHEN** a P2P peer sends `browser_action` with `action: "eval"`
- **THEN** the handler returns `ErrEvalBlockedP2P`

#### Scenario: Eval allowed for local user
- **WHEN** a local user sends `browser_action` with `action: "eval"`
- **THEN** the JavaScript is executed normally

### Requirement: Browser sentinel errors
The browser package MUST define `ErrBlockedURL` and `ErrEvalBlockedP2P` sentinel errors.
