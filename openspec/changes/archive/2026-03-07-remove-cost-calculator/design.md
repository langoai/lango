## Context

The observability system currently includes a hardcoded model pricing table (`cost.go`) that estimates USD costs from token counts. No LLM provider offers a pricing API, so these values become stale on every model release. The `EstimatedCost` field flows through types, collector, tracker, store, API routes, and CLI — a deep cross-cutting concern.

## Goals / Non-Goals

**Goals:**
- Remove all cost estimation code and the hardcoded pricing table
- Remove `EstimatedCost` from all observability types, API responses, and CLI output
- Remove the `estimated_cost` column from the Ent schema
- Keep token count tracking fully intact

**Non-Goals:**
- Updating hardcoded default model names in `clitypes/providers.go` or `onboard/steps.go` (separate change)
- Adding external pricing API integration (no provider offers one)
- Changing context window budget values in `adk/state.go` (stable, not cost-related)

## Decisions

1. **Full removal over deprecation**: Cost fields are removed entirely rather than deprecated. The pricing data was never accurate enough to warrant a migration path — it was always a rough estimate.

2. **Ent schema regeneration**: Remove the `estimated_cost` field from the schema and run `go generate`. Existing database rows will have the column dropped on next migration. No data migration needed since the cost data was never reliable.

3. **Tracker signature simplification**: `NewTracker(collector, store, costCalc)` becomes `NewTracker(collector, store)`. The `costCalc` function parameter is removed entirely rather than made optional, since there is no cost calculation to perform.

## Risks / Trade-offs

- **Database migration**: Dropping `estimated_cost` column is a one-way change. Existing cost data is lost. → Acceptable because the data was never accurate.
- **API breaking change**: Consumers relying on `estimatedCost` fields in JSON responses will see them disappear. → This is intentional; inaccurate data is worse than missing data.
- **CLI output change**: Users accustomed to cost columns will no longer see them. → Token counts remain, which are the accurate metric.
