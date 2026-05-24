## Why

Phase 4 of the zero-config UX roadmap. Phases 1–3 shipped elasticity within a session (interrupt/redirect, retry, inline compaction), per-turn capability shaping (session modes, discovery, cost), and cross-turn continuity (hygiene compaction, recall, learning suggestions). What's still missing is a clean way for users — or third parties — to **grow** an agent's capability surface without editing config by hand, without running arbitrary code, and without a plugin story that turns into a supply-chain liability.

The infrastructure for the data plane already exists: `internal/skill/` stores file-based skills with YAML frontmatter and git-clone/HTTP import support; `config.SessionMode` narrows the per-turn tool and skill set; prompts compose additively through `PromptsDir`. What's missing is the **packaging**, **inspection**, and **install contract** that turns "a directory of markdown" into "a named, versioned, removable capability pack with a trust gate." Phase 4 adds that.

## What Changes

- **Extension pack manifest** (`extension.yaml`) — a small, typed schema that declares a pack's name, version, author, license, and bundled contents. Phase 4 contents are the three safest surfaces: `skills`, `modes`, and `prompts`. No tool, MCP, provider, or arbitrary-code contribution in this phase — those require a stronger trust review and ship later.
- **Trust model = inspect + confirm**. `lango extension inspect <source>` is mandatory before any install would write to disk. Inspect prints the manifest, the SHA-256 of the manifest and each bundled file, what will be written where, and each skill/mode/prompt's summary. `lango extension install <source>` re-runs inspect and then prompts for explicit confirmation (a `--yes` flag exists for scripted installs but does not suppress the inspect output — opt-in pre-approval, not invisibility).
- **Source types**: local directory, git repository (reuse existing `skill-import` git-clone + HTTP fallback pattern). Tarball and registry sources are out of scope for Phase 4.
- **On-disk layout**: each installed pack lives under `~/.lango/extensions/<name>/` with its manifest, bundled files, and an `.installed` metadata record (install time, source URL, manifest SHA-256). Removal deletes the directory atomically.
- **Integration with existing systems**:
  - Skills from packs are copied into the user's `skills.skillsDir` under a pack-prefixed subdirectory so existing skill discovery works unchanged, and pack removal removes exactly the files the pack wrote.
  - Modes from packs merge into `config.ResolveModes()` alongside built-in and user-defined modes; name collisions resolve in favor of user config, then extensions, then built-ins.
  - Prompts from packs append to the effective system prompt via the existing `PromptsDir` composition (pack prompts are read from the pack dir directly, not copied into `PromptsDir`, to keep removal atomic).
- **CLI commands**: `lango extension inspect`, `install`, `list`, `remove`. `update` is deferred.
- **Config surface**: new `extensions` top-level block with `enabled` (default `true`) and `dir` (default `~/.lango/extensions`). Additive — no migration for existing users.

## Capabilities

### New Capabilities
- `extension-pack-core`: manifest schema, source loaders (local/git), installer with inspect+confirm trust gate, installed-pack registry, startup merge into skills/modes/prompts.
- `extension-pack-cli`: `lango extension inspect | install | list | remove` command surface with consistent `--output` modes (`table` default, `json`, `plain`).

### Modified Capabilities
- `config-types`: add `ExtensionsConfig` struct with `Enabled` and `Dir` fields; wire into the config defaults pipeline.
- `skill-system`: document that skills installed by an extension pack participate in the normal skill-discovery surface; no behavior change.

## Impact

- **Code**: new `internal/extension/` package (manifest parsing, validation, installer, registry, inspect formatter). New `cmd/lango/extension_cmd.go` wiring the CLI. `internal/config/types.go` and `internal/config/loader.go` for the new config block. `internal/app/wiring.go` for startup merge into `config.ResolveModes` and skill discovery.
- **APIs**: no external (HTTP/API) surface changes. Internal: new `extension.Manifest`, `extension.Registry`, `extension.Installer` types. `skill.Registry` gains an optional "source" attribution field so the TUI/CLI can show where a skill came from (additive, `omitempty`).
- **Storage**: new `~/.lango/extensions/` directory. Pack-owned subdirectories inside `skills.skillsDir` with a reserved prefix (`ext-<pack-name>/`) to make pack-owned skills identifiable without a separate registry file. No database schema changes.
- **Config**: additive `extensions.enabled` and `extensions.dir`. Built-in defaults keep the subsystem enabled but inert until a pack is installed.
- **Dependencies**: reuse existing `go-git` (already present for skill import) and YAML parser. No new third-party modules.
