# Proposal: Codebase Consolidation

## Problem
As the project rapidly expanded with P2P, Security, KMS, Sandbox, and ZKP features, duplicate boilerplate patterns and underused packages accumulated. Five async buffer implementations share nearly identical lifecycle code, three small packages (ctxutil, passphrase, zkp) are imported by only 1-2 consumers, and CLI --help output lacks logical grouping.

## Solution
1. **Generic AsyncBuffer**: Create `internal/asyncbuf/` with `BatchBuffer[T]` and `TriggerBuffer[T]` generics, then migrate all 5 existing buffers to thin wrappers.
2. **Package Consolidation**: Merge `ctxutil` into `types`, `passphrase` into `security/passphrase`, `zkp` into `p2p/zkp`.
3. **CLI UX**: Add Cobra command groups and cross-references between config-related commands.
4. **Type Deduplication**: Consolidate identical `MessageProvider` type (skipped due to import cycle).

## Goals
- Reduce boilerplate by ~400 lines across 5 buffer packages.
- Improve package tree clarity by merging orphaned packages into logical parents.
- Improve CLI discoverability via grouped --help output.
- Zero breaking changes to any public API.

## Non-Goals
- Restructuring P2P/Security packages (already well-organized).
- Moving `session.Message` into `types` (would require large refactor).
- Adding new features or changing behavior.
