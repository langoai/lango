# Signed Challenge Protocol Spec

## Overview

Extends the P2P handshake protocol to sign Challenge messages, preventing initiator identity spoofing.

## Protocol

### v1.0 (Legacy)
- Protocol ID: `/lango/handshake/1.0.0`
- Challenge: `{nonce, timestamp, senderDID}`
- No signature, no timestamp validation, no nonce replay protection

### v1.1 (Signed)
- Protocol ID: `/lango/handshake/1.1.0`
- Challenge: `{nonce, timestamp, senderDID, publicKey, signature}`
- Signature: ECDSA over `Keccak256(nonce || bigEndian(timestamp, 8) || utf8(senderDID))`
- Verification: `SigToPub(payload, signature)` → compare `CompressPubkey(recovered)` vs `publicKey`

### Challenge Validation (HandleIncoming)
1. Timestamp validation: reject if > 5 min old or > 30s in future
2. Nonce replay: NonceCache.CheckAndRecord() — reject duplicates
3. Signature verification (if present): ECDSA recovery + public key comparison
4. If signature absent: check `requireSignedChallenge` config → reject or allow legacy

### NonceCache
- Data structure: `map[[32]byte]time.Time` with `sync.Mutex`
- TTL: 2 × handshake timeout (default 60s)
- Periodic cleanup via `time.Ticker` goroutine (interval = TTL/2)
- Start/Stop lifecycle

## Configuration

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `p2p.requireSignedChallenge` | bool | false | Reject unsigned challenges |

## Backward Compatibility

- Both v1.0 and v1.1 stream handlers registered on host
- Initiate() always signs (falls back gracefully if wallet unavailable)
- HandleIncoming() accepts both signed and unsigned (unless requireSignedChallenge=true)
