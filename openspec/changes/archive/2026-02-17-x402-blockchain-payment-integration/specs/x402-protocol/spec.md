## ADDED Requirements

### Requirement: X402 challenge parsing
The system SHALL parse HTTP 402 responses by extracting payment parameters from headers: `X-Payment-Amount`, `X-Payment-Token`, `X-Payment-Recipient`, `X-Payment-Network`, `X-Payment-ChainId`.

#### Scenario: Parse valid 402 response
- **WHEN** an HTTP 402 response is received with payment headers
- **THEN** a Challenge struct is returned with amount, tokenAddress, recipientAddress, network, and chainID

#### Scenario: Missing required headers
- **WHEN** an HTTP 402 response is missing `X-Payment-Amount` or `X-Payment-Recipient`
- **THEN** an error is returned

### Requirement: X402 payment proof header
The system SHALL build an `X-PAYMENT` header containing a JSON payload with txHash, from address, and chainId for retrying the paid request.

#### Scenario: Build payment header
- **WHEN** `BuildPaymentHeader` is called with a PaymentPayload
- **THEN** a JSON-encoded string suitable for the X-PAYMENT header is returned

### Requirement: X402 interceptor with auto-pay limit
The system SHALL provide an Interceptor that processes X402 challenges by making a payment through PaymentService, subject to a configurable `maxAutoPayAmount` limit.

#### Scenario: Auto-pay within limit
- **WHEN** an X402 challenge amount is below `maxAutoPayAmount`
- **THEN** the payment is made automatically and the PaymentPayload is returned

#### Scenario: Auto-pay exceeds limit
- **WHEN** an X402 challenge amount exceeds `maxAutoPayAmount`
- **THEN** an error is returned and no payment is made

#### Scenario: Auto-intercept disabled
- **WHEN** X402 auto-intercept is disabled in config
- **THEN** challenges are not auto-paid and an error is returned
