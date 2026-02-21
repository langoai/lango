## Why

The project grew rapidly and const/enum/error/type patterns were not consistently applied. Magic strings are scattered across 6+ files, "failed to" error prefixes violate project rules in 100+ places, callback/token types are duplicated, and existing typed enums lack `Valid()`/`Values()` methods. This refactoring establishes a systematic code style with the `Enum[T]` interface + `Valid()`/`Values()` pattern across the entire codebase.

## What Changes

- Create shared `internal/types/` package for Enum interface, callbacks, token estimation, and cross-cutting types
- Introduce typed enums: `ChannelType`, `ProviderType`, `MessageRole`, `Confidence`, plus package-local enums
- Add `Valid()`/`Values()` methods to all existing typed enums (8+ types)
- Create sentinel errors in package-level `errors.go` files, replacing string matching with `errors.Is()`
- Remove "failed to" prefix from 100+ error messages to comply with project error guidelines
- Consolidate duplicate callback types (`EmbedCallback`, `TripleCallback`) and token estimation logic
- **BREAKING**: Rename `internal/cli/common/` to `internal/cli/clitypes/` (Go naming guideline)
- Introduce structured error types (e.g., `gateway.RPCError`)
- Consolidate duplicate `SenderFunc` types into `types.RPCSenderFunc`
- Strengthen config field types (`ProviderConfig.Type` → `types.ProviderType`)

## Capabilities

### New Capabilities
- `shared-types`: Central `internal/types/` package with Enum interface, callback types, token estimation, confidence/channel/provider/role enums
- `sentinel-errors`: Package-level sentinel errors replacing string matching patterns
- `enum-validation`: `Valid()`/`Values()` methods on all typed enum types

### Modified Capabilities
- `error-handling`: Error message format standardization (remove "failed to" prefix)
- `config-types`: Provider config type strengthening with typed enums

## Impact

- **Code**: 60+ files across `internal/` packages modified
- **APIs**: No external API changes; internal interface signatures updated (callback types, config fields)
- **Dependencies**: No new external dependencies; `internal/types/` is a leaf package (stdlib only)
- **Breaking**: `cli/common` → `cli/clitypes` rename affects 3 importers
- **Risk**: Low per-PR; each phase independently mergeable
