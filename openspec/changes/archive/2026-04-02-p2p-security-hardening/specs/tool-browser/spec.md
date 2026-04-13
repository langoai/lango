## ADDED Requirements

### Requirement: Private network URL blocking for P2P
The `browser_navigate` handler MUST validate URLs against a private network blocklist when the context carries a P2P origin marker. Blocked addresses: `localhost`, `127.0.0.0/8`, `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`, `169.254.0.0/16`, `[::1]`, and `file://` scheme.

#### Scenario: Internal URL blocked in P2P context
- **WHEN** a P2P peer navigates to `http://127.0.0.1:8080/admin`
- **THEN** the handler returns `ErrBlockedURL` without creating a browser session

#### Scenario: Private network IP blocked
- **WHEN** a P2P peer navigates to `http://10.0.0.1/internal`
- **THEN** the handler returns `ErrBlockedURL`

#### Scenario: File scheme blocked
- **WHEN** a P2P peer navigates to `file:///etc/passwd`
- **THEN** the handler returns `ErrBlockedURL`

#### Scenario: External URL allowed in P2P context
- **WHEN** a P2P peer navigates to `https://example.com`
- **THEN** navigation proceeds normally

#### Scenario: URL validation skipped for local context
- **WHEN** a local (non-P2P) user navigates to `http://localhost:3000`
- **THEN** navigation proceeds normally (no restriction)

### Requirement: Eval action blocking for P2P
The `browser_action` handler MUST reject `eval` actions when the context carries a P2P origin marker, returning `ErrEvalBlockedP2P` before creating a browser session.

#### Scenario: Eval blocked for P2P peer
- **WHEN** a P2P peer sends `browser_action` with `action: "eval"`
- **THEN** the handler returns `ErrEvalBlockedP2P`

#### Scenario: Eval allowed for local user
- **WHEN** a local user sends `browser_action` with `action: "eval"`
- **THEN** the JavaScript is executed normally

### Requirement: Browser sentinel errors
The browser package MUST define `ErrBlockedURL` and `ErrEvalBlockedP2P` sentinel errors.
