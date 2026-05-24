## Why

The P2P CLI package repeated the same pretty-JSON writer setup across many subcommands. That duplication increases maintenance cost without changing user-facing behavior.

## What Changes

- route P2P CLI JSON responses through a shared package-local helper
- keep existing payload shapes unchanged while removing repeated encoder setup

## Impact

- lower maintenance cost for the P2P CLI package
- one place to adjust pretty-JSON writer behavior if needed later
