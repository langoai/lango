# Design

## Context

The knowledge-exchange escrow dispute path now has:

- background post-adjudication execution
- canonical adjudication preserved across async dispatch

What is still missing is the first bounded retry slice that can resubmit failed background post-adjudication work and eventually mark terminal dead-letter failure.

## Goals / Non-Goals

**Goals**

- retry post-adjudication background execution up to three times
- use exponential backoff
- record retry scheduling and terminal dead-letter evidence
- keep canonical adjudication intact
- publish a bounded public architecture page and keep docs navigation aligned
- sync the OpenSpec requirements and archive the completed slice

**Non-Goals**

- no generic background-manager-wide retry policy
- no operator replay UI
- no dead-letter browsing surface
- no broader dispute engine behavior
