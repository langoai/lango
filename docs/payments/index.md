---
title: Payments
---

# Payments

Lango includes a blockchain payment system for USDC transactions on Base L2, with support for the X402 auto-pay protocol.

!!! warning "Experimental"

    The payments system is under active development. APIs, configuration keys, and behavior may change between releases. Enable it explicitly via `payment.enabled`.

<div class="grid cards" markdown>

-   :coin: **[USDC Payments](usdc.md)**

    ---

    Send and receive USDC on Base L2 with wallet management, spending limits, and transaction history.

    [:octicons-arrow-right-24: Learn more](usdc.md)

-   :arrows_counterclockwise: **[X402 Protocol](x402.md)**

    ---

    Automatic HTTP 402 payment handling using Coinbase's X402 Go SDK with EIP-3009 off-chain signatures.

    [:octicons-arrow-right-24: Learn more](x402.md)

</div>

## Related

- [Security](../security/index.md) -- Authentication and access control
- [Configuration](../getting-started/configuration.md) -- General configuration reference
