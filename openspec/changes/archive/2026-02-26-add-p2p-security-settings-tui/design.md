## Context

The `lango settings` TUI editor is a Bubble Tea-based interactive configuration editor. It follows a consistent pattern: menu categories → form builders → config write-back via a centralized `UpdateConfigFromForm()` switch. All P2P and advanced security config types already exist in `internal/config/types.go` and are consumed by `internal/app/wiring.go`, but lack TUI exposure.

## Goals / Non-Goals

**Goals:**
- Expose all P2P and advanced security settings through the existing TUI settings editor
- Follow established patterns (form builders, state update switch, menu categories)
- Handle `*bool` config fields correctly with helper functions
- Maintain full test coverage for new forms and config mappings

**Non-Goals:**
- Changing config types or initialization logic
- Adding validation beyond what existing forms use (type checks, range checks)
- Implementing list management UIs for FirewallRules (complex struct arrays — out of scope)

## Decisions

### Split P2P into 5 sub-categories instead of 1 monolithic form
P2PConfig has 6 nested sub-domains totaling 30+ fields. A single form would be unwieldy. Splitting into P2P Network (14), ZKP (5), Pricing (3), Owner Protection (5), Sandbox (11) keeps each form manageable.

**Alternative**: One "P2P" form with all fields — rejected because forms lack section dividers in the current TUI framework.

### Separate security sub-categories instead of expanding existing Security form
The existing Security form has 15 fields (interceptor + signer). Adding Keyring (1), DB Encryption (2), and KMS (12) would create a 30-field form. Separate menu entries are clearer.

### Reuse existing patterns for complex types
- `[]string` → comma-separated text with `splitCSV()` (same as RAG Collections)
- `map[string]string` → `key:value` comma-separated with `parseCustomPatterns()` (same as PII custom patterns)
- `*bool` → new `derefBool()`/`boolPtr()` helpers (new pattern, minimal)

### Expand signer provider options in existing Security form
Adding KMS options (`aws-kms`, `gcp-kms`, `azure-kv`, `pkcs11`) to the existing dropdown avoids needing a separate form just for provider selection.

## Risks / Trade-offs

- [Menu length increases from 21 to 29 items] → Menu scrolls with j/k keys; acceptable for comprehensive settings
- [FirewallRules not editable in TUI] → Complex struct arrays need a list management UI; deferred to future change
- [`*bool` is new to the form system] → Contained to 2 fields with clear helper functions; well-tested
