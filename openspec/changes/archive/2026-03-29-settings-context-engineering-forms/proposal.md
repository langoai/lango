# Proposal: Settings TUI Forms + Doctor Validation for Context Engineering

## Problem
Steps 1-15 Context Engineering features (retrieval coordinator, relevance auto-adjust, context budget, context profile) have no Settings TUI forms or Doctor validation. Users must edit lango.json directly. Doctor doesn't validate retrieval config or budget allocation ratios.

## Solution
- 4 new Settings TUI forms: Context Profile, Retrieval, Auto-Adjust, Context Budget
- New RetrievalCheck doctor check with 7 validation rules
- Enhanced ContextHealthCheck: allocation sum (±0.001) + RAG-without-provider warning
- Context defaults in loader.go (spec allocation: 0.30/0.25/0.25/0.10/0.10)
- Downstream: CLI help text, README, docs/configuration.md, docs/cli/core.md updated
- Pre-existing migrate.go build fix (Load → Config return type mismatch)
