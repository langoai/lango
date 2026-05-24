## ADDED Requirements

### Requirement: Extension pack manifest schema v1
The system SHALL parse `extension.yaml` manifests conforming to the `lango.extension/v1` schema. The schema SHALL be a closed set: a v1 parser encountering an unknown top-level field under `contents` (e.g., `tools`, `mcp`, `providers`) SHALL reject the manifest with a validation error referencing the offending field. Allowed `contents` keys in v1 are `skills`, `modes`, and `prompts`.

#### Scenario: Valid v1 manifest parses
- **WHEN** an `extension.yaml` with `schema: lango.extension/v1`, required identity fields (`name`, `version`, `description`), and a well-formed `contents` block is parsed
- **THEN** parsing SHALL succeed and the result SHALL expose `Name`, `Version`, `Description`, `Author`, `License`, `Homepage`, `Contents.Skills`, `Contents.Modes`, and `Contents.Prompts`

#### Scenario: Unknown contents key rejected
- **WHEN** a v1 manifest includes `contents.tools` or any other key outside `{skills, modes, prompts}`
- **THEN** validation SHALL fail with an error naming the unexpected key
- **AND** the installer SHALL NOT attempt to apply any part of the manifest

#### Scenario: Future schema version rejected by v1 parser
- **WHEN** a manifest declares `schema: lango.extension/v2`
- **THEN** the v1 parser SHALL reject it with an explicit version-mismatch error
- **AND** the error message SHALL instruct the user to upgrade lango

### Requirement: Manifest identity and versioning
The manifest SHALL require `name` (kebab-case, 2–64 chars), `version` (semver), and `description`. `author`, `license` (SPDX identifier), and `homepage` (URL) are optional. `name` uniquely identifies a pack within a given lango installation.

#### Scenario: Invalid name rejected
- **WHEN** the manifest declares `name: Python_Dev` (non-kebab)
- **THEN** validation SHALL fail with an error naming the offending field

#### Scenario: Invalid version rejected
- **WHEN** the manifest declares `version: 0.1` (not semver)
- **THEN** validation SHALL fail with an error naming the offending field

#### Scenario: Optional fields absent
- **WHEN** the manifest omits `author`, `license`, and `homepage`
- **THEN** validation SHALL succeed and the omitted fields SHALL be empty strings on the parsed manifest

### Requirement: Manifest path safety
Every path referenced under `contents.skills[].path` and `contents.prompts[].path` SHALL be a relative path within the pack directory. The validator SHALL reject any path that (a) is absolute, (b) contains a `..` segment, or (c) resolves via symlinks to a location outside the pack root. These checks SHALL run at manifest parse time AND again at file-copy time.

#### Scenario: Absolute path rejected
- **WHEN** a manifest declares `path: /etc/passwd` for a skill
- **THEN** validation SHALL fail with a path-safety error

#### Scenario: Parent-traversal rejected
- **WHEN** a manifest declares `path: ../outside/SKILL.md`
- **THEN** validation SHALL fail with a path-safety error

#### Scenario: Symlink escape rejected at copy time
- **WHEN** a path passes manifest validation but, at copy time, the resolved target is outside the pack root
- **THEN** the copy step SHALL fail with a path-safety error
- **AND** the installer SHALL abort without writing any files for that pack

### Requirement: Local-directory pack source loader
The system SHALL support installing a pack from a local directory. The loader SHALL read `extension.yaml` from the directory root, validate the manifest, compute the SHA-256 of the manifest and of every file referenced by the manifest, and return a read-only working copy identical to what would be produced from a remote source.

#### Scenario: Local directory load succeeds
- **WHEN** a directory `/tmp/python-dev/` contains a valid `extension.yaml` with readable referenced files
- **THEN** the loader SHALL return a working copy with computed hashes for the manifest and each file

#### Scenario: Missing manifest fails loudly
- **WHEN** the directory has no `extension.yaml`
- **THEN** the loader SHALL return an error naming the missing file

### Requirement: Git-repository pack source loader
The system SHALL support installing a pack from a git repository URL. The loader SHALL clone into a system temp directory, run the same manifest validation and hashing as the local loader, and return a working copy. When the URL carries a `#<commit-sha>` suffix, the loader SHALL pin the clone to that revision; otherwise it SHALL use the default branch and record the resolved HEAD SHA for inclusion in `.installed`.

#### Scenario: Git URL without ref uses default branch
- **WHEN** the URL is `https://example.com/langoai/pack.git`
- **THEN** the loader SHALL clone the default branch and record the resolved HEAD SHA in the working copy metadata

#### Scenario: Git URL with ref pins to commit
- **WHEN** the URL is `https://example.com/langoai/pack.git#abc1234`
- **THEN** the loader SHALL pin the clone to commit `abc1234`
- **AND** SHALL fail cleanly if the commit does not exist

#### Scenario: Temp directory cleaned on inspect
- **WHEN** `lango extension inspect` completes, succeeds or fails
- **THEN** the temp clone directory SHALL be removed
- **AND** no state SHALL remain under the user's home directory

### Requirement: Inspect report is side-effect free
`lango extension inspect <source>` SHALL produce a human-readable report of the pack's identity, hashes, and contents without writing any files under `extensions.dir`, `skills.skillsDir`, or anywhere else outside the system temp directory used for fetching. The report SHALL list the exact paths that an install would write and SHALL state explicitly that v1 packs cannot install tools, MCP servers, or providers.

#### Scenario: Inspect prints identity and hashes
- **WHEN** inspect is run against a valid pack
- **THEN** the report SHALL include the pack name, version, author, license, homepage (or "<none>"), the SHA-256 of `extension.yaml`, and the SHA-256 of each bundled file

#### Scenario: Inspect prints planned writes
- **WHEN** inspect is run
- **THEN** the report SHALL list every file that would be written to `<extensions.dir>/<pack-name>/` and to `<skills.skillsDir>/ext-<pack-name>/`

#### Scenario: Inspect prints non-contribution disclaimer
- **WHEN** inspect is run
- **THEN** the report SHALL include a disclosure that v1 packs do not install tools, MCP servers, providers, or executable code

### Requirement: Install is inspect + confirm + atomic move
`lango extension install <source>` SHALL (a) run the loader, (b) produce and print the inspect report, (c) prompt for explicit confirmation unless `--yes` is passed (inspect output is still printed with `--yes`), (d) stage the pack under `<extensions.dir>/.staging/<name>.<pid>/`, (e) copy skill files into `<skills.skillsDir>/ext-<name>/`, (f) write `.installed` metadata, and (g) atomically rename the staging directory into `<extensions.dir>/<name>/`. Any failure in steps (d)–(g) SHALL roll back all files written during that install.

#### Scenario: Install with confirmation succeeds
- **WHEN** the user runs `lango extension install ./python-dev` and types `y` at the prompt
- **THEN** the pack directory SHALL appear at `<extensions.dir>/python-dev/` with the manifest and bundled files
- **AND** pack-owned skills SHALL appear under `<skills.skillsDir>/ext-python-dev/`
- **AND** `.installed` SHALL contain the install timestamp, source URL, and SHA-256 of manifest and each file

#### Scenario: Install with --yes still prints inspect
- **WHEN** the user runs `lango extension install --yes ./python-dev`
- **THEN** the inspect report SHALL be printed to stdout before the install proceeds
- **AND** no confirmation prompt SHALL be shown

#### Scenario: Rollback on copy failure
- **WHEN** step (e) fails partway because of an I/O error
- **THEN** all files written during this install SHALL be removed
- **AND** no partial pack SHALL appear under `<extensions.dir>/<name>/`
- **AND** no `ext-<name>/` subdir SHALL remain under `<skills.skillsDir>/`

#### Scenario: Duplicate-name install rejected
- **WHEN** the user runs `install` for a pack whose `name` already exists under `<extensions.dir>/`
- **THEN** the command SHALL fail with a clear "already installed" error
- **AND** SHALL suggest `lango extension remove <name>` to proceed

### Requirement: Cross-pack skill and mode collision rejected at install
Before writing, the installer SHALL check the existing on-disk registry for skill names and mode names declared by the new pack. If any declared name collides with a name already owned by *another extension pack* (not by the user and not by built-ins), the install SHALL fail with an error identifying the conflicting name and pack.

#### Scenario: Cross-extension skill collision blocks install
- **WHEN** pack `python-A` is installed and ships skill `pytest-refactor`
- **AND** pack `python-B` attempts to install a skill also named `pytest-refactor`
- **THEN** the `python-B` install SHALL fail with an error naming both packs

#### Scenario: User-authored skill with same bare name does not block install
- **WHEN** the user has a hand-authored skill at `<skills.skillsDir>/pytest-refactor/`
- **AND** pack `python-A` ships a skill named `pytest-refactor`
- **THEN** the install SHALL succeed
- **AND** skill-name resolution (see skill-system spec) SHALL favor the user's skill at runtime

### Requirement: Removal is atomic and sweeps all pack-owned files
`lango extension remove <name>` SHALL (a) delete `<extensions.dir>/<name>/.installed` first, (b) delete `<skills.skillsDir>/ext-<name>/`, and (c) delete `<extensions.dir>/<name>/`. If (c) fails after (a) and (b) succeeded, the pack is considered removed from the effective config even if filesystem state is partial, and subsequent startups SHALL log `extension.orphan.detected` for any lingering `ext-<name>/` subdir.

#### Scenario: Normal remove
- **WHEN** the user runs `lango extension remove python-dev`
- **THEN** after the command exits, neither `<extensions.dir>/python-dev/` nor `<skills.skillsDir>/ext-python-dev/` SHALL exist

#### Scenario: Remove of unknown pack
- **WHEN** the user runs `lango extension remove missing`
- **THEN** the command SHALL exit with an error naming the pack and SHALL NOT touch other pack state

### Requirement: Startup registry loads installed packs
At app startup, before `config.ResolveModes()` runs, the system SHALL walk `<extensions.dir>/*/extension.yaml`, parse each manifest, and build an in-memory registry of `InstalledPack` records. The registry SHALL surface (a) the list of modes to be merged into `ResolveModes`, (b) the list of prompt files to append to the system prompt, and (c) the list of `ext-<name>/` skill subdirs that must be discoverable by the existing skill walker.

#### Scenario: Empty extensions dir is a no-op
- **WHEN** `<extensions.dir>/` does not exist or is empty
- **THEN** startup SHALL succeed with zero pack records and no warnings

#### Scenario: Invalid manifest is skipped with log
- **WHEN** one pack under `<extensions.dir>/broken/` has an invalid `extension.yaml`
- **THEN** startup SHALL log a warning naming the pack
- **AND** SHALL continue loading the remaining packs
- **AND** SHALL NOT panic

### Requirement: Startup tamper detection
For each loaded pack, the startup registry SHALL recompute the SHA-256 of the manifest and each bundled file and compare against the values recorded in `.installed`. Any mismatch SHALL emit a structured `extension.tamper.detected` warning log including pack name and mismatched file path. When `extensions.enforceIntegrity` is `true` (default `false`), the pack SHALL NOT be loaded on mismatch.

#### Scenario: Tamper detected, default mode loads with warning
- **WHEN** a user edits a pack's SKILL.md after install and starts lango
- **AND** `extensions.enforceIntegrity` is `false`
- **THEN** a warning SHALL be logged naming the pack and file
- **AND** the pack SHALL still participate in the effective config

#### Scenario: Tamper detected, enforcement skips pack
- **WHEN** a mismatch is detected and `extensions.enforceIntegrity` is `true`
- **THEN** the pack SHALL NOT be added to the registry
- **AND** none of its modes, skills, or prompts SHALL be merged into the effective config

### Requirement: Orphan ext-* skill subdirs logged, not deleted
At startup, for every `<skills.skillsDir>/ext-<name>/` subdir, the system SHALL verify that `<extensions.dir>/<name>/` exists. For any orphan, the system SHALL log a `extension.orphan.detected` warning identifying the subdir. The system SHALL NOT auto-delete orphans in Phase 4.

#### Scenario: Orphan logged
- **WHEN** `<skills.skillsDir>/ext-python-dev/` exists but `<extensions.dir>/python-dev/` does not
- **THEN** a warning SHALL be logged naming `ext-python-dev`
- **AND** the orphan files SHALL remain on disk

### Requirement: Gated by extensions.enabled
When `extensions.enabled` is `false`, the startup registry SHALL skip pack discovery entirely and the CLI `install`/`remove` subcommands SHALL refuse to run with a clear error directing the user to enable the subsystem.

#### Scenario: Disabled subsystem is inert
- **WHEN** `extensions.enabled=false` and `<extensions.dir>/python-dev/` exists
- **THEN** the startup registry SHALL be empty
- **AND** the pack's modes, skills, and prompts SHALL NOT be merged

#### Scenario: Install refuses when disabled
- **WHEN** `extensions.enabled=false` and the user runs `lango extension install ./python-dev`
- **THEN** the command SHALL exit with an error instructing the user to set `extensions.enabled=true` first
- **AND** no files SHALL be written
