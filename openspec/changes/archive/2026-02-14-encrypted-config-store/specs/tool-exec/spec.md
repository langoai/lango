## MODIFIED Requirements

### Requirement: Environment variable blacklist
The exec tool's default environment variable blacklist SHALL include `LANGO_PASSPHRASE` in addition to existing entries (AWS_SECRET, ANTHROPIC_API_KEY, OPENAI_API_KEY, GOOGLE_API_KEY, SLACK_BOT_TOKEN, DISCORD_TOKEN, TELEGRAM_BOT_TOKEN).

#### Scenario: LANGO_PASSPHRASE filtered
- **WHEN** an agent executes a command and `LANGO_PASSPHRASE` is set in the parent environment
- **THEN** `LANGO_PASSPHRASE` is not passed to the child process

#### Scenario: Safe variables preserved
- **WHEN** an agent executes a command
- **THEN** non-sensitive variables (PATH, HOME, etc.) are passed through
