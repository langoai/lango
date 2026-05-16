## Why

The payment CLI package repeated the same pretty-JSON writer setup across multiple subcommands. That duplication increases maintenance overhead without changing behavior.

## What Changes

- route payment CLI JSON responses through a shared package-local helper
- keep existing output shapes unchanged while removing repeated encoder setup

## Impact

- lower maintenance cost for payment CLI
- one place to adjust pretty-JSON writer behavior later if needed
