## Why

The current Librarian agent operates only reactively — it responds to explicit user requests for knowledge search, storage, and skill management. By leveraging the existing Observational Memory pipeline, the Librarian can proactively extract knowledge from conversations, detect knowledge gaps, and ask clarifying questions to strengthen Lango's knowledge base over time — making the agent smarter with every interaction.

## What Changes

- New `internal/librarian/` package implementing proactive knowledge extraction and gap detection
- New Ent schema `Inquiry` for persisting pending questions to the user
- New `LibrarianConfig` in the config system to control proactive behavior
- Extended `ContextRetriever` to inject pending inquiries into the system prompt
- Updated Librarian agent spec with proactive behavior instructions and new tool prefixes
- New agent tools: `librarian_pending_inquiries`, `librarian_dismiss_inquiry`
- New async `ProactiveBuffer` wired into the gateway's OnTurnComplete callback pipeline

## Capabilities

### New Capabilities
- `proactive-librarian`: Autonomous knowledge extraction from observations, knowledge gap detection, inquiry lifecycle management, and answer-to-knowledge conversion

### Modified Capabilities
- `context-retriever`: Added `LayerPendingInquiries` context layer and `InquiryProvider` interface for injecting pending questions into the system prompt
- `multi-agent-orchestration`: Updated librarian agent spec with proactive behavior instructions, new prefixes (`librarian_`), and additional routing keywords

## Impact

- **New files**: `internal/ent/schema/inquiry.go`, `internal/librarian/*.go` (7 files)
- **Modified files**: `internal/config/types.go`, `internal/knowledge/types.go`, `internal/knowledge/retriever.go`, `internal/orchestration/tools.go`, `internal/app/types.go`, `internal/app/tools.go`, `internal/app/wiring.go`, `internal/app/app.go`
- **Dependencies**: Requires `knowledge.enabled`, `observationalMemory.enabled`, and `librarian.enabled` config flags
- **Database**: New `inquiries` table via Ent schema migration
