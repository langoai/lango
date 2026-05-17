# Proposal: Config Set Explicit Keys

## Summary

Preserve encrypted profile `ExplicitKeys` when CLI mutation paths save a loaded profile, and mark context-related keys explicit when the user sets them through `lango config set`.

## Problem

`lango config set` currently saves the updated active profile with `explicitKeys=nil`. `lango onboard --profile <existing>` also loads an existing profile, drops the loaded explicit-key metadata, and saves with `explicitKeys=nil`. Both paths discard context-key metadata that bootstrap uses to distinguish intentional user settings from auto-enabled defaults.

When that metadata is lost, later bootstrap can treat intentional disables as legacy unset values and auto-enable context features such as knowledge, retrieval, observational memory, graph, or embedding provider selection.

## Goals

- Preserve existing `bootstrap.Result.ExplicitKeys` across `config set` saves.
- Mark the target key explicit when the user sets a key from `config.ContextRelatedKeys()`.
- Preserve loaded explicit keys when onboarding an existing profile.
- Save preset explicit keys when onboarding a new preset-backed profile.
- Keep the existing single-bootstrap and cleanup behavior for `config set`.
- Keep invalid path and save-error behavior unchanged.

## Non-Goals

- Do not change context auto-enable policy.
- Do not migrate legacy profiles proactively.
- Do not alter config get/export/import semantics.
