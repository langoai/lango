# Design

## Problem

The current product center of gravity is mission-native, but the interactive surface contract is still ambiguous. Today:

- code routes bare `lango` to `runCockpit(...)`
- some main specs say bare `lango` means cockpit
- another main spec still says bare `lango` means direct chat

That drift is no longer cosmetic. It blocks honest implementation of a workbench-first surface.

## Goals

- make bare `lango` a standalone mission workbench contract
- keep `lango cockpit` explicit and multi-page
- keep `lango chat` direct and focused
- make Mission Control reusable across workbench and cockpit
- close the existing main-spec contradiction around bare `lango`

## Non-Goals

- redesigning plain chat into a mission UI
- redesigning cockpit diagnostics pages
- changing the cockpit page set in this first slice
- changing the cockpit startup page in this first slice
- introducing page-lifetime EventBus unsubscribe semantics
- starting live channels from bare `lango`

## Surface Contract

The first Wave 6 slice defines three distinct interactive contracts:

1. `lango`
   - launches a standalone mission workbench
   - hosts Mission Control content without the full cockpit shell
   - hints to `lango chat` and `lango cockpit`

2. `lango chat`
   - remains transcript-first and focused
   - remains explicitly invoked through `lango chat`
   - does not become mission-aware by default

3. `lango cockpit`
   - remains the explicit multi-page shell
   - keeps the current page set and startup page in the first slice
   - keeps `--with-channels` as cockpit-only

## Reuse Contract

Wave 6 reuses the existing Mission Control assets instead of forking the mission domain:

- Mission Control page behavior remains shared
- proposal, approval, loop, and collaboration semantics remain shared
- the first slice may still keep those reusable assets under `internal/cli/cockpit/...`

This task documents the product split, not a package-namespace cleanup.

## Spec Delta Strategy

This change introduces one new main contract and updates the existing conflicting ones:

- add `mission-workbench-tui` as the new bare-`lango` contract
- modify `mission-control-tui` so Mission Control is reusable across workbench and cockpit
- modify `tui-cockpit-layout` so bare `lango` no longer aliases cockpit
- modify `cockpit-shell` so cockpit remains explicit rather than root-default
- modify `cockpit-pages` so page routing remains cockpit-internal and does not own the bare-`lango` contract
- modify `interactive-tui-chat` so `lango chat` remains direct focused chat and bare `lango` no longer means chat

## Honesty Constraints

The deltas must stay truthful to current intended first-slice implementation:

- cockpit page set remains unchanged
- cockpit startup page remains unchanged
- `--with-channels` remains cockpit-only
- no EventBus unsubscribe semantics are claimed

## Success Criteria

This OpenSpec task is complete when:

- the `surface-split-wave-six` change validates cleanly
- the change adds the new workbench contract
- the change removes the bare-`lango` ambiguity from the affected delta specs
