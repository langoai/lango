## MODIFIED Requirements

### Requirement: URL validation with DNS resolution
`ValidateURLForP2P` MUST resolve non-IP hostnames via `net.LookupIP` and check all resolved IP addresses against private network ranges. If any resolved IP falls within a private range, the URL SHALL be blocked.

#### Scenario: Hostname resolving to private IP
- **WHEN** a P2P peer requests navigation to `http://metadata.internal` which resolves to `169.254.169.254`
- **THEN** `ValidateURLForP2P` SHALL return `ErrBlockedURL`

#### Scenario: DNS failure allows request
- **WHEN** DNS lookup fails for a hostname
- **THEN** the request SHALL be allowed (the browser will fail on its own)

### Requirement: Post-navigation redirect validation
After `Navigate()` completes in a P2P context, the handler MUST retrieve the final page URL and re-validate it via `ValidateURLForP2P`. If the final URL targets a blocked address, the browser MUST navigate to `about:blank` and return an error.

#### Scenario: Redirect to internal address blocked
- **WHEN** a P2P peer navigates to `https://external.com` which redirects to `http://127.0.0.1:8080`
- **THEN** the handler SHALL navigate to `about:blank`
- **AND** return an error wrapping `ErrBlockedURL`

#### Scenario: No redirect passes without re-validation
- **WHEN** the final URL matches the requested URL
- **THEN** no re-validation SHALL occur
