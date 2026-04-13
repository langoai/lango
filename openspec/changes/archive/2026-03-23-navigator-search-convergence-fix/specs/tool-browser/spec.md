## MODIFIED Requirements

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
