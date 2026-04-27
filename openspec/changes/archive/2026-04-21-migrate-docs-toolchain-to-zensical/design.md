## Context

The project is finishing a docs-toolchain migration. Zensical is now the canonical path, but the OpenSpec contract still needs to reflect the new site definition, the structural removal of hidden docs from the public tree, and the repository documentation updates that describe the new build flow.

## Goals / Non-Goals

**Goals:**

- Make Zensical the canonical docs toolchain in the spec layer.
- Represent hidden documentation as outside the public `docs/` tree instead of relying on MkDocs-only exclusion behavior.
- Keep the public cockpit operator guidance consolidated on `docs/features/cockpit.md`.
- Update repository-facing docs references so they describe the Zensical-native build path.

**Non-Goals:**

- Rewrite unrelated documentation content.
- Redesign the public site visuals.
- Broaden the public docs surface beyond the compatibility contract already established.

## Decisions

### Decision: Zensical is the canonical docs toolchain

The docs site should be described and maintained as a Zensical-native site, with `zensical.toml` as the source of truth.

Why:
- It matches the migration direction and avoids carrying forward MkDocs-era assumptions.

### Decision: Hidden docs live outside the public docs tree

Hidden support material and withdrawn cockpit sub-guides should be structurally removed from `docs/` instead of being treated as public build inputs.

Why:
- The public documentation tree becomes explicit and easier to reason about.

### Decision: Cockpit stays consolidated on the main public page

`docs/features/cockpit.md` remains the single public cockpit entry after the hidden guide move.

Why:
- It keeps the public operator surface easy to find and avoids fragmented entry points.

### Decision: Repository docs should name the new toolchain

README and docs-architecture references should point at the Zensical toolchain and its build path instead of treating MkDocs as the canonical path.

Why:
- Repository-facing docs need to match the actual documented workflow.

## Risks / Trade-offs

- The public docs tree is smaller, but that is intentional.
- Moving hidden docs out of `docs/` changes long-standing paths, so link integrity must be preserved.
- Toolchain wording in repository docs needs to stay in sync with the site configuration as the migration settles.
