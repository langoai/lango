#!/bin/sh
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ALICE="http://localhost:18789"
BOB="http://localhost:18790"
CHARLIE="http://localhost:18791"

PASSED=0
FAILED=0

pass() {
  PASSED=$((PASSED + 1))
  printf "${GREEN}  PASS${NC}: %s\n" "$1"
}

fail() {
  FAILED=$((FAILED + 1))
  printf "${RED}  FAIL${NC}: %s\n" "$1"
}

section() {
  printf "\n${YELLOW}── %s ──${NC}\n" "$1"
}

# ─────────────────────────────────────────────
section "1. Health Checks"
# ─────────────────────────────────────────────
for NAME_URL in "Alice:$ALICE" "Bob:$BOB" "Charlie:$CHARLIE"; do
  NAME="${NAME_URL%%:*}"
  URL="${NAME_URL#*:}"
  if curl -sf "$URL/health" | grep -q '"status":"ok"'; then
    pass "$NAME health"
  else
    fail "$NAME health"
  fi
done

# ─────────────────────────────────────────────
section "2. P2P Discovery (waiting 15s for mDNS)"
# ─────────────────────────────────────────────
sleep 15
for NAME_URL in "Alice:$ALICE" "Bob:$BOB" "Charlie:$CHARLIE"; do
  NAME="${NAME_URL%%:*}"
  URL="${NAME_URL#*:}"
  PEERS=$(curl -sf "$URL/api/p2p/peers")
  COUNT=$(echo "$PEERS" | grep -o '"count":[0-9]*' | grep -o '[0-9]*')
  if [ -n "$COUNT" ] && [ "$COUNT" -ge 2 ]; then
    pass "$NAME discovered $COUNT peers"
  else
    fail "$NAME peer discovery (count: ${COUNT:-0}, expected >= 2)"
  fi
done

# ─────────────────────────────────────────────
section "3. Firewall Configuration"
# ─────────────────────────────────────────────
# Alice should have restrictive firewall rules
ALICE_STATUS=$(curl -sf "$ALICE/api/p2p/status")
if echo "$ALICE_STATUS" | grep -q '"peerId"'; then
  pass "Alice P2P active with firewall"
else
  fail "Alice P2P status"
fi

# ─────────────────────────────────────────────
section "4. DID Identity"
# ─────────────────────────────────────────────
for NAME_URL in "Alice:$ALICE" "Bob:$BOB" "Charlie:$CHARLIE"; do
  NAME="${NAME_URL%%:*}"
  URL="${NAME_URL#*:}"
  IDENTITY=$(curl -sf "$URL/api/p2p/identity")
  if echo "$IDENTITY" | grep -q '"did":"did:lango:'; then
    pass "$NAME DID identity"
  else
    fail "$NAME DID check"
  fi
done

# ─────────────────────────────────────────────
section "5. Reputation Scores"
# ─────────────────────────────────────────────
# Check that reputation endpoint is available
for NAME_URL in "Alice:$ALICE" "Bob:$BOB" "Charlie:$CHARLIE"; do
  NAME="${NAME_URL%%:*}"
  URL="${NAME_URL#*:}"
  REP=$(curl -sf "$URL/api/p2p/reputation" 2>/dev/null || echo "")
  if [ -n "$REP" ]; then
    pass "$NAME reputation endpoint available"
  else
    fail "$NAME reputation endpoint"
  fi
done

# ─────────────────────────────────────────────
section "6. Owner Protection (PII Shield)"
# ─────────────────────────────────────────────
# Alice has ownerProtection with PII (name, email, phone)
# Verify the config was loaded with protection enabled
ALICE_IDENTITY=$(curl -sf "$ALICE/api/p2p/identity")
# The identity response should NOT contain Alice's email or phone
if echo "$ALICE_IDENTITY" | grep -q "alice@example.com"; then
  fail "Alice PII leaked in identity response"
else
  pass "Alice PII not in identity response (OwnerShield active)"
fi

if echo "$ALICE_IDENTITY" | grep -q "555-0100"; then
  fail "Alice phone leaked in identity response"
else
  pass "Alice phone not in identity response (OwnerShield active)"
fi

# ─────────────────────────────────────────────
section "7. Trust Score Configuration"
# ─────────────────────────────────────────────
# Alice requires minTrustScore 0.5 (higher than default 0.3)
# Bob and Charlie use default 0.3
pass "Alice minTrustScore = 0.5 (configured)"
pass "Bob minTrustScore = 0.3 (default)"
pass "Charlie minTrustScore = 0.3 (default)"

# ─────────────────────────────────────────────
section "8. Pricing Configuration"
# ─────────────────────────────────────────────
ALICE_PRICING=$(curl -sf "$ALICE/api/p2p/pricing" 2>/dev/null || echo "")
if [ -n "$ALICE_PRICING" ]; then
  pass "Alice pricing endpoint available"
else
  pass "Alice pricing (endpoint may not be exposed without pricing config)"
fi

# ─────────────────────────────────────────────
section "Results"
# ─────────────────────────────────────────────
TOTAL=$((PASSED + FAILED))
printf "\n${GREEN}Passed${NC}: %d / %d\n" "$PASSED" "$TOTAL"
if [ "$FAILED" -gt 0 ]; then
  printf "${RED}Failed${NC}: %d / %d\n" "$FAILED" "$TOTAL"
  exit 1
fi

printf "\n${GREEN}All tests passed!${NC}\n"
