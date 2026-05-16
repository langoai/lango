## ADDED Requirements

### Requirement: Core CLI reference delegates config commands to the dedicated config page
The core CLI reference SHALL not keep a duplicated embedded config command manual once a dedicated config CLI reference exists.

#### Scenario: Config command duplication stays removed from core docs
- **WHEN** a maintainer updates `docs/cli/core.md`
- **THEN** it SHALL hand off `lango config ...` coverage to `docs/cli/config.md` instead of embedding standalone config subcommand sections
