## ADDED Requirements

### Requirement: Authentication Middleware
The `requireAuth` middleware SHALL validate the `lango_session` cookie against the session store for all protected routes. When `auth` is nil (no OIDC configured), the middleware SHALL pass all requests through without validation.

#### Scenario: Nil auth passes through
- **WHEN** the middleware is initialized with nil AuthManager
- **THEN** all requests SHALL pass through to the next handler without checking cookies

#### Scenario: Missing cookie returns 401
- **WHEN** a request arrives without a `lango_session` cookie
- **AND** auth is configured
- **THEN** the middleware SHALL respond with HTTP 401 Unauthorized

#### Scenario: Invalid session returns 401
- **WHEN** a request has a `lango_session` cookie with a value not found in the session store
- **THEN** the middleware SHALL respond with HTTP 401 Unauthorized

#### Scenario: Valid session sets context
- **WHEN** a request has a valid `lango_session` cookie matching a session in the store
- **THEN** the middleware SHALL store the session key in the request context
- **AND** pass the request to the next handler

### Requirement: Session Context Extraction
The `SessionFromContext` function SHALL extract the authenticated session key from a request context. It SHALL return an empty string if no session is present.

#### Scenario: Extract session from context
- **WHEN** `SessionFromContext` is called on a context set by `requireAuth`
- **THEN** it SHALL return the authenticated session key

#### Scenario: Empty context returns empty string
- **WHEN** `SessionFromContext` is called on a context without a session
- **THEN** it SHALL return an empty string

### Requirement: Origin Checker
The `makeOriginChecker` function SHALL build a WebSocket CheckOrigin function based on a list of allowed origins. It SHALL normalize trailing slashes during comparison.

#### Scenario: Empty list returns nil
- **WHEN** the allowed origins list is empty or nil
- **THEN** `makeOriginChecker` SHALL return nil (gorilla default same-origin behavior)

#### Scenario: Wildcard allows all
- **WHEN** the allowed origins list contains `*`
- **THEN** the checker SHALL allow all origins

#### Scenario: Specific origins whitelist
- **WHEN** the allowed origins list contains specific URLs
- **THEN** the checker SHALL allow only those origins
- **AND** reject origins not in the list

#### Scenario: No Origin header allowed
- **WHEN** a request has no Origin header (same-origin request)
- **THEN** the checker SHALL allow the request

### Requirement: Secure Detection
The `isSecure` function SHALL detect HTTPS connections both directly (TLS) and behind reverse proxies (X-Forwarded-Proto header).

#### Scenario: Direct TLS
- **WHEN** `r.TLS` is non-nil
- **THEN** `isSecure` SHALL return true

#### Scenario: Reverse proxy HTTPS
- **WHEN** `X-Forwarded-Proto` header is "https" (case-insensitive)
- **THEN** `isSecure` SHALL return true

#### Scenario: Plain HTTP
- **WHEN** neither TLS nor X-Forwarded-Proto indicates HTTPS
- **THEN** `isSecure` SHALL return false
