# Settings TUI Discoverability Improvement

## Problem

The Settings TUI introduced a Basic/Advanced tier system but Advanced options (34+) were hidden behind a Tab toggle, making detailed settings hard to find. Additionally, Logging, Gatekeeper, and OutputManager configs had no UI exposure at all. On-chain escrow state persistence had a bug where settings weren't being saved.

## Solution

1. **Show All + Visual Badges**: Default `showAdvanced=true`, add `ADV` badge to advanced categories
2. **Smart Search Filters**: `@basic`, `@advanced`, `@enabled`, `@modified` prefix filters
3. **Missing Config Forms**: Logging, Gatekeeper, OutputManager forms exposed in TUI
4. **Welcome Screen Enhancement**: Category summary and filter tips on welcome screen
5. **`lango config get/set` CLI**: Dot-path config access with passphrase protection on writes
6. **On-Chain Escrow Bug Fix**: Missing state_update handlers for on-chain escrow and settlement fields
