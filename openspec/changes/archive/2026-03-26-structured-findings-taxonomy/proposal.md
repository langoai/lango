## Why

Knowledge ingestion quality directly limits retrieval quality. Currently, only 2 of 6 knowledge categories are actively populated by conversation/session analyzers (fact, preference), while rule/definition are ignored and pattern/correction are routed to Learning only. LLM prompts across 4 analyzers request different subsets of types, and there is no temporal distinction between evergreen facts and current-state facts that may change over time. Duplicate category mapping functions exist across packages.

This step standardizes all knowledge ingestion to the full 6-category taxonomy with temporal classification, improving storage quality before retrieval improvements in later steps.

## What Changes

- All 4 LLM analyzer prompts standardized to request all 6 knowledge types (rule, definition, preference, fact, pattern, correction) with temporal hint (evergreen/current_state)
- ConversationAnalyzer and SessionLearner routing expanded: ALL 6 types save as Knowledge; pattern/correction additionally save as Learning (dual-save for backward compat)
- Librarian analyzers (ProactiveBuffer, InquiryProcessor) also apply dual-save rule for pattern/correction
- `analysisResult` and `ObservationKnowledge` structs gain `Temporal` field, stored as tag on KnowledgeEntry
- `mapLearningCategory()` fixed to return `(Category, error)` instead of silent fallback to General
- Shared category mappers created: `knowledge.MapKnowledgeCategory()` and `knowledge.MapLearningCategory()` — single source of truth
- Content-dedup in `SaveKnowledge`: same `(category, content)` re-extraction is no-op (prevents version churn from Step 4)
- Shared `saveAnalysisResult()` helper consolidates duplicate save/dual-save/triple logic from ConversationAnalyzer and SessionLearner

## Capabilities

### New Capabilities
- `structured-findings-taxonomy`: Shared category mappers, content-dedup, temporal classification tags, dual-save routing, standardized 6-category LLM prompts

### Modified Capabilities
- `knowledge-store`: SaveKnowledge content-dedup (same category+content = no-op)
- `learning-engine`: mapLearningCategory returns error, all 6 types save as knowledge, dual-save for pattern/correction
- `conversation-analysis`: Prompt expanded to 6 types + temporal, routing changed to all-as-knowledge + dual-save

## Impact

- **learning/parse.go**: analysisResult gains Temporal field, mapLearningCategory returns error, shared saveAnalysisResult helper
- **learning/conversation_analyzer.go**: Prompt 6-type + temporal, saveResult delegates to shared helper
- **learning/session_learner.go**: Same as conversation_analyzer
- **librarian/types.go**: ObservationKnowledge gains Temporal field
- **librarian/observation_analyzer.go**: Prompt adds temporal hint
- **librarian/inquiry_processor.go**: Prompt 6-type + temporal, dual-save for pattern/correction
- **librarian/proactive_buffer.go**: Temporal tag in SaveKnowledge, dual-save for pattern/correction
- **librarian/parse.go**: matchedKnowledge gains Temporal field
- **knowledge/category.go**: NEW — shared MapKnowledgeCategory/MapLearningCategory
- **knowledge/store.go**: Content-dedup check in saveKnowledgeOnce
- **app/tools_meta.go**: save_knowledge description updated
- **README.md**: Knowledge Store description updated
