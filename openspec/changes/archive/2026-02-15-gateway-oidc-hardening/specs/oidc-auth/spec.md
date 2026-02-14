## MODIFIED Requirements

### Requirement: Login Flow
The system SHALL provide HTTP endpoints to initiate OIDC login and handle the callback. Auth routes SHALL be rate-limited to a maximum of 10 concurrent requests. State cookies SHALL use per-provider names (`oauth_state_{provider}`) to prevent collision during concurrent multi-provider logins. All cookies SHALL use `isSecure(r)` for the Secure flag to support reverse proxy deployments. The callback response SHALL return structured JSON without exposing user email addresses.

#### Scenario: Login Initiation
- **WHEN** a user accesses `/auth/login/google`
- **THEN** they are redirected to Google's consent page with the correct state and scopes
- **AND** the state cookie SHALL be named `oauth_state_google`
- **AND** the cookie Secure flag SHALL be set via `isSecure(r)`

#### Scenario: Callback Handling
- **WHEN** the user returns to `/auth/callback/google` with a valid code
- **THEN** the system exchanges the code for an ID token
- **AND** issues a `lango_session` cookie to the user
- **AND** deletes the `oauth_state_google` cookie
- **AND** responds with JSON `{"status":"authenticated","sessionKey":"..."}` without email

#### Scenario: State cookie per provider
- **WHEN** a user initiates login with provider "google"
- **AND** simultaneously initiates login with provider "github"
- **THEN** the state cookies SHALL be `oauth_state_google` and `oauth_state_github` respectively
- **AND** neither SHALL overwrite the other

## ADDED Requirements

### Requirement: Logout Endpoint
The system SHALL provide a `POST /auth/logout` endpoint that invalidates the user's session and clears the session cookie.

#### Scenario: Successful logout
- **WHEN** a user sends POST to `/auth/logout` with a valid `lango_session` cookie
- **THEN** the session SHALL be deleted from the session store
- **AND** the `lango_session` cookie SHALL be cleared (MaxAge -1)
- **AND** the response SHALL be JSON `{"status":"logged_out"}`

#### Scenario: Logout without session
- **WHEN** a user sends POST to `/auth/logout` without a session cookie
- **THEN** the response SHALL still clear the cookie and return `{"status":"logged_out"}`

### Requirement: Auth Rate Limiting
The auth endpoints (`/auth/login/*`, `/auth/callback/*`, `/auth/logout`) SHALL be rate-limited to a maximum of 10 concurrent requests using chi middleware.

#### Scenario: Rate limiting applied
- **WHEN** more than 10 concurrent requests are made to auth endpoints
- **THEN** excess requests SHALL be throttled (HTTP 503)
