## Context

The Librarian sub-agent currently only reacts to explicit user tool calls (search_knowledge, save_knowledge, etc.). The Observational Memory pipeline already compresses conversation history into observations, but no component analyzes those observations for extractable knowledge or knowledge gaps. This design adds a proactive analysis layer that bridges observations to knowledge.

## Goals / Non-Goals

**Goals:**
- Automatically extract high-confidence knowledge from conversation observations
- Detect knowledge gaps and create inquiry records for natural follow-up questions
- Inject pending inquiries into the system prompt so the agent weaves questions conversationally
- Provide agent tools for inquiry lifecycle management (list, dismiss)
- Follow existing buffer/callback patterns for async processing

**Non-Goals:**
- Real-time conversation interruption (inquiries are injected at next turn, not mid-response)
- Multi-session inquiry correlation (each session manages its own inquiries)
- Complex NLP entity extraction (relies on LLM for all analysis)
- UI/TUI for inquiry management (agent tools only)

## Decisions

### 1. Separate `internal/librarian/` package
**Decision**: Create a new package rather than extending `internal/learning/`.
**Rationale**: The learning package handles error patterns and conversation analysis. The librarian's concern (proactive knowledge extraction + inquiry lifecycle) is orthogonal. A separate package avoids bloating learning and keeps clear boundaries.
**Alternative**: Extending learning package — rejected due to divergent responsibilities.

### 2. Ent schema for Inquiry persistence
**Decision**: Use Ent schema with `(session_key, status)` composite index.
**Rationale**: Inquiries must survive process restarts and need efficient queries by session + status. Ent is already the project's ORM. The composite index optimizes the primary query pattern (pending inquiries per session).
**Alternative**: In-memory map — rejected because inquiries would be lost on restart.

### 3. Buffer pattern for async processing
**Decision**: Mirror `learning.AnalysisBuffer` pattern with Start/Trigger/Stop lifecycle.
**Rationale**: Proven pattern in the codebase (memory.Buffer, graph.GraphBuffer, learning.AnalysisBuffer). Trigger fires on OnTurnComplete, processing happens in a background goroutine.
**Alternative**: Synchronous processing in OnTurnComplete — rejected because LLM calls would block the turn response.

### 4. Two-phase processing: InquiryProcessor then ObservationAnalyzer
**Decision**: Each trigger first checks for answers to pending inquiries, then analyzes observations.
**Rationale**: Answer detection should run before creating new inquiries to avoid re-asking answered questions. The two phases are complementary and share the same session context.

### 5. Cooldown and limit controls
**Decision**: Configurable cooldown turns (default: 3) and max pending inquiries (default: 2) per session.
**Rationale**: Prevents the agent from overwhelming users with questions. The cooldown ensures minimum spacing between inquiry injections. The max pending limit caps the total outstanding questions.

### 6. InquiryProvider interface for context injection
**Decision**: Add `InquiryProvider` interface to knowledge package, wire into ContextRetriever.
**Rationale**: Follows the existing provider pattern (ToolRegistryProvider, RuntimeContextProvider, SkillProvider). The interface avoids direct librarian→knowledge import cycles. ContextRetriever already assembles the system prompt, so this is the natural injection point.

## Risks / Trade-offs

- **LLM cost**: Each trigger makes 1-2 LLM calls (observation analysis + answer detection) → Mitigated by observation threshold (minimum 2 observations before analysis) and cooldown turns.
- **False positive knowledge extraction**: LLM may extract incorrect knowledge → Mitigated by auto-save only at "high" confidence level by default; medium-confidence items become inquiries for user confirmation.
- **Inquiry fatigue**: Too many questions may annoy users → Mitigated by maxPendingInquiries limit (2) and cooldownTurns (3).
- **Duplicate knowledge**: Extracted knowledge may already exist → Should be mitigated by checking existing knowledge before saving (future enhancement via SearchKnowledge dedup).
