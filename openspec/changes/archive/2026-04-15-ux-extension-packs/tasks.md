## 1. Config surface

- [x] 1.1 Add `ExtensionsConfig` struct to `internal/config/types.go` with `Enabled *bool`, `Dir string`, `EnforceIntegrity bool` and tags
- [x] 1.2 Add `Extensions ExtensionsConfig` field to `Config` in `internal/config/types.go`
- [x] 1.3 Register defaults (`Enabled=boolPtr(true)`, `Dir="~/.lango/extensions"`, `EnforceIntegrity=false`) in `internal/config/loader.go`
- [x] 1.4 Implement `(ExtensionsConfig) ResolveExtensions()` in a new `internal/config/extensions.go` with non-mutating copy-in/copy-out behavior
- [x] 1.5 Unit-test defaults, tilde expansion at consumption site, and ResolveExtensions non-mutation

## 2. Manifest parsing and validation

- [x] 2.1 Create `internal/extension/manifest.go` with `Manifest`, `Contents`, `SkillRef`, `ModeRef`, `PromptRef` structs (YAML tagged)
- [x] 2.2 Implement `ParseManifest(r io.Reader) (*Manifest, error)` that enforces `schema: lango.extension/v1` and rejects unknown `contents` keys
- [x] 2.3 Implement `Manifest.Validate()` with name kebab-case regex, semver regex, SPDX-like license check (accept any non-empty string for now, log warn on non-SPDX), URL parse for homepage
- [x] 2.4 Implement `ValidateContentPath(pathInPackRoot string) error` rejecting absolute paths, `..` segments, and paths that resolve outside pack root via symlinks
- [x] 2.5 Unit-test: valid v1 manifest, unknown `contents.tools` key rejected, `schema: lango.extension/v2` rejected, invalid name, invalid version, invalid paths (abs, traversal, symlink escape)

## 3. Source loaders + working copy

- [x] 3.1 Define `internal/extension/source.go` with a `Source` interface: `Fetch(ctx context.Context) (*WorkingCopy, error)`
- [x] 3.2 Implement `LocalSource` that reads a directory, parses manifest, and computes SHA-256 of the manifest and each declared file
- [x] 3.3 Implement `GitSource` that clones to a system temp directory using the existing skill-import git path, supports `#<ref>` URL suffix, records resolved HEAD SHA, and runs the same hash pipeline as LocalSource
- [x] 3.4 `WorkingCopy` carries: `Manifest`, `ManifestSHA256 string`, `FileHashes map[string][32]byte`, `RootDir string`, `Cleanup func() error`
- [x] 3.5 Unit-test: LocalSource happy path, LocalSource missing manifest, GitSource with ref pins commit, GitSource without ref records HEAD, temp dir cleanup runs on success and error

## 4. Installer

- [x] 4.1 Create `internal/extension/installer.go` with `Installer` holding `extensionsDir`, `skillsDir`, `registry *Registry`
- [x] 4.2 Implement `Installer.Inspect(ctx, src) (*InspectReport, error)` that produces a side-effect-free report (identity, hashes, planned writes, non-contribution disclaimer)
- [x] 4.3 Implement `Installer.Install(ctx, src, opts InstallOptions)` with pipeline: Fetch → Validate → check cross-pack collisions → stage → copy pack files → copy `ext-<name>/` skills → write `.installed` → atomic rename
- [x] 4.4 Implement rollback: on any error in stage/copy phase, `os.RemoveAll` staging dir + any partial `ext-<name>/` skill subdir
- [x] 4.5 Implement `Installer.Remove(ctx, name)` with ordered: delete `.installed` → delete `ext-<name>/` → delete pack dir; continue past later failures with structured log
- [x] 4.6 Implement `detectCollisions(manifest *Manifest, reg *Registry) error` that scans existing ext-* packs for overlapping skill and mode names
- [x] 4.7 Unit-test: install happy path, --yes still prints inspect, duplicate-name rejected, cross-extension skill collision rejected, rollback on copy failure, remove of unknown pack errors

## 5. Installed registry + startup merge

- [x] 5.1 Create `internal/extension/registry.go` with `Registry` type holding loaded `InstalledPack` records
- [x] 5.2 Implement `LoadRegistry(extensionsDir string, enforceIntegrity bool) (*Registry, error)` that walks `*/extension.yaml`, parses each, recomputes hashes, and flags tampered/orphan packs
- [x] 5.3 Implement `Registry.Modes() []config.SessionMode` returning extension-origin modes for merge
- [x] 5.4 Implement `Registry.PromptSources() []PromptSource` returning pack-rooted file paths for the prompt builder
- [x] 5.5 Extend `config.ResolveModes(userModes, extensionModes, builtins)` with the new `extensionModes` arg; user > extension > built-in precedence
- [x] 5.6 Wire `app.New` to load `Registry`, feed modes into `ResolveModes`, and append prompts to `PromptsDir` composition when `extensions.enabled`
- [x] 5.7 Unit-test: empty dir is no-op, broken manifest skipped with warn, tamper warning logged, enforceIntegrity skips tampered pack, orphan ext-* subdir logged

## 6. Skill system integration

- [x] 6.1 Update `FileSkillStore.ListActive()` to walk into `ext-<pack>/` subdirs and discover their SKILL.md files
- [x] 6.2 Add `SourcePack` field to `SkillEntry` with `omitempty` JSON tag
- [x] 6.3 Populate `SourcePack` from the `ext-<name>/` path segment during load; leave empty for non-ext skills
- [x] 6.4 Implement precedence rule: user-authored skill shadows extension-authored skill with same bare name; log debug `skill.name.shadowed_by_user`
- [x] 6.5 Implement cross-extension collision detection at load: error naming both packs + the skill name; surfaced as fatal by wiring layer
- [x] 6.6 Unit-test: pack-owned skill discovered, user-override shadows extension, cross-ext collision errors

## 7. CLI: `lango extension` command group

- [x] 7.1 Create `cmd/lango/extension_cmd.go` with a root `extension` command + `inspect`, `install`, `list`, `remove` subcommands (cobra)
- [x] 7.2 Implement `inspect` subcommand reading source, calling `Installer.Inspect`, rendering report to stdout
- [x] 7.3 Implement `install` subcommand with `--yes`, interactive confirmation, non-TTY-without-yes → exit 3
- [x] 7.4 Implement `list` subcommand with `--output table|json|plain`, TTY-aware default
- [x] 7.5 Implement `remove` subcommand with `--yes`, pre-delete file list print, confirmation prompt
- [x] 7.6 Add `--output` flag validation with exit code 2 for unknown formats
- [x] 7.7 Add help text: top-level notes inspect+confirm trust model, each subcommand has at least one example
- [x] 7.8 Unit-test CLI wiring (exit codes, `--yes` behavior, JSON shape for `list`); smoke-test table formatter

## 8. Documentation

- [x] 8.1 Add a README.md "Extension Packs" section describing the trust model, the four CLI subcommands, and the v1 manifest shape
- [x] 8.2 Add a short "writing a pack" note inside the Extension Packs section with a minimal example `extension.yaml`
- [x] 8.3 Document `extensions.*` config fields next to the existing context/learning config docs
- [x] 8.4 If `.claude/guides/openspec/workflows.md` mentions CLI subcommand patterns, add `extension` to the list (otherwise skip)

## 9. Build, test, verify

- [x] 9.1 `go build ./...` passes
- [x] 9.2 `go test ./...` passes including new unit tests in `internal/extension/`, `internal/cli/extension/`, `internal/config/`, `internal/skill/`
- [x] 9.3 `openspec validate ux-extension-packs --strict` passes
- [ ] 9.4 Manual smoke test: create a sample pack directory with one skill + one mode + one prompt, run `inspect` (confirm report shape), run `install --yes`, restart lango, confirm the mode appears in `/mode` list and the skill in `list_skills`, run `remove --yes`, confirm cleanup (requires interactive terminal; deferred)
