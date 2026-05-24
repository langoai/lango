## Why

The README internal package tree now includes the full `internal/p2p` subtree, but the parent `p2p/` summary still describes only the older networking/identity/firewall/ZKP slice and omits collaborative workspaces, git/provenance exchange, trust policy, and payments.

## What Changes

- update the parent `p2p/` summary in the README internal package tree to reflect the current shipped scope
- extend the existing P2P package-subtree guard so it enforces that parent summary too
- sync the main docs-only and test-coverage specs

## Impact

- more truthful top-level P2P package summary
- better discoverability of the collaboration and trust/payment layers
- stronger regression protection against stale narrow wording
