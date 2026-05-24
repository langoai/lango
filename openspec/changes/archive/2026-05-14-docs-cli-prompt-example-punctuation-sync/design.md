## Context

Recent CLI prompt unification work standardized destructive confirmation prompts on the shared helper form `... [y/N]: `. A handful of public examples still show the old form without the colon.

## Goals / Non-Goals

**Goals:**
- Make the documented examples match current runtime prompt output exactly

**Non-Goals:**
- Changing any runtime prompt text
- Revising unrelated CLI docs in this turn

## Decisions

Update only the affected example lines where the runtime output is already verified by command tests.
Rationale: this keeps the docs change tightly scoped and evidence-backed.

## Risks / Trade-offs

- [Trade-off] This is a narrow docs-only change. → Mitigation: it directly improves documentation trustworthiness without introducing runtime risk.
