# Proposal: dynamic-budget-reallocation

## Problem
ContextBudgetManager computes fixed per-section budgets from allocation ratios. If a section is completely empty (e.g., no RAG configured), its budget is wasted — not redistributed to sections that have content.

## Solution
Transform the budget manager from a static allocator into an orchestrator via empty-section redistribution. Sections with measured token count of 0 donate their entire budget proportionally to sections that have content. Restructure GenerateContent to a two-phase flow: retrieve all → measure → reallocate → truncate+format.

## Scope
- `ReallocateBudgets(measured SectionTokens)` on ContextBudgetManager
- Two-phase GenerateContent (retrieve → reallocate → assemble)
- Split monolithic assembly functions into retrieve + format
- Token estimation helpers for pre-assembly measurement
- No config changes, headroom not redistributed
