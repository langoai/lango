## ADDED Requirements

### Requirement: Browser search hard limit per request

The browser `Search()` function SHALL enforce a maximum of `MaxSearchesPerRequest` (2) calls per agent request. When the limit is exceeded, the function SHALL return a structured `SearchResponse` with `LimitReached=true` and a `NextStep` advisory instead of executing the search.

#### Scenario: Third search attempt is blocked

- **WHEN** the agent calls `browser_search` for the 3rd time in the same request
- **THEN** `Search()` SHALL return a `SearchResponse` with `PageType="search_limit"`, `LimitReached=true`, `URL` set to the current page URL, and `NextStep` containing guidance to use `browser_extract` or `browser_navigate` instead

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

## MODIFIED Requirements

### Requirement: Browser search churn diagnostics

The browser tool SHALL track per-request search count via `RequestState`. `RecordSearch()` SHALL return `(count int, queries []string, shouldWarn bool, limitReached bool)`. When `count > MaxSearchesPerRequest`, `limitReached` SHALL be true. The `shouldWarn` flag SHALL trigger at `count >= 3` as before.

#### Scenario: RecordSearch returns limitReached at 3rd call

- **WHEN** `RecordSearch` is called for the 3rd time (`MaxSearchesPerRequest=2`)
- **THEN** `limitReached` SHALL be true and `shouldWarn` SHALL be true

#### Scenario: RecordSearch preserves currentURL on empty input

- **WHEN** `RecordSearch` is called with an empty `currentURL`
- **THEN** the previously stored `currentURL` SHALL be preserved
