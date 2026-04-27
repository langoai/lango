# Design

## Context

The knowledge-exchange escrow dispute path now has:

- dead-letter evidence
- retry handling
- manual replay

What is still missing is the first read-only operator surface for listing current dead-lettered executions and inspecting the current canonical status of a specific transaction.

## Goals / Non-Goals

**Goals**

- list current dead-lettered post-adjudication executions
- inspect one transaction's canonical status and latest retry/dead-letter summary
- keep the read model transaction-centered and read-only
- publish a bounded public architecture page and keep docs navigation aligned
- sync the OpenSpec requirements and archive the completed slice

**Non-Goals**

- no replay or repair action
- no raw background-task snapshot dump
- no richer filtering or pagination
- no broader dispute engine behavior
