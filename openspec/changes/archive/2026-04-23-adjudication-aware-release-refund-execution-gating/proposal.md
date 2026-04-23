# Proposal

## Why

The knowledge-exchange escrow dispute path already records canonical release-vs-refund adjudication, but the actual release and refund executors still need that adjudication to be enforced at execution time.

## What Changes

- make escrow adjudication update progression atomically
- require matching adjudication on `release_escrow_settlement` and `refund_escrow_settlement`
- deny execution when opposite-branch evidence already exists
- publish the first public architecture page for adjudication-aware release/refund execution gating
- wire the page into the architecture landing page, track page, and docs navigation
- sync the OpenSpec requirements for docs and meta tools
- archive the completed change after sync
