# Proposal

## Why

The knowledge-exchange replay path already supports dead-letter recovery, but it still needs a policy gate so replay is not available to every actor by default.

## What Changes

- add actor- and outcome-aware replay authorization
- resolve actor from runtime context
- back replay authorization with config allowlists
- publish the first public architecture page for policy-driven replay controls
- wire the page into the architecture landing page, track page, and docs navigation
- sync the OpenSpec requirements for docs and meta tools
- archive the completed change after sync
