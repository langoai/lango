# Settings TUI Discoverability

## Overview

Improve settings TUI discoverability by showing all categories by default with visual tier badges, adding smart search filters, exposing missing config forms, and providing CLI get/set access.

## Requirements

### Requirement: All Categories Visible by Default

The `showAdvanced` field in `MenuModel` defaults to `true`. All categories (basic and advanced) are rendered on initial load. Users can toggle to basic-only view via Tab.

### Requirement: Visual Tier Badges

Advanced categories display an `ADV` badge (gray `#6B7280` background) after the description text. Basic categories have no badge for a clean default appearance. The badge style is defined as `BadgeAdvancedStyle` in `tui/styles.go`.

### Requirement: Smart Search Filters

The search bar (`/`) supports special prefix filters:

| Prefix | Behavior |
|--------|----------|
| `@basic` | Show only TierBasic categories |
| `@advanced` | Show only TierAdvanced categories |
| `@enabled` | Show categories whose feature is currently enabled |
| `@modified` | Show categories with unsaved changes (dirty state) |

Filter hints (`@basic  @advanced  @enabled  @modified`) are displayed when search mode is active but the query is empty.

### Requirement: Missing Config Forms

Three config sections previously had no TUI exposure:

| Form | Config Struct | Fields |
|------|--------------|--------|
| Logging | `LoggingConfig` | Level (select), Format (select), OutputPath (text) |
| Gatekeeper | `GatekeeperConfig` | Enabled, StripThoughtTags, StripInternalMarkers, StripRawJSON, RawJSONThreshold, CustomPatterns |
| Output Manager | `OutputManagerConfig` | Enabled, TokenBudget, HeadRatio, TailRatio |

All three are placed in the Core section with `TierAdvanced`.

### Requirement: Welcome Screen Enhancement

The welcome screen displays:
- Category count summary: `"{total} categories ({basic} basic, {advanced} advanced)"`
- Search/filter usage tips

### Requirement: CLI Config Get/Set/Keys

| Command | Auth | Description |
|---------|------|-------------|
| `lango config get <dot.path>` | Read-only (bootstrap passphrase for DB access) | Resolve value via reflect + mapstructure tags |
| `lango config set <dot.path> <value>` | Passphrase required (bootstrap) | Set value and save to encrypted profile |
| `lango config keys [prefix]` | None (static reflection) | List available config keys |

`config set` is passphrase-protected because bootstrap acquires the passphrase before DB access, preventing AI agents from silently modifying settings.

### Requirement: On-Chain Escrow State Persistence Fix

The `economy_escrow_onchain_*` and `economy_escrow_settlement_*` field handlers were missing from `state_update.go`, causing data loss when saving on-chain escrow settings through the TUI. All 11 handlers are added:

- `economy_escrow_onchain_enabled`, `_mode`, `_hub_address`, `_vault_factory`, `_vault_impl`, `_arbitrator`, `_token`, `_poll_interval`, `_confirmation_depth`
- `economy_escrow_settlement_receipt_timeout`, `_max_retries`
