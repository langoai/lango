## ADDED Requirements

### Requirement: Fetch web page content via HTTP
The system SHALL provide a `web_fetch` agent tool that fetches a web page via HTTP GET and returns extracted content. The tool SHALL use a 30-second timeout and follow up to 5 redirects. The tool SHALL set a `Lango/1.0` User-Agent header.

#### Scenario: Successful text extraction
- **WHEN** the agent invokes `web_fetch` with a valid URL and mode `text`
- **THEN** the system returns a `FetchResult` containing the page title, extracted clean text content (nav/footer/script stripped), content length, and truncation status

#### Scenario: Successful HTML extraction
- **WHEN** the agent invokes `web_fetch` with a valid URL and mode `html`
- **THEN** the system returns the raw HTML body truncated to max_length

#### Scenario: Successful markdown extraction
- **WHEN** the agent invokes `web_fetch` with a valid URL and mode `markdown`
- **THEN** the system returns simplified markdown with headings as `#`, links as `[text](url)`, lists as `- item`, and bold/italic preserved

#### Scenario: Non-200 HTTP response
- **WHEN** the target URL returns a non-2xx HTTP status code
- **THEN** the system returns an error containing the HTTP status code

### Requirement: Content truncation with configurable max_length
The system SHALL truncate extracted content to the `max_length` parameter (default 5000 characters). When content is truncated, the `Truncated` field in the result SHALL be `true`.

#### Scenario: Content exceeds max_length
- **WHEN** extracted content exceeds the configured max_length
- **THEN** the content is truncated to exactly max_length characters and `Truncated` is `true`

#### Scenario: Content within max_length
- **WHEN** extracted content is within the configured max_length
- **THEN** the full content is returned and `Truncated` is `false`

### Requirement: Readability extraction
The system SHALL extract main content from HTML by locating `<article>`, `<main>`, or the largest `<div>` by text length. The system SHALL strip `<nav>`, `<footer>`, `<aside>`, `<script>`, `<style>`, `<noscript>`, `<iframe>`, and `<svg>` elements.

#### Scenario: Article element present
- **WHEN** the HTML contains an `<article>` element
- **THEN** the system extracts content from within that element

#### Scenario: Fallback to main element
- **WHEN** the HTML has no `<article>` but has a `<main>` element
- **THEN** the system extracts content from within `<main>`

#### Scenario: Fallback to largest div
- **WHEN** the HTML has neither `<article>` nor `<main>`
- **THEN** the system extracts content from the `<div>` with the most text

### Requirement: P2P URL safety validation
In P2P peer request contexts, the system SHALL validate URLs before fetching and after redirect resolution. The system SHALL block `file://` scheme, `localhost`, loopback addresses, and private network ranges (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 169.254.0.0/16, 127.0.0.0/8).

#### Scenario: P2P request to private IP
- **WHEN** a P2P peer requests fetch of `http://192.168.1.1/admin`
- **THEN** the system returns an error indicating the URL targets a blocked address

#### Scenario: Non-P2P request to private IP
- **WHEN** a local (non-P2P) request fetches a private IP
- **THEN** the system allows the request without URL validation

### Requirement: Tool registration with capability metadata
The `BuildTools()` function SHALL return a single `web_fetch` tool with `SafetyLevelModerate`, category `web`, aliases `["fetch_url", "get_page"]`, activity `ActivityRead`, and search hints `["fetch", "download", "page", "url", "content"]`.

#### Scenario: Tool metadata verification
- **WHEN** `BuildTools()` is called
- **THEN** it returns exactly one tool named `web_fetch` with the specified capability metadata
