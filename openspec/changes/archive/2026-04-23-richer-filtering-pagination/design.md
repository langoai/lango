# Design

## Context

The first dead-letter browsing / status observation slice already landed:

- a transaction-centered read model
- a read-only dead-letter backlog list
- a per-transaction canonical status view

What was still missing was operator-grade backlog triage:

- adjudication-based filtering
- retry-attempt filtering
- receipt-ID text search
- basic pagination metadata
- lightweight detail hints for replay navigation

## Goals / Non-Goals

**Goals**

- add practical list filtering for the dead-letter backlog
- add simple `offset` / `limit` pagination plus `total`
- expose minimal navigation hints on detail reads
- keep the read model transaction-centered and read-only
- update public docs and sync the OpenSpec requirements

**Non-Goals**

- no actor filter
- no time-range filter
- no alternate sort modes
- no raw background-task snapshot dump
- no replay or repair action changes
