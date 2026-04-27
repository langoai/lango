# Design

## Context

The dead-letter browsing / status observation surface already supports current-submission lifecycle summaries, filters, and grouping.

What was still missing was a compact transaction-global signal that spans all submissions belonging to the same transaction.

## Goals / Non-Goals

**Goals**

- add transaction-global retry count to each backlog row
- add transaction-global any-match family set to each backlog row
- add matching transaction-global filters to the existing backlog list

**Non-Goals**

- no transaction-global dominant family
- no per-submission breakdown
- no timeline view
- no aggregation cache
