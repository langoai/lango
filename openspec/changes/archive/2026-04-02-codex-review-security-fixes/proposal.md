## Why

Two rounds of Codex review identified 9 security issues (5 P1, 4 P2) in the P2P security hardening patch. Several hardening features were implemented but not actually wired, allowing silent bypass. These fixes close the gaps so that every advertised security control is enforced at runtime.

## What Changes

**Round 1 fixes (5 issues):**
- Fix fs_delete P2P context marker mismatch — use `ctxkeys.IsP2PRequest` instead of local `IsP2PContext`
- Wire `SetSafetyGate()` in `app.go` to connect tool safety-level checking to the P2P handler
- Add DNS resolution to URL validator to block hostnames resolving to private IPs
- Add post-navigation URL re-validation to catch redirect-based SSRF
- Fix `MaxSessions <= 0` semantics to mean unlimited (no eviction)

**Round 2 fixes (4 issues):**
- Update all ADK version references in docs from v0.5.0 to v0.6.0 to match `go.mod`
- Add `MaxSafetyLevel: "moderate"` default in `DefaultConfig()` and handle invalid values
- Honor `requireContainer` by refusing subprocess fallback when container mode is required
- Move safety gate check before payment gate in paid tool invocation flow

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `tool-filesystem`: P2P delete restriction now uses the correct context key
- `p2p-protocol`: Safety gate wired and enforced; paid-tool flow reordered
- `tool-browser`: URL validation includes DNS resolution and post-redirect re-validation
- `container-sandbox`: `requireContainer` honored in app wiring (no subprocess fallback)
- `observability`: `MaxSessions <= 0` means unlimited

## Impact

- **Code**: `internal/tools/filesystem/filesystem.go`, `internal/app/app.go`, `internal/tools/browser/url_validator.go`, `internal/tools/browser/browser.go`, `internal/tools/browser/tools.go`, `internal/p2p/protocol/handler.go`, `internal/config/loader.go`, `internal/observability/collector.go`
- **Docs**: `docs/architecture/index.md`, `docs/architecture/overview.md`, `docs/architecture/project-structure.md`
- **CI**: `.github/workflows/ci.yml` docs-version-check now passes
- **Behavior**: All P2P security controls now actually enforce their advertised policies
