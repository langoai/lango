# Design

## Context

The dead-letter browsing / status observation surface already supports:

- adjudication filtering
- retry-attempt filtering
- receipt-ID query
- actor/time filtering
- reason/dispatch filtering

What was still missing was a compact way to:

- filter by latest retry/dead-letter subtype
- filter by manual replay count
- sort by the latest dead-letter, retry-attempt, or manual replay activity

## Goals / Non-Goals

**Goals**

- add subtype filtering to the existing backlog list
- add manual replay count range filtering to the existing backlog list
- add bounded alternate sort modes to the existing backlog list
- expose the corresponding sort/filter values in each row

**Non-Goals**

- no detail-view changes
- no custom sort order
- no multi-column sort
- no total retry-count filters
- no subtype-family grouping
