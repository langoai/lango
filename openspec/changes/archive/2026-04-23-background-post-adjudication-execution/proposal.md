# Proposal

## Why

The knowledge-exchange escrow dispute path already supports adjudication and inline post-adjudication execution, but it still needs an async convenience path that can enqueue the selected release or refund branch onto the existing background task substrate.

## What Changes

- add optional `background_execute=true` to `adjudicate_escrow_dispute`
- enforce mutual exclusivity with `auto_execute`
- return a background dispatch receipt instead of a synchronous execution result
- publish the first public architecture page for background post-adjudication execution
- wire the page into the architecture landing page, track page, and docs navigation
- sync the OpenSpec requirements for docs and meta tools
- archive the completed change after sync
