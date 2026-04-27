# Proposal

## Why

The knowledge-exchange escrow path already lands `create + fund`, `release`, and `refund`, but it still lacks the first dispute-aware hold slice that records when funded escrow is intentionally paused after canonical dispute handoff.

## What Changes

- add the first dispute hold service and `hold_escrow_for_dispute` meta tool
- append dispute hold success and failure evidence to the current submission receipt trail
- publish the first public architecture page for dispute hold
- wire the page into the architecture landing page, track page, and docs navigation
- sync the OpenSpec requirements for docs and meta tools
- archive the completed change after sync
