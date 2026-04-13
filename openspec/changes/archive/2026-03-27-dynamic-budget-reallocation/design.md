# Design: dynamic-budget-reallocation

## Architecture

GenerateContent now uses a two-phase pipeline:

```
Phase 1 (Retrieve): All sections retrieve in parallel, no budget truncation
Phase 2 (Reallocate): Measure actual content tokens → ReallocateBudgets
Phase 3 (Assemble): Truncate + format each section with reallocated budgets
```

## Empty-Section Redistribution Algorithm

1. Compute initial budgets from ratios (same as SectionBudgets)
2. Sections with measured=0 donate their entire budget
3. Non-empty sections keep their full initial budget + proportional share of surplus
4. No recursive redistribution — surplus distributed once
5. All sections empty → all-zero budgets (Degraded=false)
6. Headroom is never redistributed — stays as safety margin

## Assembly Function Split

Monolithic `assembleXxxSection` functions split into:
- `retrieveXxxData` — fetches raw data (parallel in Phase 1)
- `formatXxxSection` — truncates + formats (sequential in Phase 3)

Old wrappers preserved for backward compatibility.
