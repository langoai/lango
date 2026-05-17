# Proposal: Sync P2P Tool Isolation Doc Defaults

## Summary

Keep public configuration documentation synchronized with the `DefaultConfig()` defaults for P2P tool isolation.

## Problem

`DefaultConfig()` currently sets `p2p.toolIsolation.maxMemoryMB` to `256`, and `docs/configuration.md` documents `256`, but `README.md` documents `512`. This gives operators conflicting guidance about the default resource limit for remote peer tool execution.

## Scope

- Add an executable documentation guard that compares the P2P tool isolation default rows in `README.md` and `docs/configuration.md` against `config.DefaultConfig()`.
- Correct stale public documentation values discovered by that guard.
- Keep runtime configuration behavior unchanged.

## Non-Goals

- Changing P2P tool isolation defaults.
- Redesigning the P2P sandbox or container runtime selection behavior.
- Rewriting the full configuration reference.
