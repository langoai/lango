# Spec: CLI Provider Management

## Overview

This capability defines how the CLI tools interact with the provider system, including listing available providers, configuring them, and verifying their health.

## Purpose

Capability spec for cli-provider-management. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: List supported providers
The system SHALL provide a way to list all supported AI providers in the CLI context.

#### Scenario: Listing providers
- **WHEN** a user or internal component requests a list of supported providers
- **THEN** it returns a list containing at least "openai", "anthropic", and "gemini"
- **AND** the list includes metadata such as the required environment variable for the API key

### Requirement: Configure provider
The system SHALL provide a mechanism to generate a valid `providers` configuration entry for a selected provider.

#### Scenario: Configuring OpenAI
- **WHEN** "openai" is selected for configuration
- **THEN** the system generates a provider config block with `type: "openai"`
- **AND** sets the `apiKey` field to `${OPENAI_API_KEY}` by default

### Requirement: Verify provider health
The system SHALL be able to verify the configuration and connectivity of any provider defined in the configuration.

#### Scenario: Healthy provider
- **WHEN** a configured provider has a valid API key set in the correct environment variable
- **THEN** the verification returns `StatusPass`

#### Scenario: Missing credentials
- **WHEN** a configured provider is missing its API key (env var not set)
- **THEN** the verification returns `StatusFail` with a message indicating the missing variable
