## MODIFIED Requirements

### Requirement: ExpandEnvVars is exported
The `config` package SHALL export `ExpandEnvVars(s string) string` as a public function that replaces `${VAR}` patterns with environment variable values. Variables not set in the environment SHALL be left as-is.

#### Scenario: Env var expansion from external package
- **WHEN** `config.ExpandEnvVars("${OPENAI_API_KEY}")` is called and `OPENAI_API_KEY` is set
- **THEN** the function SHALL return the environment variable value

#### Scenario: Unset env var preserved
- **WHEN** `config.ExpandEnvVars("${UNSET_VAR}")` is called and `UNSET_VAR` is not set
- **THEN** the function SHALL return `"${UNSET_VAR}"` unchanged
