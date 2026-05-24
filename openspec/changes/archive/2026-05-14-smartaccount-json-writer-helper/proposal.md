## Why

The smart-account CLI package repeated the same pretty-JSON writer setup across multiple subcommands. That duplication made maintenance noisier without adding behavior.

## What Changes

- route smart-account JSON responses through a shared package-local helper
- keep existing output shapes unchanged while reducing duplication

## Impact

- lower maintenance cost for the smart-account CLI package
- one place to adjust JSON writer behavior if needed later
