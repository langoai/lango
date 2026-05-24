# Extension Pack Core (Delta: fix-extension-installer-security)

## Requirement Change: File-copy containment re-validation

The `copyTree` function SHALL accept a `rootDir` parameter and call `ResolvePath(rootDir, rel)` for every file discovered during `filepath.Walk`, matching the per-file validation pattern used by `fetchFromDir` during Inspect. Files that fail containment SHALL be rejected with an error. Additionally, `copyFile` SHALL reject symlinks via `os.Lstat` before opening the source file.

### Scenario: Symlink replaced between Inspect and Install
- **WHEN** a symlink inside a skill directory is replaced to point outside the pack root after Inspect
- **THEN** `copyTree` detects the escape via per-file `ResolvePath` and returns an error
- **AND** the Install operation fails without copying the escaped file

### Scenario: copyFile rejects symlink source
- **WHEN** `copyFile` is called with a source path that is a symlink
- **THEN** it returns an error before opening the file

## Requirement Change: Commit-pinned extension sources

The `GitSource.Fetch` method SHALL support commit SHA references in the `repo.git#<sha>` format. When the ref looks like a hex SHA (7-40 characters), the system SHALL clone without `--branch` and without `--depth=1`, then checkout the specified commit. Branch and tag refs SHALL continue using the existing `--branch` + `--depth=1` strategy.

### Scenario: Fetch with commit SHA
- **WHEN** user installs an extension with source `https://github.com/org/repo.git#abc1234def`
- **THEN** the system clones the repo and checks out commit `abc1234def`
- **AND** the working copy's SourceRef records the pinned SHA

### Scenario: Fetch with branch ref unchanged
- **WHEN** user installs an extension with source `https://github.com/org/repo.git#main`
- **THEN** the system uses `git clone --depth=1 --branch main` as before

## Requirement Change: Inspect preview completeness

The `plannedWrites` function SHALL enumerate all files in skill directories, not just the manifest-listed path. When the skill path points to a file inside a directory (e.g., `skills/x/SKILL.md`), the function SHALL walk the parent directory and include all discovered files in the preview output.

### Scenario: Skill directory with sibling resources
- **WHEN** a pack contains `skills/x/SKILL.md` and `skills/x/references/guide.md`
- **THEN** `extension inspect` reports both files in the planned writes list
