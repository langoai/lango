## Why

When the primary provider fails and fallback is triggered, `ProviderProxy.Generate()` passes the original `params` (with `params.Model` still set to the primary model name, e.g., `gpt-5.3-codex`) to the fallback provider. `Supervisor.Generate()` only applies the fallback model when `params.Model == ""`, so the primary model leaks through to the fallback provider — causing requests like `gemini-3-flash-preview` provider receiving `gpt-5.3-codex` as the model name, resulting in API errors (e.g., Gemini URL: `models/gpt-5.3-codex:streamGenerateContent`).

## What Changes

- **Fix**: `proxy.go` — copy `params` before fallback call and reset `Model` to `""` so `Supervisor.Generate()` correctly applies the fallback model.
- **Safety net**: Add heuristic provider-model compatibility validation (`ValidateModelProvider`) that catches obviously wrong model names at runtime and startup.
- **Startup validation**: `config.Validate()` — verify `fallbackProvider` exists in providers map; check primary and fallback model-provider compatibility.
- **Runtime validation**: Gemini and Anthropic providers validate the model name before making API calls.

## Capabilities

### New Capabilities
- `provider-model-validation`: Heuristic prefix-based blocklist to detect cross-provider model routing errors (e.g., `gpt-*` sent to Gemini).

### Modified Capabilities
- `supervisor-architecture`: Fallback path in `ProviderProxy` resets `params.Model` before delegating.
- `config-system`: `Validate()` adds `fallbackProvider` existence check and model-provider compatibility checks.
- `provider-interface`: Gemini and Anthropic providers add `ValidateModelProvider` guard in `Generate()`.

## Impact

- `internal/supervisor/proxy.go` — critical bug fix
- `internal/provider/validate.go` — new file (heuristic validation)
- `internal/config/loader.go` — extended validation
- `internal/provider/gemini/gemini.go` — runtime guard
- `internal/provider/anthropic/anthropic.go` — runtime guard
- New test files: `proxy_test.go`, `validate_test.go`
