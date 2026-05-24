# Design

## Context

The knowledge-exchange track now has:

- settlement progression state
- a receipts-backed `apply_settlement_progression` meta tool
- direct settlement execution service and `execute_settlement`

What is still missing is the public architecture page and the OpenSpec closeout that make the first execution slice discoverable and keep the docs contract aligned with the landed code.

## Goals / Non-Goals

**Goals**

- publish a bounded public architecture page for actual settlement execution
- wire the page into the architecture landing page, P2P knowledge-exchange track, and docs navigation
- record the `execute_settlement` meta tool in OpenSpec
- archive the completed change after syncing main specs

**Non-Goals**

- no new escrow lifecycle work
- no partial-settlement execution design
- no dispute engine design
- no human UI work

## Decisions

### 1. Treat actual settlement execution as a direct-settlement first slice

The page should describe only the currently landed direct path and not imply escrow release or broader settlement orchestration.

### 2. Keep the track doc focused on landed slices plus remaining gaps

The track page should move actual settlement execution from follow-on work into landed work and leave partial settlement, escrow completion, and dispute engine work as remaining gaps.

### 3. Record the tool in `meta-tools`

The OpenSpec update should add `execute_settlement` directly to the existing `meta-tools` contract.
