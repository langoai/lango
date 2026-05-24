# Design

## Context

The dead-letter browsing / status observation surface already supports:

- adjudication filtering
- retry-attempt filtering
- receipt-ID query
- actor filtering
- time-window filtering

What was still missing was a small but practical way to narrow the backlog by:

- latest dead-letter reason
- latest dispatch reference

## Goals / Non-Goals

**Goals**

- add reason substring filtering to the existing backlog list
- add dispatch-reference exact-match filtering to the existing backlog list
- keep the response shape unchanged
- keep the implementation read-only and transaction-centered

**Non-Goals**

- no detail-view changes
- no response-field additions
- no subtype filters
- no replay-count filters
- no sort-mode expansion
