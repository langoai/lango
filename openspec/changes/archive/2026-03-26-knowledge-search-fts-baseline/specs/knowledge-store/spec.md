## MODIFIED Requirements

### Requirement: Knowledge Search
The system SHALL support full-text search across knowledge entries. When an FTS5 index is available, `SearchKnowledge` SHALL use FTS5 MATCH with BM25 ranking as the primary search path. When FTS5 is unavailable, `SearchKnowledge` SHALL fall back to per-keyword `ContentContains`/`KeyContains` LIKE predicates combined with OR logic, ordered by `RelevanceScore` descending. The system SHALL NOT use a single concatenated query string as a LIKE pattern in either path.

#### Scenario: FTS5 search path
- **WHEN** `SearchKnowledge` is called with a query string and FTS5 index is available
- **THEN** the system SHALL return entries matching the FTS5 query, ordered by BM25 relevance
- **AND** results SHALL be limited to the specified limit (default 10)

#### Scenario: LIKE fallback search path
- **WHEN** `SearchKnowledge` is called with a query string and FTS5 index is NOT available
- **THEN** the system SHALL return entries where the content or key contains any of the individual keywords
- **AND** results SHALL be ordered by relevance score descending
- **AND** results SHALL be limited to the specified limit (default 10)

#### Scenario: Multi-keyword FTS5 search
- **WHEN** `SearchKnowledge` is called with query "deploy server config" and FTS5 is available
- **THEN** the FTS5 MATCH query SHALL search for all keywords and rank by BM25

#### Scenario: Multi-keyword LIKE fallback
- **WHEN** `SearchKnowledge` is called with query "deploy server config" and FTS5 is NOT available
- **THEN** the SQL query uses per-keyword LIKE predicates: `(content LIKE '%deploy%' OR key LIKE '%deploy%') OR (content LIKE '%server%' OR key LIKE '%server%') OR (content LIKE '%config%' OR key LIKE '%config%')`

#### Scenario: Search with category filter
- **WHEN** `SearchKnowledge` is called with a query and a category
- **THEN** the system SHALL return only entries matching both the query and the category (in both FTS5 and LIKE paths)

#### Scenario: FTS5 error graceful degradation
- **WHEN** `SearchKnowledge` via FTS5 encounters an error
- **THEN** the system SHALL log a warning and fall back to the LIKE path for that query
