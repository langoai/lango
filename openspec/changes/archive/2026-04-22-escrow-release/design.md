# Design

## Context

The knowledge-exchange track now has:

- escrow recommendation execution through `create + fund`
- direct settlement execution
- partial settlement execution
- the first escrow release service and `release_escrow_settlement`

What is still missing is the public architecture page and the OpenSpec closeout that make the first escrow release slice discoverable and keep the docs contract aligned with the landed code.

## Goals / Non-Goals

**Goals**

- publish a bounded public architecture page for escrow release
- wire the page into the architecture landing page, P2P knowledge-exchange track, and docs navigation
- record the `release_escrow_settlement` meta tool in OpenSpec
- archive the completed change after syncing main specs

**Non-Goals**

- no refund slice
- no dispute-linked escrow branching
- no milestone-aware release
- no human UI work

## Decisions

### 1. Treat escrow release as the first funded-escrow settlement-completion slice

The public page should describe only the currently landed funded release path and not imply refund or dispute orchestration.

### 2. Keep the track doc focused on landed slices plus remaining gaps

The track page should move escrow release from follow-on work into landed work and leave refund, dispute-linked escrow handling, and milestone-aware release as remaining gaps.

### 3. Record the tool in `meta-tools`

The OpenSpec update should add `release_escrow_settlement` directly to the existing `meta-tools` contract.
