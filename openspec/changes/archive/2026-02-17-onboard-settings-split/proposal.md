# Proposal: Onboard/Settings Split

## Problem
The current `lango onboard` command is a full menu-based settings editor with 17 categories and free navigation. This is overwhelming for first-time users who need a guided setup, while advanced users want the full editor without the "onboard" branding.

## Solution
Split the single command into two:
1. **`lango onboard`** — Guided 5-step wizard for first-time setup with progress bar
2. **`lango settings`** — Full configuration editor using the current menu-based UI

## Approach
- Extract shared form components into `internal/cli/tuicore/` package
- Move existing menu-based editor to `internal/cli/settings/`
- Rewrite `internal/cli/onboard/` as a 5-step stepper wizard
- Both commands share the same `tuicore` types and `config` mapping

## Non-goals
- No changes to the config schema or storage format
- No changes to the underlying `configstore` encryption
