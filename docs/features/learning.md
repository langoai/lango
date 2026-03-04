---
title: Learning System
---

# Learning System

Lango includes a self-learning system that observes tool execution, extracts knowledge from conversations, and builds a knowledge graph of error-fix relationships. Over time, the agent becomes better at handling recurring situations.

## Architecture

```
Tool Execution ──► Engine ──► Audit Log + Learning Store
                     │
                     ▼
              GraphEngine ──► Knowledge Graph (error→fix triples)
                     │
                     ▼
        ConversationAnalyzer ──► Knowledge + Learning entries
                     │
                     ▼
           SessionLearner ──► High-confidence session summaries
```

## Components

### Engine

The core `Engine` observes tool execution results via the `ToolResultObserver` interface:

```go
type ToolResultObserver interface {
    OnToolResult(ctx context.Context, sessionKey, toolName string,
        params map[string]interface{}, result interface{}, err error)
}
```

**On error**, the engine:
1. Extracts a normalized error pattern (removing UUIDs, timestamps, paths)
2. Searches for existing learnings matching the pattern
3. If a known fix exists with confidence >= 0.7, logs it for auto-application
4. Otherwise, saves the error as a new learning entry

**On success**, the engine:
1. Searches for learnings triggered by this tool
2. Boosts the confidence of matching entries

### Error Categorization

Errors are automatically categorized:

| Category | Trigger |
|----------|---------|
| `timeout` | Deadline exceeded, timeout errors |
| `permission` | Permission denied, access denied, forbidden |
| `provider_error` | API, model, provider, rate limit errors |
| `tool_error` | Tool-specific errors |
| `general` | All other errors |

### Error Pattern Normalization

The analyzer normalizes error messages for pattern matching by:
- Removing UUIDs
- Removing timestamps
- Replacing file paths with `<path>`
- Replacing port numbers with `:<port>`

### User Corrections

Users can explicitly teach the agent via `RecordUserCorrection()`, which saves a high-confidence learning entry that takes priority over auto-detected patterns.

## Graph Engine

The `GraphEngine` extends the base engine with knowledge graph relationships:

- **Error triples**: `error:tool:pattern` → `causedBy` → `tool:name`
- **Session triples**: `error:tool:pattern` → `inSession` → `session:key`
- **Similarity triples**: `error:pattern1` → `similarTo` → `error:pattern2`
- **Fix triples**: `error:pattern` → `resolvedBy` → `fix:description`

### Confidence Propagation

When a tool succeeds, the graph engine propagates confidence boosts to similar error-fix relationships. The propagation rate is configurable (default: 0.3), meaning 30% of the confidence delta is applied to related learnings.

## Conversation Analyzer

The `ConversationAnalyzer` uses LLM analysis to extract structured knowledge from conversation turns. It identifies:

| Type | Description |
|------|-------------|
| `fact` | Domain knowledge and verified information |
| `pattern` | Repeated workflows and approaches |
| `correction` | User corrections of agent behavior |
| `preference` | User preferences and requirements |

Extracted items include optional graph triple fields (`subject`, `predicate`, `object`) for automatic knowledge graph enrichment.

### Analysis Buffer

Conversation analysis runs asynchronously via an `AnalysisBuffer` that batches messages and triggers analysis when:
- The turn count exceeds `analysisTurnThreshold` (default: 5)
- The token count exceeds `analysisTokenThreshold` (default: 2000)

## Session Learner

The `SessionLearner` runs at session end to extract high-confidence learnings from the complete conversation. It:

1. Skips sessions shorter than 4 messages
2. Samples long sessions (> 20 messages) for efficient LLM processing
3. Only stores learnings with `high` confidence
4. Saves both knowledge entries and graph triples

### Sampling Strategy

For sessions longer than 20 messages:
- First 3 messages (context setting)
- Every 5th message (representative sample)
- Last 5 messages (conclusions)

## Auto-Apply Confidence Threshold

The minimum confidence required to auto-apply a learned fix is **0.7**. Learnings below this threshold are stored but not automatically suggested.

## Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| `knowledge.enabled` | `true` | Enable the knowledge and learning system |
| `knowledge.maxContextPerLayer` | `3` | Max context entries per retrieval layer |
| `knowledge.analysisTurnThreshold` | `5` | Turns before triggering conversation analysis |
| `knowledge.analysisTokenThreshold` | `2000` | Token count before triggering analysis |
| `agent.errorCorrectionEnabled` | `true` | Enable error correction via learned fixes |
| `graph.enabled` | `false` | Enable knowledge graph for relationship tracking |

## CLI Commands

```bash
lango learning status        # Show learning system configuration
lango learning history       # Show recent learning entries
```
