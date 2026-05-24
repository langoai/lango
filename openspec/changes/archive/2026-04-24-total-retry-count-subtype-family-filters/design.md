# Design

## Context

The dead-letter browsing / status observation surface already supports:

- actor/time filters
- reason/dispatch filters
- manual retry count and raw subtype filtering
- alternate sort modes

What was still missing was:

- total retry count across the current submission retry lifecycle
- a grouped family view for the latest retry/dead-letter subtype

## Goals / Non-Goals

**Goals**

- add total retry-count filtering to the existing backlog list
- add latest subtype-family filtering to the existing backlog list
- expose the corresponding values in each row

**Non-Goals**

- no cross-submission aggregation
- no dominant family
- no any-match family
- no detail-view changes
