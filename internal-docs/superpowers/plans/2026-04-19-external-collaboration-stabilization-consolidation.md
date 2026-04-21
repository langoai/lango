# External Collaboration Stabilization And Consolidation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the completed external-collaboration audit into a coherent first implementation slice that truth-aligns identity, trust, pricing, negotiation, settlement, and the operator-facing team/workspace surfaces.

**Architecture:** This plan intentionally avoids a deep protocol rewrite. It stabilizes the existing runtime by aligning public surfaces with real behavior, unifying the most important trust defaults, and clarifying which control plane owns each external-exchange capability. Phase 3/4 live team/workspace APIs are explicitly treated as follow-on work once the Phase 1/2 external market surface is truthful and consistent.

**Tech Stack:** Go, Cobra CLI, Chi HTTP routes, MkDocs/Markdown docs, existing P2P runtime (`internal/p2p/*`), app wiring (`internal/app/*`)

---

## Scope Split

The completed audit covers four partially independent families:

- P2P identity / trust / reputation
- pricing / negotiation / settlement
- team formation / role coordination
- workspace / shared artifacts

This plan deliberately covers only the first executable stabilization/consolidation slice:

- make Phase 1/2 external exchange truthful and consistent,
- make Phase 3/4 docs and CLI surfaces honest about what is live vs server-backed vs tool-backed,
- preserve the current runtime architecture unless a change directly removes a documented inconsistency.

This plan does **not** add full live team/workspace HTTP APIs. That should be a separate follow-on plan after this slice lands.

## File Map

- Modify: `internal/cli/p2p/identity.go`
  - Make CLI identity output match its own help/docs by exposing DID when available.
- Create: `internal/cli/p2p/identity_test.go`
  - Add focused tests for the identity rendering/view payload.
- Modify: `internal/p2p/paygate/trust.go`
  - Canonicalize post-pay trust defaults.
- Modify: `internal/p2p/team/payment.go`
  - Align team payment trust defaults with the canonical external market threshold.
- Create: `internal/p2p/trustpolicy/defaults.go`
  - Single shared trust-threshold constant for P2P paygate/team payment.
- Create: `internal/p2p/trustpolicy/defaults_test.go`
  - Verify default-threshold invariants.
- Modify: `docs/features/p2p-network.md`
  - Truth-align identity, auth, pricing, payment-tier, team, workspace, and shared-artifact wording.
- Modify: `docs/features/economy.md`
  - Clarify that economy pricing/negotiation/escrow are policy engines layered above the P2P market path, not duplicate public quote surfaces.
- Modify: `docs/cli/p2p.md`
  - Make P2P operator docs honest about live vs server-backed vs tool-backed behavior.
- Modify: `internal/cli/p2p/pricing.go`
  - Clarify help text so this command is the provider-side public quote configuration surface.
- Modify: `internal/cli/economy/pricing.go`
  - Clarify help text so this command is the local dynamic pricing policy surface.
- Modify: `internal/cli/economy/negotiate.go`
  - Clarify help text so this command is the local negotiation engine/config surface.
- Modify: `internal/cli/p2p/team.go`
  - Tighten operator messaging so the current runtime-only/server-backed model is explicit and consistent.
- Modify: `internal/cli/p2p/workspace.go`
  - Tighten workspace CLI messaging to match actual behavior.
- Modify: `internal/cli/p2p/git.go`
  - Tighten git bundle CLI messaging to match actual behavior.
- Modify: `internal/app/p2p_routes.go`
  - Keep route comments and public contract text aligned with authenticated gateway behavior.
- Modify: `internal/app/p2p_routes_test.go`
  - Add route-level tests for identity/reputation surface expectations where practical.
- Modify: `docs/architecture/external-collaboration-audit.md`
  - Update the completed-audit narrative after implementation lands.

## Task 1: Truth-Align P2P Identity Surfaces

**Files:**
- Modify: `internal/cli/p2p/identity.go`
- Create: `internal/cli/p2p/identity_test.go`
- Modify: `docs/features/p2p-network.md`
- Modify: `docs/cli/p2p.md`
- Modify: `internal/app/p2p_routes.go`
- Modify: `internal/app/p2p_routes_test.go`

- [ ] **Step 1: Write the failing tests for the identity view model**

Create `internal/cli/p2p/identity_test.go`:

```go
package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildIdentityView_IncludesDIDWhenAvailable(t *testing.T) {
	view := buildIdentityView(
		"did:lango:v2:abcdef1234567890abcdef1234567890abcdef12",
		"QmPeer123",
		"secrets-store",
		[]string{"/ip4/127.0.0.1/tcp/9000"},
	)

	assert.Equal(t, "did:lango:v2:abcdef1234567890abcdef1234567890abcdef12", view["did"])
	assert.Equal(t, "QmPeer123", view["peerId"])
	assert.Equal(t, "secrets-store", view["keyStorage"])
}

func TestBuildIdentityView_PreservesListenAddresses(t *testing.T) {
	view := buildIdentityView(
		"did:lango:test",
		"QmPeer123",
		"legacy-keydir",
		[]string{"/ip4/127.0.0.1/tcp/9000", "/ip4/127.0.0.1/udp/9000/quic-v1"},
	)

	assert.Equal(t, []string{
		"/ip4/127.0.0.1/tcp/9000",
		"/ip4/127.0.0.1/udp/9000/quic-v1",
	}, view["listenAddrs"])
}
```

- [ ] **Step 2: Run the new identity tests and confirm they fail**

Run:

```bash
go test ./internal/cli/p2p/... -run 'TestBuildIdentityView' -count=1
```

Expected:

```text
FAIL
```

with an undefined-symbol error for `buildIdentityView`.

- [ ] **Step 3: Implement the identity view helper and wire DID into CLI output**

Modify `internal/cli/p2p/identity.go`:

```go
func buildIdentityView(did string, peerID string, keyStorage string, listenAddrs []string) map[string]interface{} {
	return map[string]interface{}{
		"did":         did,
		"peerId":      peerID,
		"listenAddrs": listenAddrs,
		"keyStorage":  keyStorage,
	}
}
```

and update the command body so it resolves a DID before rendering:

```go
did := ""
if deps.identity != nil {
	if got, err := deps.identity.DID(cmd.Context()); err == nil && got != nil {
		did = got.ID
	}
}

view := buildIdentityView(did, peerID, deps.keyStorage, listenAddrs)
if jsonOutput {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(view)
}

fmt.Println("P2P Identity")
if did != "" {
	fmt.Printf("  DID:          %s\n", did)
}
fmt.Printf("  Peer ID:      %s\n", peerID)
fmt.Printf("  Key Storage:  %s\n", deps.keyStorage)
fmt.Printf("  Listen Addrs:\n")
for _, a := range listenAddrs {
	fmt.Printf("    %s\n", a)
}
```

- [ ] **Step 4: Align docs with the actual identity/auth model**

Update `docs/features/p2p-network.md` and `docs/cli/p2p.md` so they say:

```md
Lango currently supports both legacy wallet-derived `did:lango:<hex>` identities and bundle-backed `did:lango:v2:<hash>` identities.

The CLI and `GET /api/p2p/identity` expose the active DID when one is available.

The `/api/p2p/*` routes are public only when gateway auth is disabled; when auth is configured, the whole subtree is protected by gateway auth.
```

and update the CLI example block to include DID:

```md
$ lango p2p identity
P2P Identity
  DID:          did:lango:v2:abcdef1234567890abcdef1234567890abcdef12
  Peer ID:      QmYourPeerId123...
  Key Storage:  secrets-store
  Listen Addrs:
    /ip4/0.0.0.0/tcp/9000
```

- [ ] **Step 5: Add/adjust route tests and rerun targeted verification**

Extend `internal/app/p2p_routes_test.go` with a route-level expectation:

```go
func TestP2PIdentityHandler_IncludesDIDWhenIdentityAvailable(t *testing.T) {
	p2pc := &p2pComponents{
		node:     newRouteTestNode(t),
		identity: newRouteTestIdentity(t, "did:lango:v2:abcdef1234567890abcdef1234567890abcdef12"),
	}

	handler := p2pIdentityHandler(p2pc)
	req := httptest.NewRequest("GET", "/api/p2p/identity", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "did:lango:v2:abcdef1234567890abcdef1234567890abcdef12", resp["did"])
}
```

Run:

```bash
go test ./internal/app/... ./internal/cli/p2p/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 6: Commit the identity truth-alignment slice**

Run:

```bash
git add internal/cli/p2p/identity.go internal/cli/p2p/identity_test.go internal/app/p2p_routes.go internal/app/p2p_routes_test.go docs/features/p2p-network.md docs/cli/p2p.md
git -c commit.gpgsign=false commit -m "fix: align p2p identity surfaces"
```

## Task 2: Canonicalize External Market Trust Defaults

**Files:**
- Create: `internal/p2p/trustpolicy/defaults.go`
- Create: `internal/p2p/trustpolicy/defaults_test.go`
- Modify: `internal/p2p/paygate/trust.go`
- Modify: `internal/p2p/team/payment.go`
- Modify: `docs/features/p2p-network.md`
- Modify: `docs/features/economy.md`
- Modify: `internal/config/types_p2p.go`

- [ ] **Step 1: Write the failing threshold-consistency test**

Create `internal/p2p/trustpolicy/defaults_test.go`:

```go
package trustpolicy

import "testing"

func TestDefaultPostPayThreshold(t *testing.T) {
	if DefaultPostPayThreshold != 0.8 {
		t.Fatalf("expected canonical post-pay threshold 0.8, got %v", DefaultPostPayThreshold)
	}
}
```

- [ ] **Step 2: Run the new trustpolicy test and confirm it fails**

Run:

```bash
go test ./internal/p2p/... -run 'TestDefaultPostPayThreshold' -count=1
```

Expected:

```text
FAIL
```

with a missing package or symbol error before the new file exists.

- [ ] **Step 3: Introduce one shared default and rewire paygate/team payment to use it**

Create `internal/p2p/trustpolicy/defaults.go`:

```go
package trustpolicy

const DefaultPostPayThreshold = 0.8
```

Modify `internal/p2p/paygate/trust.go`:

```go
import "github.com/langoai/lango/internal/p2p/trustpolicy"

const DefaultPostPayThreshold = trustpolicy.DefaultPostPayThreshold
```

Modify `internal/p2p/team/payment.go`:

```go
import "github.com/langoai/lango/internal/p2p/trustpolicy"

const DefaultPostPayThreshold = trustpolicy.DefaultPostPayThreshold

func NewNegotiator(cfg NegotiatorConfig) *Negotiator {
	threshold := cfg.PostPayThreshold
	if threshold <= 0 {
		threshold = DefaultPostPayThreshold
	}
	// ...
}
```

- [ ] **Step 4: Truth-align the docs and config comments to the canonical default**

Update these comments/sections to `0.8` consistently:

```go
// PostPayMinScore is the minimum reputation score for post-pay eligibility (default: 0.8).
```

```md
| **PostPay** | ≥ `postPayMinScore` (default: 0.8) | Tool executes first, payment settles after |
| **PrePay**  | < `postPayMinScore` | Payment must confirm before tool execution |
```

Also add a short operator note:

```md
Lango uses a stricter threshold for deferred payment than for admission trust.
Admission trust and payment trust are separate gates and should be documented separately.
```

- [ ] **Step 5: Run targeted tests, then the broader P2P package tests**

Run:

```bash
go test ./internal/p2p/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 6: Commit the canonical-threshold change**

Run:

```bash
git add internal/p2p/trustpolicy/defaults.go internal/p2p/trustpolicy/defaults_test.go internal/p2p/paygate/trust.go internal/p2p/team/payment.go internal/config/types_p2p.go docs/features/p2p-network.md docs/features/economy.md
git -c commit.gpgsign=false commit -m "fix: unify external payment trust defaults"
```

## Task 3: Consolidate The Phase 1/2 Pricing And Settlement Story

**Files:**
- Modify: `docs/features/p2p-network.md`
- Modify: `docs/features/economy.md`
- Modify: `internal/cli/p2p/pricing.go`
- Modify: `internal/cli/economy/pricing.go`
- Modify: `internal/cli/economy/negotiate.go`
- Modify: `docs/architecture/external-collaboration-audit.md`

- [ ] **Step 1: Add the failing documentation assertions**

Create or extend a shell-based docs check script snippet in the plan implementation notes by asserting the new wording is absent first:

```bash
rg -n "provider-side public quote surface|local dynamic pricing policy|local negotiation engine configuration" docs/features/p2p-network.md docs/features/economy.md docs/cli/p2p.md || true
```

Expected:

```text
<no matches>
```

- [ ] **Step 2: Clarify the public-vs-policy split in docs**

Update `docs/features/p2p-network.md` with wording like:

```md
`p2p.pricing` is the provider-side public quote surface used by remote peers.
It does not, by itself, imply that dynamic pricing, negotiation, or escrow are enabled.
Those higher-level policies are owned by the economy subsystem.
```

Update `docs/features/economy.md` with wording like:

```md
The economy subsystem is the local policy layer for dynamic pricing, negotiation, risk, and escrow.
It may influence public P2P exchange behavior, but it is not the same thing as the provider-side quote surface exposed through `p2p.pricing`.
```

- [ ] **Step 3: Align CLI help text with the same split**

Modify `internal/cli/p2p/pricing.go`:

```go
Short: "Show provider-side P2P quote configuration",
Long:  "Display the current provider-side P2P quote configuration including default per-query price and tool-specific public quote overrides.",
```

Modify `internal/cli/economy/pricing.go`:

```go
Short: "Show dynamic pricing policy configuration",
```

Modify `internal/cli/economy/negotiate.go`:

```go
Short: "Show negotiation engine configuration",
```

- [ ] **Step 4: Update the audit ledger to record the merged control-plane decision**

Adjust `docs/architecture/external-collaboration-audit.md` so the pricing row's assessment explicitly says:

```md
- `p2p.pricing` is the public quote surface.
- `economy.pricing` is the local policy engine.
- `economy.negotiation` is the negotiation engine.
- `paygate` and `settlement` are the runtime payment path.
```

- [ ] **Step 5: Verify the wording and docs build**

Run:

```bash
rg -n "provider-side public quote surface|local dynamic pricing policy|negotiation engine configuration" docs/features/p2p-network.md docs/features/economy.md internal/cli/p2p/pricing.go internal/cli/economy/pricing.go internal/cli/economy/negotiate.go
python3 -m mkdocs build --strict
```

Expected:

```text
<matches for all new wording>
INFO    -  Documentation built in ...
```

- [ ] **Step 6: Commit the control-plane consolidation**

Run:

```bash
git add docs/features/p2p-network.md docs/features/economy.md internal/cli/p2p/pricing.go internal/cli/economy/pricing.go internal/cli/economy/negotiate.go docs/architecture/external-collaboration-audit.md
git -c commit.gpgsign=false commit -m "docs: clarify external pricing control planes"
```

## Task 4: Truth-Align Team And Workspace Operator Surfaces

**Files:**
- Modify: `docs/cli/p2p.md`
- Modify: `docs/features/p2p-network.md`
- Modify: `internal/cli/p2p/team.go`
- Modify: `internal/cli/p2p/workspace.go`
- Modify: `internal/cli/p2p/git.go`
- Modify: `docs/architecture/external-collaboration-audit.md`

- [ ] **Step 1: Capture the current misleading examples before editing**

Run:

```bash
rg -n "Workspace created:|Joined workspace|Left workspace|Team .* disbanded|ID +STATUS +MEMBERS|Use the server API" docs/cli/p2p.md internal/cli/p2p/team.go internal/cli/p2p/workspace.go internal/cli/p2p/git.go
```

Expected:

```text
<matches in docs examples and CLI placeholder text>
```

- [ ] **Step 2: Rewrite docs examples to match the current server-backed/tool-backed reality**

Replace direct-success examples with honest examples such as:

```md
$ lango p2p workspace create "research-project" --goal "Collaborative research on RAG optimization"
Workspace creation requires a running server.
Start the server with 'lango serve' and use the agent tools.
```

and:

```md
$ lango p2p git push a1b2c3d4-5678-9012-abcd-ef1234567890
Push requires a running server. Use 'lango serve' and the p2p_git_push tool.
```

Also replace team examples with the current runtime-backed guidance where appropriate.

- [ ] **Step 3: Tighten CLI long descriptions so they match the docs**

Modify `internal/cli/p2p/team.go`, `internal/cli/p2p/workspace.go`, and `internal/cli/p2p/git.go` so the `Long` text consistently describes these commands as:

```go
Long: `... runtime/server-backed control surface.
Use the running server or the corresponding p2p_* agent tools for live operations.`
```

- [ ] **Step 4: Narrow the chronicler claim until the triple-adder wiring is complete**

Update `docs/features/p2p-network.md` from:

```md
When `chroniclerEnabled` is true, workspace messages are persisted as graph triples...
```

to:

```md
When `chroniclerEnabled` is true, workspace messages are prepared for graph-triple persistence.
Full triple-store persistence depends on the graph-store triple-adder wiring being available in the running application.
```

- [ ] **Step 5: Update the audit ledger to reflect the truthful operator surface**

Modify the workspace and team assessments in `docs/architecture/external-collaboration-audit.md` so they explicitly note:

```md
- the runtime subsystems are real,
- the current CLI is mostly truth-aligned guidance rather than full live control,
- live provenance exchange is the exception.
```

- [ ] **Step 6: Verify docs and targeted CLI tests**

Run:

```bash
go test ./internal/cli/p2p/... -count=1
python3 -m mkdocs build --strict
```

Expected:

```text
ok
INFO    -  Documentation built in ...
```

- [ ] **Step 7: Commit the Phase 3/4 truth-alignment slice**

Run:

```bash
git add docs/cli/p2p.md docs/features/p2p-network.md internal/cli/p2p/team.go internal/cli/p2p/workspace.go internal/cli/p2p/git.go docs/architecture/external-collaboration-audit.md
git -c commit.gpgsign=false commit -m "docs: align team and workspace operator surfaces"
```

## Task 5: Final Verification And Audit Closeout

**Files:**
- Modify: `docs/architecture/external-collaboration-audit.md`
- Test: repo verification

- [ ] **Step 1: Update the audit ledger summary to reflect implemented decisions**

Append a short closeout block under the relevant detailed audits:

```md
### Post-Implementation Notes

- Identity surface aligned with DID-aware CLI/API output.
- Post-pay threshold canonicalized to 0.8 across Phase 1/2 trust routing.
- Pricing and negotiation operator story clarified across `p2p` and `economy`.
- Team/workspace surfaces truth-aligned to the current server-backed runtime model.
```

- [ ] **Step 2: Run the full repository verification required by this repo**

Run:

```bash
go build ./...
go test ./...
python3 -m mkdocs build --strict
```

Expected:

```text
<successful go build>
ok
INFO    -  Documentation built in ...
```

- [ ] **Step 3: Commit the closeout**

Run:

```bash
git add docs/architecture/external-collaboration-audit.md
git -c commit.gpgsign=false commit -m "docs: close external collaboration stabilization slice"
```

## Follow-On Plan

Do **not** fold the following into this implementation slice. Create a separate plan after this one lands:

- live `GET/DELETE /api/p2p/teams/*` operator surfaces
- live workspace create/list/status/join/leave HTTP/operator surfaces
- real trust-weighted team conflict resolution using explicit trust score rather than duration proxy
- deeper unification of provenance bundles, workspace artifacts, and git bundles into one canonical Phase 4 collaboration model

## Self-Review

### Spec Coverage

This plan covers the completed audit findings by priority:

- `P2P identity / trust / reputation` → Task 1 and Task 2
- `pricing / negotiation / settlement` → Task 2 and Task 3
- `team formation / role coordination` → Task 4
- `workspace / shared artifacts` → Task 4

It intentionally stops short of implementing the larger Phase 3/4 live-control expansion and records that as a follow-on plan.

### Placeholder Scan

This plan contains none of the banned placeholder patterns inside executable tasks. Every task includes concrete file paths, code/doc snippets, commands, and commit messages.

### Type Consistency

The plan consistently uses:

- `External Collaboration & Economic Exchange` as the primary capability area
- `P2P Knowledge Exchange Track` for Phase 1/2 work
- `Leader-Led Team Execution Track` for Phase 3/4 operator-surface truth alignment
- `DefaultPostPayThreshold = 0.8` as the canonical post-pay threshold for this implementation slice
