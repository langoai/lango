# tool-websearch Specification

## Purpose
HTTP-based web search tool providing lightweight search without browser sessions. Uses DuckDuckGo's HTML endpoint via `net/http` to return structured search results with title, URL, and snippet.

## Requirements

### Requirement: HTTP-only DuckDuckGo search
The system SHALL provide a `web_search` tool that searches via DuckDuckGo's HTML endpoint (`https://html.duckduckgo.com/html/?q=`) using `net/http` without requiring a browser session. The HTTP client SHALL use a 15-second timeout and set a `Mozilla/5.0 (compatible; Lango/1.0)` User-Agent header.

#### Scenario: Basic search
- **WHEN** `web_search` is invoked with query "golang concurrency" and limit 5
- **THEN** it SHALL return up to 5 `SearchResult` entries, each with `title`, `url`, and `snippet` fields
- **AND** the response SHALL include the original `query` and a `count` of returned results

#### Scenario: Empty query
- **WHEN** `web_search` is invoked with an empty query string
- **THEN** it SHALL return an error "query is required"

#### Scenario: No results found
- **WHEN** the DuckDuckGo endpoint returns HTML with no result containers
- **THEN** the system SHALL return an empty result slice with no error

#### Scenario: Non-200 HTTP response
- **WHEN** the DuckDuckGo endpoint returns a non-200 HTTP status code
- **THEN** the system SHALL return an error containing the HTTP status code

### Requirement: DuckDuckGo redirect URL resolution
The system SHALL resolve DuckDuckGo redirect URLs (of the form `//duckduckgo.com/l/?uddg=<encoded_url>&...`) to their actual target URL by extracting the `uddg` query parameter.

#### Scenario: DDG redirect URL
- **WHEN** a search result contains a DuckDuckGo redirect URL with `uddg` parameter
- **THEN** the system SHALL resolve it to the decoded target URL

#### Scenario: Protocol-relative URL
- **WHEN** a search result URL starts with `//`
- **THEN** the system SHALL prepend `https:` to form a full URL

#### Scenario: Non-HTTP scheme filtered
- **WHEN** a resolved URL has a scheme other than `http` or `https` (e.g., `ftp://`)
- **THEN** the result SHALL be excluded (resolveURL returns empty string)

### Requirement: HTML result parsing
The system SHALL parse DuckDuckGo's HTML response by locating `<div>` elements with class containing both `result` and `web-result`. Within each container, the title and URL SHALL be extracted from `<a class="result__a">` and the snippet from elements with class `result__snippet`.

#### Scenario: Standard result extraction
- **WHEN** the HTML contains result container divs with title links and snippet elements
- **THEN** the system SHALL extract the title text, href URL, and snippet text for each result

#### Scenario: Result missing title or URL
- **WHEN** a result container has no `result__a` link or the link has an empty title/URL
- **THEN** that result SHALL be skipped without error

### Requirement: P2P URL safety
The system SHALL filter out results pointing to private/internal network addresses when `ctxkeys.IsP2PRequest(ctx)` is true. Filtering SHALL use `browser.ValidateURLForP2P` which blocks `file://` scheme, `localhost`, loopback addresses, and private network ranges (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 169.254.0.0/16, 127.0.0.0/8).

#### Scenario: P2P filters private IPs
- **WHEN** search results include URLs pointing to 192.168.1.1 or 10.x.x.x in P2P context
- **THEN** those results SHALL be excluded from the returned list

#### Scenario: P2P keeps public URLs
- **WHEN** search results include URLs pointing to public addresses in P2P context
- **THEN** those results SHALL be included in the returned list

#### Scenario: Non-P2P keeps all URLs
- **WHEN** a local (non-P2P) request returns results containing private IP URLs
- **THEN** all results SHALL be returned without filtering

### Requirement: Tool metadata
The `web_search` tool SHALL be registered with `SafetyLevel` Safe, Capability category `web`, aliases `["search_web", "internet_search"]`, search hints `["search", "find", "lookup", "web"]`, `ReadOnly` true, `ConcurrencySafe` true, and Activity `ActivityQuery`.

#### Scenario: Tool metadata verification
- **WHEN** `BuildTools()` is called
- **THEN** it returns exactly one tool named `web_search` with the specified capability metadata

#### Scenario: Tool parameter schema
- **WHEN** `BuildTools()` is called
- **THEN** the tool SHALL have `query` (string, required) and `limit` (integer, optional) parameters

### Requirement: Limit clamping
The system SHALL clamp the `limit` parameter to a range of 1-20. If `limit` is 0 or negative, it SHALL default to 5. If `limit` exceeds 20, it SHALL be clamped to 20.

#### Scenario: Default limit
- **WHEN** `limit` is 0 or not provided
- **THEN** the system SHALL use 5 as the default limit

#### Scenario: Negative limit
- **WHEN** `limit` is negative (e.g., -1)
- **THEN** the system SHALL use the default limit of 5

#### Scenario: Limit within range
- **WHEN** `limit` is between 1 and 20
- **THEN** the system SHALL use the provided limit value

#### Scenario: Limit above maximum
- **WHEN** `limit` exceeds 20 (e.g., 25)
- **THEN** the system SHALL clamp it to 20
