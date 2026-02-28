## Context

The Lango project has grown significantly with security hardening (P0-P2), settings UI enhancements, P2P features, and automation systems. Four core files exceeded maintainable size thresholds: `tools.go` (2,709 lines), `wiring.go` (1,871 lines), `forms_impl.go` (1,790 lines), and `types.go` (943 lines). These files became difficult to navigate, prone to merge conflicts, and hard to review in PRs.

The existing architecture is healthy — no circular dependencies, callback patterns work correctly, and the layer boundaries (Core → Application → UI) are well-maintained. The issue is purely structural: too many concerns packed into single files.

## Goals / Non-Goals

**Goals:**
- Reduce the largest files to under 600 lines each by splitting into domain-focused files
- Improve code navigability — developers can find code by domain (e.g., P2P tools in `tools_p2p.go`)
- Reduce merge conflicts when multiple developers work on different domains
- Maintain 100% API compatibility — no consumer changes required

**Non-Goals:**
- Refactoring business logic or changing behavior
- Introducing new abstractions or interfaces
- Changing package boundaries or moving code between packages
- Splitting files that are already well-structured (handler.go, app.go, store.go, server.go)

## Decisions

### Decision 1: Same-package file splits only

**Choice**: Move functions/types to new files within the same Go package.

**Rationale**: Go packages are the compilation unit — files within a package share the same namespace. This means splitting files has zero impact on consumers, requires no import changes, and carries minimal risk.

**Alternative considered**: Extracting sub-packages. Rejected because it would introduce new import paths, potentially create circular dependencies, and require API redesign.

### Decision 2: Domain-based file naming convention

**Choice**: Use `<base>_<domain>.go` naming pattern (e.g., `tools_p2p.go`, `wiring_graph.go`, `types_security.go`).

**Rationale**: Consistent naming makes files discoverable. The `<base>` prefix groups related files in directory listings, and the `<domain>` suffix indicates the concern.

**Alternative considered**: Flat naming (e.g., `p2p_tools.go`). Rejected because it breaks the visual grouping in file listings.

### Decision 3: Keep orchestrator functions in the original file

**Choice**: The original file retains orchestration/entry-point functions (e.g., `buildTools` stays in `tools.go`), while domain-specific builders move to new files.

**Rationale**: Developers looking for the entry point naturally check the original file. Domain implementations are then one click away.

### Decision 4: Two-phase incremental approach

**Choice**: Phase 1 splits the three largest files (~6,370 lines), Phase 2 splits the config types (~943 lines). Each phase is independently deployable.

**Rationale**: Smaller, verifiable batches reduce risk and enable incremental review.

## Risks / Trade-offs

- **[Risk] Duplicate declarations during split** → Mitigated by removing moved code from the original file immediately after creating the new file, verified with `go build ./...` after each phase.
- **[Risk] Missing imports in new files** → Each new file must be analyzed for its specific import needs; verified by compilation.
- **[Trade-off] More files to navigate** → Accepted; domain-focused files are easier to find than scrolling through 2,700-line monoliths. IDE "Go to Definition" works regardless of file count.
- **[Trade-off] Some small files (~90-140 lines)** → Accepted; consistent domain grouping is more valuable than minimum file size thresholds.
