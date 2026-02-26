## Context

`mapCategory()` and `mapKnowledgeCategory()` are internal functions that translate LLM-produced type strings into `entknowledge.Category` enum values. Both silently default to `CategoryFact` for unrecognized input, creating a hallucination risk when these "facts" are injected into the agent system prompt. The same unsafe raw-cast pattern also exists in `InquiryProcessor` and the `save_knowledge` tool.

## Goals / Non-Goals

**Goals:**
- Eliminate silent miscategorization by returning errors for unrecognized types
- Add missing `"pattern"` and `"correction"` cases to all mapping functions
- Validate category values before persisting in all code paths
- Maintain backward compatibility — existing valid types continue to work identically

**Non-Goals:**
- Changing the ent schema (already has `CategoryPattern`/`CategoryCorrection`)
- Adding new UI/CLI commands
- Modifying the LLM prompts beyond adding missing type options to the observation analyzer

## Decisions

**1. Return `(Category, error)` instead of silent fallback**
- Rationale: Callers can log and skip unknown types rather than polluting the knowledge store with misclassified data
- Alternative considered: Using `CategoryValidator` at call sites — rejected because it duplicates the switch logic and doesn't add the missing cases

**2. Callers skip + warn on error**
- Rationale: Non-fatal handling keeps the pipeline running while preventing bad data from being stored
- The extraction/inquiry continues processing remaining items

**3. `save_knowledge` tool uses `CategoryValidator` before cast**
- Rationale: This is the tool boundary (external input from LLM tool calls). The ent-generated validator is the canonical validation source and catches any future enum additions automatically

## Risks / Trade-offs

- [Risk] LLM outputs a type that was previously silently accepted → Mitigation: Warning log provides visibility; the data simply isn't stored rather than being stored incorrectly
- [Risk] Future category additions require updating switch statements → Mitigation: Tests cover all valid cases; a missing case will surface as a test failure
