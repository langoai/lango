## MODIFIED Requirements

### Requirement: Config export command
The system SHALL provide a `lango config export <name>` command that outputs decrypted config as JSON. Passphrase verification is required (handled implicitly by the bootstrap process).

#### Scenario: Export profile
- **WHEN** `lango config export default` is run
- **THEN** the passphrase is verified via bootstrap
- **AND** the decrypted config is printed to stdout as formatted JSON, with a WARNING on stderr

#### Scenario: Config JSON output remains decodable on the command writer
- **WHEN** `lango config export` or `lango config get <path> --format json` renders JSON output
- **THEN** the command writer SHALL receive valid pretty-printed JSON that can be decoded without stripping extra framing text
