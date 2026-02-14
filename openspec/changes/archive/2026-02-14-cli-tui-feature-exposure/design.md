## Context

Observational Memory, Secret Management, and Output Scanning are fully implemented at the Core layer (`internal/memory/`, `internal/security/`). However, there is no CLI surface for users to manage these subsystems. The `lango doctor` command does not diagnose OM or output scanning configuration, and the onboard wizard lacks an Observational Memory form. Users must edit `lango.json` manually to configure OM and have no way to manage secrets outside of LLM tool calls.

The existing CLI architecture follows a consistent pattern: Cobra commands with lazy config loading via `cfgLoader func() (*config.Config, error)`, session store access via `session.NewEntStore`, and tabwriter/JSON output formatting.

## Goals / Non-Goals

**Goals:**
- Expose observational memory CRUD operations via `lango memory list|status|clear`
- Expose secret management via `lango security secrets list|set|delete`
- Provide security status overview via `lango security status`
- Add diagnostic checks for OM configuration and output scanning alignment
- Add Observational Memory configuration form to the onboard TUI

**Non-Goals:**
- Modifying Core layer behavior or APIs (only consuming existing interfaces)
- Adding new output scanning CLI commands (it is an internal subsystem)
- Changing the existing security or passphrase migration commands
- Adding remote/RPC crypto provider support to CLI secrets commands (local only)

## Decisions

### D1: Memory CLI as separate package `internal/cli/memory/`

**Decision**: Create a new `internal/cli/memory/` package rather than adding to an existing package.

**Rationale**: Memory management is a distinct concern from security. The existing `internal/cli/security/` package handles encryption/passphrase concerns. Keeping them separate follows the project's per-concern package structure (`cli/auth/`, `cli/doctor/`, `cli/onboard/`).

**Alternative**: Adding memory commands under `cli/security/` — rejected because memory is not a security feature.

### D2: Shared crypto initialization helper

**Decision**: Extract a `initLocalCrypto(cfg, store)` helper in `internal/cli/security/crypto_init.go` for passphrase resolution, salt management, and checksum verification.

**Rationale**: The secrets list/set/delete commands all need the same crypto setup pattern that `migrate-passphrase` uses. Centralizing this avoids duplication and ensures consistent passphrase resolution (env > config > interactive prompt).

**Alternative**: Inline crypto setup in each command — rejected due to significant code duplication across 3+ commands.

### D3: Doctor checks as pure config validators

**Decision**: ObservationalMemoryCheck validates config values only (thresholds, provider references). OutputScanningCheck queries the database for secret counts but does not require crypto.

**Rationale**: Doctor checks should be runnable without a passphrase. Crypto initialization would make doctor checks fail or require interactive input, which conflicts with the `--json` non-interactive mode.

### D4: Onboard TUI form placement

**Decision**: Place the Observational Memory form between Knowledge and Providers in the menu, using the `🔬` emoji for visual distinction from `🧠 Knowledge`.

**Rationale**: OM is conceptually adjacent to Knowledge (both deal with agent intelligence) but distinct enough to warrant a separate menu entry. The microscope emoji signals a different subsystem than the brain emoji.

## Risks / Trade-offs

- **[Risk] Secrets commands require local crypto only** → Commands explicitly depend on `LocalCryptoProvider`. If a user has `provider: "rpc"`, secrets CLI will not work. Mitigation: Document this limitation; future work can add RPC support.
- **[Risk] Memory clear is destructive** → `DeleteObservationsBySession` + `DeleteReflectionsBySession` are irreversible. Mitigation: Confirmation prompt (skip with `--force`), positional arg makes session key explicit.
- **[Risk] Doctor OutputScanningCheck opens database** → May fail on encrypted databases without passphrase. Mitigation: Detect encryption errors (out of memory / file is not a database) and return StatusSkip gracefully.
