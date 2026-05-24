## Why

The README internal package tree currently shows only part of `internal/p2p/`. Several shipped subpackages such as `gitbundle`, `ontologybridge`, `provenanceproto`, `trustpolicy`, and `workspace` are missing entirely.

## What Changes

- add the missing shipped `internal/p2p` subpackage rows to the README internal tree
- add an executable guard that requires the full current `internal/p2p` subtree
- sync the main docs-only and test-coverage specs

## Impact

- more complete internal package inventory
- better discoverability of collaborative and trust/provenance runtime slices
- stronger regression protection against partial subtree docs
