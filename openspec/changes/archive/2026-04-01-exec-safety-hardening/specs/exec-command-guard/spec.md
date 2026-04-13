## ADDED Requirements

### Requirement: classifyLangoExec returns structured reason codes
The `classifyLangoExec` function SHALL return a `(message string, reason ReasonCode)` tuple that distinguishes between lango CLI invocation (`ReasonLangoCLI`) and skill import redirects (`ReasonSkillImport`). The existing `blockLangoExec` function SHALL delegate to `classifyLangoExec` and return only the message string for backward compatibility.

#### Scenario: Lango subcommand classified as LangoCLI
- **WHEN** `classifyLangoExec` receives `lango cron list`
- **THEN** it SHALL return a non-empty message and `ReasonLangoCLI`

#### Scenario: Catch-all lango classified as LangoCLI
- **WHEN** `classifyLangoExec` receives `lango unknown-cmd`
- **THEN** it SHALL return a non-empty message and `ReasonLangoCLI`

#### Scenario: Git clone skill classified as SkillImport
- **WHEN** `classifyLangoExec` receives `git clone https://github.com/org/repo skill-name`
- **THEN** it SHALL return a non-empty message and `ReasonSkillImport`

#### Scenario: Curl skill classified as SkillImport
- **WHEN** `classifyLangoExec` receives `curl https://example.com/skill`
- **AND** the command contains the word "skill"
- **THEN** it SHALL return a non-empty message and `ReasonSkillImport`

#### Scenario: Non-matching command returns empty
- **WHEN** `classifyLangoExec` receives `go build ./...`
- **THEN** it SHALL return empty message and `ReasonNone`

#### Scenario: blockLangoExec delegates to classifyLangoExec
- **WHEN** `blockLangoExec` receives any command
- **THEN** it SHALL call `classifyLangoExec` and return only the message string
