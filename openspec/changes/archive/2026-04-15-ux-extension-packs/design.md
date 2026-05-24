## Context

Phase 4 closes the zero-config UX roadmap with **extensibility without a trust crater**. Three rails shape the design:

1. **Existing assets do the work where possible.** `internal/skill/` already stores skills as `<dir>/<name>/SKILL.md`, discovers them at startup, imports from git, and respects `cfg.Skill.AllowImport`. `config.SessionMode` already narrows tools/skills per turn and participates in the merge pipeline (`ResolveModes()`). `PromptsDir` already composes additively. None of this needs to change — the pack installer just feeds these pipes.
2. **Install is a trust transaction, not a sync command.** The user must see what's in a pack before it lands on disk. Inspect is always-on and free of side effects; install is inspect + explicit confirm. The flow is identical whether the source is local or remote.
3. **Phase 4 intentionally excludes the execution-bearing surfaces.** Tools, MCP servers, providers, and arbitrary code contributions each have their own trust review and land in later phases. A pack that can only ship skills, modes, and prompts has a bounded blast radius — worst case is prompt injection, which is already a risk surface the user accepts for their own config.

## Goals / Non-Goals

**Goals:**
- A clear, typed manifest (`extension.yaml`) that describes a pack's identity and contents without requiring the user to edit their config to consume it.
- An inspect-before-install trust model that works identically for local dirs and git repos.
- Atomic install/remove: pack removal cleans up everything the pack wrote, no orphans.
- Zero required configuration. A pack installed into `~/.lango/extensions/` and the next `lango` invocation picks up its skills, modes, and prompts automatically.
- Reuse existing skill discovery, mode resolution, and prompt composition — no parallel pipelines.

**Non-Goals:**
- Executing pack-shipped code. No Go plugins, no script loaders, no prompt-triggered install scripts.
- Tool, MCP server, or provider contributions — future phases.
- A central registry, star rankings, or discovery UI — users install from URLs or paths they chose.
- Cross-pack dependency resolution, semver, or upgrade planning — `update` is deferred.
- Implicit auto-update on startup — a pack never mutates itself after install.

## Decisions

### 1. Manifest shape is small, explicit, versioned

**Decision:** `extension.yaml` has a fixed, typed schema:

```yaml
schema: lango.extension/v1
name: python-dev           # kebab-case, unique
version: 0.1.0             # semver
description: ...
author: ...
license: Apache-2.0        # SPDX identifier
homepage: ...              # optional URL
contents:
  skills:
    - name: pytest-refactor
      path: skills/pytest-refactor/SKILL.md
  modes:
    - name: python-review
      systemHint: ...
      tools: [@python]
      skills: [pytest-refactor]
  prompts:
    - path: prompts/PYTHON_CONVENTIONS.md
      section: python       # optional; groups into a named prompt section
```

**Rationale:** A single schema version lets future packs declare `lango.extension/v2` without breaking existing v1 parsers. The `contents` block is a closed enum — a v1 parser that encounters `contents.tools` MUST reject the manifest rather than silently ignore. This prevents a future Phase 5 pack from installing tools via a v1 parser that skipped the field.

**Alternatives considered:**
- *Open-ended `contents` with arbitrary keys:* rejected — silent-ignore is exactly the attack path we want to close.
- *TOML or JSON:* rejected — YAML matches the existing `SKILL.md` frontmatter and `modes` style; consistency wins.

### 2. Trust gate is identical for local, git, and future sources

**Decision:** The installer treats every source as "produce a read-only working copy in a temp dir, hash everything, print inspect report, confirm, atomically move into place." Local-directory sources skip the fetch step but follow the same hash-and-inspect path.

**Rationale:** One code path means the trust guarantee is identical regardless of source, and the inspect output is byte-identical between `lango extension inspect` and the pre-confirm inspect inside `lango extension install`. No `if source == "local"` branches in the trust code.

### 3. Skills are copied into `skills.skillsDir`, not linked

**Decision:** On install, pack skills copy into `<skillsDir>/ext-<pack-name>/<skill-name>/SKILL.md` and its resource files. On remove, the `ext-<pack-name>/` subtree is deleted.

**Rationale:**
- Existing skill discovery walks `skillsDir` — it finds pack-owned skills without any new logic.
- `ext-` prefix is reserved and the installer rejects any user-authored skill under that prefix (`<skillsDir>/ext-…`) to avoid collisions with pack-owned space.
- Copy (not symlink) means `lango extension remove` is a single `os.RemoveAll` of the pack dir + the `ext-<pack-name>/` subdir, with no dangling pointers.
- Name collision with a non-ext skill of the same bare name (e.g., pack `python-dev` ships `pytest-refactor` and user already has `pytest-refactor`) is not a collision because the pack's on-disk path is `ext-python-dev/pytest-refactor/`; the `skill.Registry` key includes the on-disk relative path or uses the bare name — we use a dedup-at-registration rule (see §4).

**Alternatives considered:**
- *Symlinks:* removal would leave dangling links if the pack dir disappears out-of-band; also complicates cross-platform support.
- *Registry manifest file tracking installed files:* introduces a second source of truth and a repair command; the `ext-<pack-name>/` convention encodes the invariant in the filesystem itself.

### 4. Skill-name de-duplication favors user config, then extensions, then built-ins

**Decision:** When `skill.Registry` encounters two skills with the same bare name, priority order is: user-authored (non-`ext-*` path) > extension-authored (`ext-*` path) > built-in. A duplicate across two *extensions* is an install-time rejection — the installer reads the existing registry and refuses to install a pack that would shadow another pack's skill.

**Rationale:** Users expect their own edits to win. Between extensions, there's no principled tiebreaker, so force the user to remove one before installing the other — this prevents a "silently-overridden-pack" failure mode.

### 5. Modes flow through `config.ResolveModes()` with a third origin tier

**Decision:** `config.ResolveModes()` already merges built-in and user-configured modes. Phase 4 adds an `ExtensionModes` input argument that slots between built-in and user — user still wins on name collision. The caller is the app wiring layer, which reads installed pack manifests at startup and feeds the collected modes in.

**Rationale:** This keeps mode resolution a pure function, testable without touching the filesystem. The extension registry is the only filesystem-aware layer that knows where the modes came from.

### 6. Prompts are read from the pack dir at runtime, not copied

**Decision:** Prompts declared in `extension.yaml` under `contents.prompts` are read from `~/.lango/extensions/<name>/<path>` at app startup and composed into the effective system prompt via the existing prompt builder. They are NOT copied into `PromptsDir`.

**Rationale:** Prompts are the smallest, highest-churn content and don't need to live inside `PromptsDir`. Keeping them in the pack dir means `lango extension remove` atomically removes them with the pack — no sweep needed.

### 7. Inspect is free of side effects, including on remote sources

**Decision:** `lango extension inspect <git-url>` clones the repo into a system temp dir, reads and hashes contents, prints the report, and removes the temp dir on exit. `lango extension install <git-url>` re-clones to a staging dir under `~/.lango/extensions/.staging/` and, after user confirm, atomically renames into `~/.lango/extensions/<name>/`.

**Rationale:** Inspect leaking disk state would be a surprise. Install uses a fresh clone to ensure the inspect output and the installed bytes are identical (a malicious server could serve different bytes to inspect vs install, so this attack is not fully closed without hashes-in-URL or signatures — but the manifest hash is recorded in `.installed` and re-verified on every startup, so tampering after install is detected).

### 8. `--yes` does not suppress inspect output

**Decision:** Both `install --yes` and scripted invocations still print the inspect report to stdout. `--yes` only skips the "Confirm install? [y/N]" prompt.

**Rationale:** The trust model is "the user saw what was installed." Hiding that output would break the property even in CI logs where humans audit installs post-hoc. Pre-approval is explicit opt-out from interaction, not from transparency.

### 9. Pack tampering detection via recorded manifest SHA-256

**Decision:** At install time, the installer computes the SHA-256 of `extension.yaml` and of each bundled content file, writes the set to `<pack-dir>/.installed` along with install time and source URL. At startup, the registry recomputes and compares; any mismatch logs a warning `extension.tamper.detected` with the pack name. The pack is loaded anyway (detection, not enforcement) unless `extensions.enforceIntegrity` is `true`.

**Rationale:** Enforced integrity at install is important; silent enforcement at every startup can lock users out after a legitimate manual edit. Phase 4 defaults to detect-only and logs loudly. The `enforceIntegrity` flag is the opt-in for regulated environments.

### 10. CLI output shape follows the lango convention

**Decision:** `lango extension <cmd>` supports `--output table | json | plain`, defaulting to `table` for interactive terminals and `plain` when stdout is not a TTY. Exit code 0 on success, 1 on user-facing error (not found, manifest invalid), 2 on internal error.

**Rationale:** Consistency with other lango CLI commands (`cli-reference` capability). Scripts get stable JSON; humans get readable tables; terminal detection avoids garbled output in pipes.

## Risks / Trade-offs

- **[Risk]** Malicious pack smuggles prompt injection via `systemHint` or a prompt file → **Mitigation:** inspect prints the full SystemHint and prompt bodies before confirm; the same trust surface the user accepts for their own `promptsDir`. Future mitigation: signed-author allowlist.
- **[Risk]** Name collision silently shadows a user skill → **Mitigation:** `ext-<pack-name>/` path prefix plus §4 resolution order; cross-extension collisions are install-time errors.
- **[Risk]** Path-traversal (`../../etc/passwd`) or symlink escape in manifest paths → **Mitigation:** manifest validator rejects any content path containing `..`, any absolute path, and any path whose resolved target escapes the pack root. Enforced in `extension.Manifest.Validate()` and re-enforced at copy time as a defense-in-depth belt-and-suspenders check.
- **[Risk]** Startup merges a pack whose author's git repo was compromised between install and next launch → **Mitigation:** `.installed` SHA-256 manifest is recomputed at startup; mismatch logs a warning. Enforced-integrity mode blocks load.
- **[Risk]** Pack removal mid-way leaves `ext-<pack-name>/` under `skills/` but `~/.lango/extensions/<name>/` gone → **Mitigation:** removal sequence is `(1) delete .installed, (2) delete ext-<name> skill subdir, (3) delete pack dir`. Startup orphan-sweep detects `ext-<name>` subdirs whose parent pack is missing, logs `extension.orphan.detected`, does not auto-delete in Phase 4.
- **[Risk]** Git-sourced install fetches a different revision on the second clone → **Mitigation:** git sources may include a `#<commit-sha>` URL suffix; when present, inspect and install both pin to that SHA. Without a SHA, Phase 4 records the resolved commit hash in `.installed` for post-facto verification but does not re-fetch.
- **[Risk]** Disk-full during install leaves a partial staging dir → **Mitigation:** staging under `~/.lango/extensions/.staging/<name>.<pid>/`; on any error, the directory is removed. Failed inspect has no cleanup obligation — the temp dir is removed unconditionally on function exit.
- **[Trade-off]** Copy-not-link doubles disk use for packs that ship skills. Acceptable — skills are markdown, and the atomic-remove property is worth more than the disk savings.
- **[Trade-off]** No update command means users must `remove` + `install` to upgrade a pack. Acceptable for Phase 4; an `update` that diffs and confirms is a natural Phase 5 follow-up.
- **[Trade-off]** Phase 4 packs cannot ship tools/MCP/providers, so some use cases (a k8s pack wanting an MCP k8s server) need a separate follow-up phase. Acceptable — those surfaces each warrant their own trust review.

## Migration Plan

All changes are additive. No existing config, storage, or user workflow is touched.

1. **Default on, empty.** `extensions.enabled=true`, `extensions.dir=~/.lango/extensions` is a no-op at install time because the directory does not exist. First `lango extension install` creates it.
2. **Rollback.** `extensions.enabled=false` fully disables merge at startup; the on-disk packs are preserved but ignored. `lango extension list` still works as an informational command. To fully revert: `rm -rf ~/.lango/extensions/` and the `ext-*` subdirectories under `~/.lango/skills/`.
3. **Upgrade path from earlier lango.** No data migration. First-run experience is identical to current behavior until a user installs a pack.

## Open Questions

- *Should `lango extension install` support `--into <dir>` to override `extensions.dir` for one install?* Deferred — the current flag set is small and consistent; overriding for a single install complicates the trust audit trail.
- *Should we surface installed packs in the TUI (new "Extensions" page)?* Deferred. The CLI commands are the canonical surface for Phase 4; a TUI page is a natural Phase 5 follow-up once signed packs and an update command exist.
- *Should pack prompts contribute to the system prompt on a per-mode basis?* Current design: all enabled-pack prompts always compose. A future extension of `SessionMode` could include a `prompts:` list referencing pack prompt sections. Not required for MVP.
- *Registry URLs (`lango extension install package://python-dev`) — who runs the registry?* Deferred indefinitely; community-curated markdown doc of trusted URLs is the Phase 4 discovery story.
