## ADDED Requirements

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
