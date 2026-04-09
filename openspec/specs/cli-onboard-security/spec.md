# TUI Security & Providers Spec

## Goal
The `lango onboard` TUI MUST support configuration of advanced Security features (Interceptor, Signer, Passphrase) and management of multiple AI Providers.

## Purpose

Capability spec for cli-onboard-security. See requirements below for scope and behavior contracts.

## Requirements

### Security Configuration
The Security form MUST be expanded to include the following sections:

#### Privacy Interceptor
-   **Enable/Disable**: Toggle switch for the interceptor.
-   **Redact PII**: Toggle switch for automatic PII redaction.
-   **Approval Policy**: Select input (options: "dangerous", "all", "configured", "none"). The value SHALL be read from `cfg.Security.Interceptor.ApprovalPolicy`; if empty, default to "dangerous".
-   **Approval Timeout**: Integer input for timeout seconds.
-   **Notify Channel**: Select input for notification channel.
-   **Sensitive Tools**: Text input (comma-separated tool names).
-   **Exempt Tools**: Text input (comma-separated tool names exempt from approval). The value SHALL be read from `cfg.Security.Interceptor.ExemptTools`.

#### Scenario: ApprovalPolicy select replaces boolean
- **WHEN** user opens the Security form in TUI onboard
- **THEN** the form SHALL display an InputSelect for "Approval Policy" with options ["dangerous", "all", "configured", "none"] instead of the legacy "Approval Req." boolean toggle

#### Scenario: ExemptTools text field present
- **WHEN** user opens the Security form in TUI onboard
- **THEN** the form SHALL display an InputText field for "Exempt Tools" below the "Sensitive Tools" field

#### Scenario: Policy value saved to config
- **WHEN** user selects an approval policy and submits the form
- **THEN** `UpdateConfigFromForm` SHALL set `Security.Interceptor.ApprovalPolicy` to the selected `config.ApprovalPolicy` value

#### Scenario: ExemptTools value saved to config
- **WHEN** user enters comma-separated tool names in Exempt Tools and submits
- **THEN** `UpdateConfigFromForm` SHALL parse and set `Security.Interceptor.ExemptTools` as a trimmed string slice

#### Secure Signer
-   **Provider**: Select input (options: "local", "rpc").
-   **Tethering**:
    -   Host connection configuration if "rpc" is selected.
    -   Key ID specification.

#### Local Secrets
-   **Passphrase**: Masked text input for setting the local encryption passphrase.

### Providers Configuration
A new "Providers" section MUST be added to the main menu.

#### Provider List
-   Display a list of all configured providers by their ID (e.g., "anthropic", "custom-openai").
-   Allow selection of a provider to edit.
-   Allow adding a new provider.

#### Provider Form
-   **ID**: Text input (unique identifier).
-   **Type**: Select input (options: "openai", "anthropic", "gemini", "ollama").
-   **API Key**: Text input (standard environment variable handling applies).
-   **Base URL**: Text input (optional, for OpenAI-compatible providers).
-   **Model**: Text input (default model for this provider).

## Constraints
-   All new configuration MUST be saved to `lango.json` correctly mapped to the `security` block and `providers` map.
-   The "Agent" configuration block remains the *default* active provider/model configuration.
