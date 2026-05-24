## Why

The public inventory docs still stop the workflow family at `run/list/status/cancel/history`, even though the repository already ships `lango workflow validate <file>` and documents it in the CLI index and README quick reference.

## What Changes

- update the README internal tree and architecture inventory to include workflow `validate`
- extend the existing remaining-inventory guard so it requires the workflow validate surface too
- sync the main docs-only and test-coverage specs

## Impact

- more truthful workflow discoverability in inventory docs
- better alignment between quick references and package-tree docs
- stronger regression protection for the workflow validate command
