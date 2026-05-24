# Proposal

## Why

The knowledge-exchange escrow dispute path already supports dead-letter evidence and bounded retry, but operators still need a manual way to replay a dead-lettered post-adjudication execution without clearing canonical adjudication or prior failure evidence.

## What Changes

- add a post-adjudication replay service
- add `manual-retry-requested` evidence
- add `retry_post_adjudication_execution`
- publish the first public architecture page for operator replay / manual retry
- wire the page into the architecture landing page, track page, and docs navigation
- sync the OpenSpec requirements for docs and meta tools
- archive the completed change after sync
