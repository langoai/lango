## ADDED Requirements

### Requirement: Core CLI reference delegates agent diagnostics to the dedicated agent page
The core CLI reference SHALL not keep a duplicated embedded agent-diagnostics manual once a dedicated agent CLI reference exists.

#### Scenario: Agent diagnostics duplication stays removed from core docs
- **WHEN** a maintainer updates `docs/cli/core.md`
- **THEN** it SHALL hand off `lango agent trace ...` and `lango agent graph ...` coverage to `docs/cli/agent.md` instead of embedding standalone diagnostics sections
