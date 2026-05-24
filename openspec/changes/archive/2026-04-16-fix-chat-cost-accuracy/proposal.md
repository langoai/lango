# Proposal: Fix Chat Cost Accuracy

## Problem

Plain chat mode recomputes cost from `cfg.Agent.Model` instead of using the `EstimatedCostUSD` already carried by `TokenUsageEvent`. When auto model selection or fallback is used, `cfg.Agent.Model` is empty or stale, producing `$0` or wrong prices.

## Proposed Solution

Accumulate `EstimatedCostUSD` from `TokenUsageTeaMsg` alongside token counts. On turn completion, use the accumulated cost instead of recalculating from config.
