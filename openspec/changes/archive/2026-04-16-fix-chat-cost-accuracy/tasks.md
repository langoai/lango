# Tasks: fix-chat-cost-accuracy

- [x] Add EstimatedCostUSD to TokenUsageTeaMsg and subscriber
- [x] Add turnCostUSD to ChatModel, accumulate from events
- [x] Remove provider.EstimateCostUSD recomputation in DoneMsg
- [x] Reset turnCostUSD in DoneMsg, ErrorMsg, and submitCmd (3 sites)
- [x] Remove unused provider import
- [x] Verify: go build ./... — ALL PASS
