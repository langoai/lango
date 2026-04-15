# session-recall Specification

## Purpose
TBD - created by archiving change ux-continuity. Update Purpose after archive.
## Requirements
### Requirement: Session end triggers
The system SHALL recognize two session end modes and process each through a unified recall pipeline.

- *Hard end*: triggered by `session.Store.End(key)`, TUI quit, or CLI exit. SHALL run a best-effort synchronous summarization and FTS indexing step, bounded by a drain timeout (default 3s).
- *Soft end*: triggered by channel idle timeout or adaptive idle expiry. SHALL set metadata key `lango.session_end_pending=true` on the session and return immediately, without synchronous work.

On next session open (for any principal), the system SHALL sweep for `lango.session_end_pending=true` sessions and run the same summarization + indexing pipeline asynchronously.

#### Scenario: Hard end runs synchronously with bounded drain
- **WHEN** the TUI exits via Ctrl+D
- **THEN** the session's final message list SHALL be summarized and indexed into the FTS recall table
- **AND** the total wait SHALL NOT exceed 3 seconds
- **AND** on timeout the session SHALL still be marked `lango.session_end_pending=true` so the next open resumes the work

#### Scenario: Soft end only marks metadata
- **WHEN** the adaptive idle timer fires for a channel session
- **THEN** the session SHALL be marked with `lango.session_end_pending=true`
- **AND** no summarization or indexing SHALL run synchronously

#### Scenario: Next-open sweep processes pending sessions
- **WHEN** any TUI or channel session is opened
- **AND** one or more sessions carry `lango.session_end_pending=true`
- **THEN** the system SHALL process them asynchronously and clear the flag on success

### Requirement: FTS5 session recall index
The system SHALL provide a dedicated FTS5 virtual table `fts_session_recall` (via the existing `internal/search.FTS5Index`) with columns `session_key`, `summary`, `role_mix`, and `ended_at` (as an UNINDEXED column-like field if needed or stored separately). Each indexed row represents exactly one ended session. The existing knowledge FTS5 index SHALL remain independent and untouched.

#### Scenario: Table created lazily on first use
- **WHEN** the first session-end indexing occurs
- **THEN** `EnsureTable()` SHALL create `fts_session_recall` if it does not yet exist
- **AND** the table SHALL be separate from the knowledge FTS5 table

#### Scenario: One row per ended session
- **WHEN** session `sess-1` is ended and indexed
- **THEN** `fts_session_recall` SHALL contain exactly one row with `rowid = "sess-1"`

#### Scenario: Re-indexing replaces the row
- **WHEN** `sess-1` is ended, later reopened with more messages, and ended again
- **THEN** the existing row for `sess-1` SHALL be replaced (delete-then-insert) rather than duplicated

### Requirement: SessionRecallRetriever
The system SHALL provide a `SessionRecallRetriever` that implements the existing context retriever interface consumed by `ContextAwareModelAdapter`. At turn start, the retriever SHALL query `fts_session_recall` using the user's current input as the MATCH string, return up to `N` top-ranked results (default 3), apply a BM25 rank floor (default `0.2`), and exclude results whose `session_key` equals the current session. Truncation to fit the available RAG section budget SHALL use the existing `SectionBudgets.RAG` value.

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

### Requirement: Recall respects section budget
The injected recall content SHALL truncate to fit within the available RAG section budget from `ContextBudgetManager.SectionBudgets()`. Lower-ranked items SHALL be dropped first if the full set would exceed the budget.

#### Scenario: Budget fits all matches
- **WHEN** the RAG budget has headroom for all 3 returned matches
- **THEN** all 3 SHALL be included

#### Scenario: Budget fits only top matches
- **WHEN** the RAG budget has headroom only for the top 2 of 3 matches
- **THEN** only the top 2 SHALL be included
- **AND** the third SHALL be dropped with a debug-level log

### Requirement: Config surface for recall
The system SHALL provide additive fields under `context.recall`: `enabled bool` (default `true`), `topN int` (default `3`, valid range `[1, 10]`), and `minRank float64` (default `0.2`, valid range `[0.0, 1.0]`). Invalid values SHALL be clamped to valid ranges with a warning log.

#### Scenario: Defaults when unset
- **WHEN** no `context.recall.*` config is set
- **THEN** recall SHALL be enabled with topN=3 and minRank=0.2

#### Scenario: Disable opts out entirely
- **WHEN** `context.recall.enabled` is `false`
- **THEN** the retriever SHALL be a no-op and no FTS queries SHALL be issued

### Requirement: Plain chat session end for recall indexing
The plain chat path (`runChat`) SHALL defer `Store.End(sessionKey)` so that session recall indexing runs on TUI exit, matching the cockpit path behavior.

#### Scenario: Exit plain chat
- **WHEN** user exits `lango chat` via Ctrl+C or Ctrl+D
- **THEN** `Store.End(sessionKey)` SHALL be called
- **AND** the session SHALL be indexed for future recall

### Requirement: Session key rebinding on /clear
After `/clear` regenerates the session key, all session-scoped bindings SHALL use the new key. This includes event subscription filters, the `defer Store.End()` call, and token accumulation.

#### Scenario: /clear then continue chatting
- **WHEN** user runs `/clear` and then sends a new message
- **THEN** continuity events (compaction, learning suggestions) SHALL filter by the NEW session key
- **AND** `Store.End()` on exit SHALL be called with the NEW session key
- **AND** the old session key's End processing SHALL NOT run (or be a safe no-op)

### Requirement: RAG budget split between recall and semantic results
When both session recall matches and RAG/GraphRAG results exist, the RAG section budget SHALL be split: recall receives 1/3, semantic RAG receives 2/3. When only one source exists, it receives the full budget.

#### Scenario: Both recall and RAG present
- **WHEN** a turn has both recall matches and RAG results
- **THEN** the combined formatted output SHALL NOT exceed `budgets.RAG` tokens
- **AND** recall section SHALL be truncated to approximately 1/3 of the budget

