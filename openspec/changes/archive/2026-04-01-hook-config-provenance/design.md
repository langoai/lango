## Context

The provenance subsystem currently records checkpoints tied to RunLedger journal events and manual creation. These checkpoints capture _what_ happened but not _how_ the system was configured when it happened. For session reproducibility, we need to know the config fingerprint and hook registry state at session start.

The `CheckpointService.create()` currently has a fixed signature without metadata support. The `Checkpoint` struct already has a `Metadata map[string]string` field but it is never populated.

## Goals / Non-Goals

**Goals:**
- Capture config fingerprint + hook registry snapshot at session start as a provenance checkpoint
- Make `create()` accept metadata so the existing `Metadata` field on `Checkpoint` is usable
- Provide `CreateManualWithMetadata` where `runID` is optional (session-init checkpoints precede runs)
- Ensure idempotency: only one config checkpoint per session per app instance

**Non-Goals:**
- Full config diff between sessions (future work)
- Config change detection during a session (would require config reload hooks)
- Persisting the full config snapshot (only fingerprint + hook summary)

## Decisions

### D1: Fingerprint algorithm — SHA-256 of JSON(ExplicitKeys) + JSON(AutoEnabled) + JSON(HooksConfig)

**Rationale**: These three values capture the user-specific config state (ExplicitKeys), auto-derived state (AutoEnabled), and hook configuration (HooksConfig). JSON serialization provides deterministic ordering via Go's `encoding/json` which sorts map keys. SHA-256 is standard and cheap.

**Alternative considered**: Hash entire Config struct — rejected because it includes volatile fields (DataRoot paths) that change across machines without affecting behavior.

### D2: Hook snapshot format — JSON array of `{name, priority}` objects

**Rationale**: `Name()` and `Priority()` are the identity of a hook. The JSON array is compact, human-readable, and sufficient for diff comparison. Storing it as a string in the metadata map keeps the interface simple.

### D3: Idempotency via sync.Map keyed by sessionKey

**Rationale**: The `rootSessionObserver` fires on both `Create()` and `Get()` in SessionServiceAdapter. Using `sync.Map.LoadOrStore` ensures exactly-once checkpoint creation per session per app instance without locks. The map is never cleaned because session keys are unique and the app process is finite.

### D4: Modify `create()` signature to accept metadata

**Rationale**: Adding metadata as the last parameter to the internal `create()` method is the minimal change. Existing callers (`CreateManual`, `OnJournalEvent`) pass `nil`. This avoids a separate code path.

## Risks / Trade-offs

- [Risk] `sync.Map` grows unboundedly per app instance → Acceptable because session count per app run is bounded (typically <1000) and entries are tiny (string key + bool value).
- [Risk] Config fingerprint does not capture all behavioral state (e.g., environment variables, provider API keys) → Acceptable for v1; fingerprint covers user-intent config. Provider state is a future enhancement.
- [Trade-off] Storing hook snapshot as JSON string in metadata rather than a typed field → Simpler interface but requires JSON parsing for programmatic access. Acceptable since this is diagnostic data, not a hot path.
