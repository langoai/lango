## Context

The approval system currently uses a `bool` return from `Provider.RequestApproval` — each tool invocation requires explicit user approval. For power users running the same dangerous tool repeatedly (e.g., multiple `exec` calls during debugging), re-approving every call creates unnecessary friction.

The system already has per-session scoping via session keys and a composite provider that routes approval requests to the correct channel. The `wrapWithApproval` function in `internal/app/tools.go` is the single point where approval gating occurs.

## Goals / Non-Goals

**Goals:**
- Allow users to grant persistent approval for a specific tool within a session via "Always Allow"
- Maintain backward compatibility with existing approve/deny flows
- Keep grants in-memory (no DB persistence) for simplicity and security
- Add "Always Allow" UI affordance to all channels (Telegram, Discord, Slack, TTY, Gateway)

**Non-Goals:**
- Cross-session persistence (grants intentionally reset on restart)
- Per-user grant tracking (scoped to session key, which may include user info)
- Admin revocation UI (programmatic `Revoke`/`RevokeSession` available for future use)
- Granular parameter-based grants (e.g., "always allow exec only for `ls`")

## Decisions

### 1. `ApprovalResponse` struct over extended bool

**Decision**: Replace `bool` return with `ApprovalResponse{Approved, AlwaysAllow}` struct.

**Rationale**: A struct is extensible for future fields (e.g., expiration, conditions) without further interface changes. Using separate `bool` return values would complicate every call site.

**Alternative**: Add a separate `RequestApprovalWithGrant()` method — rejected because it doubles the interface surface and every provider must implement both.

### 2. In-memory GrantStore with null-byte key separator

**Decision**: `GrantStore` uses `map[string]struct{}` with keys formed as `sessionKey + "\x00" + toolName`.

**Rationale**: Null byte cannot appear in session keys or tool names, preventing collision. A `map[string]struct{}` is the idiomatic Go set pattern with zero per-entry allocation overhead.

**Alternative**: Nested `map[string]map[string]struct{}` — rejected for added complexity with no real benefit given the small grant set.

### 3. Grant check before provider call

**Decision**: In `wrapWithApproval`, check `GrantStore.IsGranted()` first, only call the provider if no grant exists.

**Rationale**: Avoids sending approval messages to channels when the user has already granted permanent access. This is the correct UX — no visual noise for already-granted tools.

### 4. Backward-compatible Gateway protocol

**Decision**: The `alwaysAllow` field in the `approval.response` JSON is optional, defaulting to `false` when absent.

**Rationale**: Existing companion apps that don't know about "Always Allow" continue working without changes. New companions can send the field when their UI supports it.

## Risks / Trade-offs

- **[Session key ambiguity]** → Grants are tied to session keys. If a session key changes mid-conversation (e.g., channel migration), grants don't carry over. This is acceptable and actually desirable for security.
- **[No expiration]** → Grants last until app restart. For long-running deployments, a single "Always Allow" persists indefinitely within that session. Mitigated by the `RevokeSession` API available for future admin tools.
- **[Interface breaking change]** → All `Provider` implementations must be updated simultaneously. Mitigated by doing all changes atomically in one commit.
