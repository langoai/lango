## Context

The provider system has three provider implementations: OpenAI, Anthropic, and Gemini. The Supervisor initializes providers by iterating over `config.Providers` (a map keyed by user-chosen ID) and calling each provider's constructor. OpenAI's constructor already accepts an `id` parameter and uses it, but Anthropic and Gemini hardcode their IDs to `"anthropic"` and `"gemini"` respectively. When users choose custom config keys (e.g., `"gemini-api-key"`), the registry stores the provider under the hardcoded ID while the Supervisor later looks it up by config key — causing a "provider not found" error.

## Goals / Non-Goals

**Goals:**
- Make Anthropic and Gemini constructors accept an explicit `id` parameter, matching OpenAI's pattern
- Ensure the Supervisor passes the config map key as the provider ID for all provider types
- Maintain backward compatibility — existing configs with `"anthropic"` or `"gemini"` keys continue to work

**Non-Goals:**
- Refactoring the provider registry or registration mechanism
- Adding provider ID validation or normalization
- Changing the OpenAI or Ollama provider constructors (already correct)

## Decisions

**Decision: Add `id` as the first parameter after receiver-less params**

The `id` string is added as the first string parameter in each constructor, matching the OpenAI pattern (`NewProvider(id, apiKey, baseURL)`). This keeps all provider constructors consistent.

Alternative considered: Using a functional options pattern — rejected as over-engineering for a simple bug fix.

**Decision: Minimal signature change**

Only the constructor signatures and the Supervisor call sites change. No new interfaces, no new types. The `id` field already exists on both provider structs.

## Risks / Trade-offs

- [Breaking change for external callers] → Mitigated: these are internal packages, not part of the public API. Only the Supervisor calls these constructors.
- [Test updates required] → Mitigated: only `anthropic_test.go` has tests using `NewProvider`; Gemini has no test files.
