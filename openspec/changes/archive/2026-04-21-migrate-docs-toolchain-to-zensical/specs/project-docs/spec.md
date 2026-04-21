## ADDED Requirements

### Requirement: Repository docs references describe the Zensical docs toolchain
The README.md, docs/architecture/project-structure.md, and docs/development/build-test.md SHALL describe Zensical as the canonical docs toolchain and reference `zensical.toml` and `.venv/bin/zensical build` instead of MkDocs as the default docs path.

#### Scenario: README and architecture docs reference Zensical
- **WHEN** a user reads README.md and docs/architecture/project-structure.md
- **THEN** they SHALL see Zensical-native docs tooling references instead of `mkdocs.yml` as the canonical site definition

#### Scenario: Build-test docs reference the Zensical build path
- **WHEN** a user reads docs/development/build-test.md
- **THEN** the docs build instructions SHALL use `.venv/bin/zensical build`
