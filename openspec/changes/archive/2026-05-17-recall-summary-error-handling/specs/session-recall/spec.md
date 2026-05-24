## MODIFIED Requirements

### Requirement: SessionRecallRetriever

The system SHALL provide a `SessionRecallRetriever` that implements the existing context retriever interface consumed by `ContextAwareModelAdapter`. At turn start, the retriever SHALL query `fts_session_recall` using the user's current input as the MATCH string, return up to `N` top-ranked results (default 3), apply a BM25 rank floor (default `0.2`), and exclude results whose `session_key` equals the current session. Truncation to fit the available RAG section budget SHALL use the existing `SectionBudgets.RAG` value. If summary loading fails for a retained match, the retriever SHALL return a non-nil error instead of returning a match with an empty or incomplete summary.

#### Scenario: Retriever returns matches above floor
- **WHEN** the user's input matches two prior-session summaries with BM25 rank 0.4 and 0.5
- **AND** the floor is 0.2
- **THEN** both summaries SHALL be returned for the turn

#### Scenario: Results below floor are filtered
- **WHEN** a candidate match has BM25 rank 0.1 and the floor is 0.2
- **THEN** that candidate SHALL NOT be returned

#### Scenario: Current session excluded
- **WHEN** the current session is `sess-42` and a match with `session_key = "sess-42"` appears in the result set
- **THEN** that match SHALL be filtered out

#### Scenario: Feature disabled returns nothing
- **WHEN** `context.recall.enabled` is `false`
- **THEN** the retriever SHALL return an empty result set without querying FTS

#### Scenario: Summary load failure is reported
- **WHEN** a search hit survives filtering but loading its summary fails
- **THEN** the retriever SHALL return a non-nil error that identifies the session key
- **AND** it SHALL NOT return a recall match with an empty summary for that hit
