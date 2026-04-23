# Proposal

## Why

The knowledge-exchange escrow dispute path already supports dead-letter evidence, retry, and replay, but operators still need a read-only visibility surface to see the current dead-letter backlog and inspect a specific transaction before deciding whether to replay it.

## What Changes

- add a post-adjudication status service
- add read-only meta tools for backlog and detail views
- publish the first public architecture page for dead-letter browsing / status observation
- wire the page into the architecture landing page, track page, and docs navigation
- sync the OpenSpec requirements for docs and meta tools
- archive the completed change after sync
