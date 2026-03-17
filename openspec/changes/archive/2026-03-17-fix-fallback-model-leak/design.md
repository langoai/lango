## Context

`ProviderProxy` in `internal/supervisor/proxy.go` manages primary/fallback provider routing. When the primary fails, it calls `Supervisor.Generate()` with the fallback provider ID and model. However, the `params` struct is passed by value with `params.Model` still set to the primary model name. `Supervisor.Generate()` (line 128-129) only overrides `params.Model` when it is empty, so the primary model leaks through.

Current call flow:
1. `proxy.Generate()` calls `supervisor.Generate(ctx, "openai-1", "gpt-5.3-codex", params{Model: "gpt-5.3-codex"})`
2. Primary fails → fallback path
3. `supervisor.Generate(ctx, "gemini-1", "gemini-3-flash-preview", params{Model: "gpt-5.3-codex"})` — **bug: Model still set**
4. Supervisor sees `params.Model != ""`, skips override → Gemini receives `gpt-5.3-codex`

## Goals / Non-Goals

**Goals:**
- Fix the fallback model leak so the correct model is used on fallback
- Add defense-in-depth validation to catch cross-provider model mismatches at startup and runtime
- No changes to `Supervisor.Generate()` semantics — the fix is localized to the caller

**Non-Goals:**
- Mid-stream fallback (retry after partial streaming) — out of scope
- Complete model registry or model aliasing system — the validation is heuristic only
- Modifying `openai.go` — ollama/github use the same code and can host any model

## Decisions

### D1: Copy params before fallback call (not modify Supervisor.Generate)

Reset `params.Model` in the proxy before calling fallback, rather than changing `Supervisor.Generate()` to always override with the `model` argument. This preserves backward compatibility — other callers of `Supervisor.Generate()` may rely on `params.Model` taking precedence.

Alternative: Make `model` arg always win in `Supervisor.Generate()`. Rejected because it changes the contract for all callers.

### D2: Heuristic prefix blocklist (not model registry)

A static prefix map (`modelExclusions`) that catches obviously wrong models. Simpler than maintaining a full model registry, and sufficient for the safety-net purpose.

Alternative: Query provider APIs for valid models. Rejected — adds latency, network dependency, and doesn't help at config validation time.

### D3: Validation at three layers

- **Config validation**: Fail fast at startup for clearly misconfigured provider-model pairs
- **Runtime validation**: Guard in Gemini/Anthropic `Generate()` for defense-in-depth
- **OpenAI excluded**: ollama and github use the same code and legitimately host any model

## Risks / Trade-offs

- [Heuristic may become stale] → Prefix list is conservative (only blocks known cross-provider prefixes). New model naming schemes may not be caught, but they also won't cause false positives.
- [False positives for custom models] → ollama/github have no exclusions; only cloud providers (OpenAI/Anthropic/Gemini) are gated, and their model names follow well-known prefixes.
