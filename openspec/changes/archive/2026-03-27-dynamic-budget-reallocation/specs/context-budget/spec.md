## MODIFIED Requirements

### Requirement: ContextBudgetManager role
`ContextBudgetManager` SHALL act as a budget orchestrator (not just a static allocator). In addition to computing initial per-section budgets via `SectionBudgets()`, it SHALL support dynamic redistribution via `ReallocateBudgets(measured SectionTokens)` where empty sections donate budget to non-empty sections.
