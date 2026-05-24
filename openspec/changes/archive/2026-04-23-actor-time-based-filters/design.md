# Design

## Context

The dead-letter browsing / status observation surface already supports:

- adjudication filtering
- retry-attempt filtering
- receipt-ID query
- offset / limit pagination

What was still missing was a lightweight way to triage the backlog by:

- latest manual replay actor
- latest dead-letter time window

## Goals / Non-Goals

**Goals**

- add actor/time-based filters to the existing backlog list
- expose the corresponding actor/time values in each backlog row
- keep the read model transaction-centered and read-only
- continue deriving status only from receipts and receipt trail evidence

**Non-Goals**

- no detail view changes
- no raw background-task bridge
- no alternate sort modes
- no reason substring filter
- no broader observability cockpit work
