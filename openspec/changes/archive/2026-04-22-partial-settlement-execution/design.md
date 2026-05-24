# Design

## Context

The knowledge-exchange track now has:

- settlement progression
- direct actual settlement execution
- partial settlement execution service and `execute_partial_settlement`

What is still missing is the public architecture page and the OpenSpec closeout that make the first direct partial-settlement slice discoverable and keep the docs contract aligned with the landed code.

## Goals / Non-Goals

**Goals**

- publish a bounded public architecture page for direct partial settlement execution
- wire the page into the architecture landing page, P2P knowledge-exchange track, and docs navigation
- record the `execute_partial_settlement` meta tool in OpenSpec
- archive the completed change after syncing main specs

**Non-Goals**

- no multi-round partial execution
- no percentage-based hint model
- no escrow partial release
- no dispute engine design
- no human UI work

## Decisions

### 1. Treat partial settlement execution as a one-shot direct execution slice

The public page should describe only the currently landed direct path, with the remaining repeated-partial and escrow work left explicit.

### 2. Keep the track doc focused on landed slices plus remaining gaps

The track page should move partial settlement execution from follow-on work into landed work and leave multi-round partials, escrow completion, and dispute work as the remaining gaps.

### 3. Record the tool in `meta-tools`

The OpenSpec update should add `execute_partial_settlement` directly to the existing `meta-tools` contract.
