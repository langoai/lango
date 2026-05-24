# Design

## Context

The knowledge-exchange escrow dispute path now has:

- funded escrow
- dispute hold evidence
- release-vs-refund adjudication
- separate release and refund executors

What is still missing is the first executor-contract slice that says adjudication must match the selected release or refund path before execution may proceed.

## Goals / Non-Goals

**Goals**

- make adjudication atomically update progression and branch decision
- enforce matching adjudication on release and refund execution
- deny execution when opposite-branch evidence already exists
- publish a bounded public architecture page and keep docs navigation aligned
- sync the OpenSpec requirements and archive the completed slice

**Non-Goals**

- no automatic post-adjudication execution
- no keep-hold or re-escalation states
- no broader dispute engine behavior
- no human adjudication UI
