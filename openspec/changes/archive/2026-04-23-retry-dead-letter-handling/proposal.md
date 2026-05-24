# Proposal

## Why

The knowledge-exchange escrow dispute path already supports background post-adjudication execution, but it still needs bounded retry and dead-letter handling when the async worker fails.

## What Changes

- add retry metadata to background tasks
- add a post-adjudication-specific retry hook with exponential backoff
- append retry scheduled and dead-letter evidence to the current submission receipt trail
- publish the first public architecture page for retry / dead-letter handling
- wire the page into the architecture landing page, track page, and docs navigation
- sync the OpenSpec requirements for docs and meta tools
- archive the completed change after sync
