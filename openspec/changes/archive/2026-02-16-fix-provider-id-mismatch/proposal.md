## Why

When users register providers with custom names (e.g., `"gemini-api-key"`) during onboarding, the Anthropic and Gemini providers ignore the config key and use hardcoded IDs (`"anthropic"`, `"gemini"`). The Supervisor then fails to find the provider by config key, causing a runtime error. OpenAI provider already handles this correctly.

## What Changes

- Anthropic provider `NewProvider` accepts an `id` parameter instead of hardcoding `"anthropic"`
- Gemini provider `NewProvider` accepts an `id` parameter instead of hardcoding `"gemini"`
- Supervisor passes the config key as the provider ID for both Anthropic and Gemini (matching the existing OpenAI pattern)
- Tests updated to verify custom ID propagation

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `provider-anthropic`: Constructor now requires explicit ID parameter to match registry key
- `provider-registry`: All provider constructors now consistently accept config key as ID

## Impact

- `internal/provider/anthropic/anthropic.go` — signature change: `NewProvider(apiKey)` → `NewProvider(id, apiKey)`
- `internal/provider/gemini/gemini.go` — signature change: `NewProvider(ctx, apiKey, model)` → `NewProvider(ctx, id, apiKey, model)`
- `internal/supervisor/supervisor.go` — caller updated to pass `id` to both constructors
- `internal/provider/anthropic/anthropic_test.go` — tests updated with new signature
