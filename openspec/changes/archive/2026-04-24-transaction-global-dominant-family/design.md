# Design

## Context

The dead-letter browsing / status observation surface already supports:

- transaction-global total retry count
- transaction-global any-match family grouping
- local and transaction-level family filters

What was still missing was a compact single-value family summary across the full transaction lifecycle.

## Goals / Non-Goals

**Goals**

- add `transaction_global_dominant_family` to each backlog row
- add exact-match filtering on that value
- reuse the existing transaction-global aggregation path

**Non-Goals**

- no family count maps
- no weighted dominance
- no per-submission breakdown
- no detail-view changes
