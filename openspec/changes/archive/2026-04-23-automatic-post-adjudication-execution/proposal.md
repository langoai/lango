# Proposal

## Why

The knowledge-exchange escrow dispute path already records canonical adjudication and enforces branch-aware release/refund gates, but operators still need a convenience path that can adjudicate and immediately execute the selected branch in one call.

## What Changes

- add optional `auto_execute=true` to `adjudicate_escrow_dispute`
- inline the matching release or refund executor after successful adjudication
- return both the adjudication result and nested execution result
- publish the first public architecture page for automatic post-adjudication execution
- wire the page into the architecture landing page, track page, and docs navigation
- sync the OpenSpec requirements for docs and meta tools
- archive the completed change after sync
