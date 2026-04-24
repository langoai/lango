# Design

## Context

The dead-letter browsing / status observation surface already supports:

- raw subtype and family filtering
- total-count and any-match family grouping
- alternate sort modes

What was still missing was a compact single-value summary of which family dominates the current submission retry lifecycle.

## Goals / Non-Goals

**Goals**

- add `dominant_family` to each backlog row
- add `dominant_family` exact-match filtering
- keep the computation current-submission-trail-only

**Non-Goals**

- no family count map
- no weighted dominance
- no cross-submission dominance
- no detail-view changes
