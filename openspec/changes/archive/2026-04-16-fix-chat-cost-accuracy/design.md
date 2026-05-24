# Design: Fix Chat Cost Accuracy

## Changes

1. Add `EstimatedCostUSD float64` to `TokenUsageTeaMsg`
2. Forward `e.EstimatedCostUSD` in subscriber
3. Add `turnCostUSD float64` field to `ChatModel`
4. Accumulate from events instead of recalculating
5. Reset at DoneMsg, ErrorMsg, and new turn start (submitCmd)
6. Remove unused `provider` import from chat.go
