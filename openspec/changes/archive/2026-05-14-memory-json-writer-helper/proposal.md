## Why

The memory CLI package repeated the same pretty-JSON writer setup across multiple subcommands. That duplication increases maintenance cost without changing user-facing behavior.

## What Changes

- route memory CLI JSON responses through a shared package-local helper
- keep existing payload shapes unchanged while removing repeated encoder setup

## Impact

- lower maintenance cost for the memory CLI package
- one place to adjust pretty-JSON writer behavior if needed later
