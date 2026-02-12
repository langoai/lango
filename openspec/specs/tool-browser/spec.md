## ADDED Requirements

### Requirement: Browser automation via go-rod
The system SHALL provide browser automation tools powered by go-rod for web page interaction.

#### Scenario: Browser navigation
- **WHEN** `browser_navigate` is called with a URL
- **THEN** the system SHALL navigate to the URL, wait for page load, and return title, URL, and text snippet

#### Scenario: Implicit session management
- **WHEN** any browser tool is called without a prior session
- **THEN** the system SHALL auto-create a browser session and reuse it for subsequent calls
- **AND** the LLM SHALL NOT need to manage session IDs

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

### Requirement: Lifecycle cleanup
The system SHALL clean up browser resources on shutdown.

#### Scenario: Graceful shutdown
- **WHEN** the application stops
- **THEN** all browser sessions SHALL be closed and the Chromium process terminated
