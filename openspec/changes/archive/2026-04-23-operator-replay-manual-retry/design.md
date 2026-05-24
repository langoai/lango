# Design

## Context

The knowledge-exchange escrow dispute path now has:

- background post-adjudication execution
- bounded retry and dead-letter evidence

What is still missing is the first operator-facing replay slice that can take a dead-lettered execution and re-enqueue it through the existing background dispatch path.

## Goals / Non-Goals

**Goals**

- require dead-letter evidence and canonical adjudication for replay
- append `manual-retry-requested` evidence
- reuse the existing background dispatch path
- publish a bounded public architecture page and keep docs navigation aligned
- sync the OpenSpec requirements and archive the completed slice

**Non-Goals**

- no inline replay
- no generic background-task replay
- no dead-letter browsing UI
- no broader dispute engine behavior
