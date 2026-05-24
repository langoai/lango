# Design

## Context

The dead-letter browsing / status observation surface already supports:

- actor/time filters
- reason/dispatch filters
- raw subtype and family filters
- count-based filters
- alternate sort modes

What was still missing was a compact way to expose and query the set of retry-lifecycle families touched by the current submission.

## Goals / Non-Goals

**Goals**

- add `any_match_families` to each backlog row
- add `any_match_family` membership filtering to the existing backlog list
- keep the implementation current-submission-trail-only

**Non-Goals**

- no dominant family
- no multi-select family query
- no family counts
- no detail-view changes
