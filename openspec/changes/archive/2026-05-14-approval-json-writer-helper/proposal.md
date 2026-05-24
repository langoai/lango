## Why

The approval CLI package still carried a one-off pretty-JSON writer setup in its status command. That duplication is small, but it is still unnecessary maintenance noise.

## What Changes

- route approval CLI JSON responses through a shared package-local helper
- keep existing payload shapes unchanged while removing repeated encoder setup

## Impact

- lower maintenance cost for the approval CLI package
- one place to adjust pretty-JSON writer behavior if needed later
