# Surface Split Wave 6

## Why

Lango now has a real mission-native control surface, but the CLI still exposes that surface through two commands that behave the same:

- `lango`
- `lango cockpit`

That makes the product contract muddy at the exact point where missions, proposals, loops, and collaboration have become first-class. It also leaves current OpenSpec and public truth drifting between three incompatible claims:

- bare `lango` means cockpit
- bare `lango` means Mission Control inside cockpit
- bare `lango` means direct chat

Wave 6 exists to make the interactive surfaces explicit and honest.

## What Changes

This change creates the first surface-split slice:

- bare `lango` becomes a standalone mission workbench contract
- `lango cockpit` remains the explicit multi-page operator dashboard
- `lango chat` remains the direct focused chat surface
- Mission Control becomes a reusable mission surface shared by workbench and cockpit
- the OpenSpec deltas close the existing bare-`lango` main-spec drift

## Scope Guardrails

Wave 6 Task 1 is spec-only and intentionally narrow:

- no code changes
- no public docs changes
- no second mission projection system
- no redesign of plain chat into a mission UI
- no change to cockpit page set or cockpit startup page in the first slice
- no claim of EventBus unsubscribe semantics
- no claim that bare `lango` starts live channels

## User Impact

After the full wave lands, users should be able to infer the right entry point from the command name alone:

- `lango`: mission workbench
- `lango chat`: focused chat
- `lango cockpit`: explicit dashboard

This task creates the OpenSpec change needed to make that split executable and honest.
